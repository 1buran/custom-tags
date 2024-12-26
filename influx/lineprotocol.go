package influx

import (
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"
)

func extractTagKeyVal(s string) (key string, val string) {
	idx := strings.LastIndex(s, ",")
	if idx == -1 {
		return
	}
	if idx-1 > 1 {
		key = s[:idx]
	}
	if idx+1 < len(s) {
		val = s[idx+1:]
	}
	return
}

func escapeMeasurement(s string) string {
	for _, v := range []string{",", " "} {
		s = strings.ReplaceAll(s, v, "\\"+v)
	}
	return s
}

func escapeTagKVFieldK(s string) string {
	for _, v := range []string{",", "=", " "} {
		s = strings.ReplaceAll(s, v, "\\"+v)
	}
	return s
}

func escapeFiledV(s string) string {
	for _, v := range []string{`"`, `\`} {
		s = strings.ReplaceAll(s, v, "\\"+v)
	}
	s = fmt.Sprintf("%q", s)
	return s
}

// Convert struct to influxdb line protocol.
//
// The structs should describe their reflection to influx line protocol format with struct tags:
// tag name represents the name of measurement, tag or field of line protocol row,
// tag value represents the type of data: measurement, tag, field or timestamp.
// The only one measurement and timestamp should be, for these types of data the name of tag
// may be omitted cos it will not be used.
//
// Example:
//
//	type Something struct {
//		Node string `influx:",measurement"` // name is omitted cos will not used
//		DataCenter    string `influx:"dc,tag"`
//		CloudProvider string `influx:"cloud,tag"`
//		Errors int           `influx:"errors,field"`
//		Uptime time.Duration `influx:"uptime,field"`
//		Timestamp time.Time `influx:",timestamp"` // name is omitted cos will not used
//	}
func ConvertToInfluxLineProtocol(v any) string {
	var measurement string
	var timestamp time.Time

	vt := reflect.ValueOf(v)
	at := reflect.TypeOf(v)

	if ms := vt.MethodByName("InfluxMeasurement"); ms.Kind() == reflect.Func {
		r := ms.Call([]reflect.Value{})
		measurement = r[0].Interface().(string)
	}
	if tm := vt.MethodByName("InfluxTimestamp"); tm.Kind() == reflect.Func {
		r := tm.Call([]reflect.Value{})
		timestamp = r[0].Interface().(time.Time)
	}

	var metric = map[string][]string{
		"tag":   {},
		"field": {},
	}

	for i := range at.NumField() {
		if tag, ok := at.Field(i).Tag.Lookup("influx"); ok {
			k, v := extractTagKeyVal(tag)
			metricName, metricType := escapeTagKVFieldK(k), v

			switch metricType {
			case "measurement":
				measurement = fmt.Sprint(vt.FieldByName(at.Field(i).Name))
			case "timestamp":
				timestamp = vt.FieldByName(at.Field(i).Name).Interface().(time.Time)
			default:
				// Check for custom marshaling of current type:
				if marshal, ok := at.Field(i).Type.MethodByName("MarshalInflux"); ok {
					r := marshal.Func.Call([]reflect.Value{
						vt.FieldByName(at.Field(i).Name),
					})

					val := r[0].Interface().(string)
					err, ok := r[1].Interface().(error)
					if ok && err != nil {
						log.Printf(
							"%s %q MarshalInflux error: %s", metricType, metricName, err)
						continue
					}
					metric[metricType] = append(
						metric[metricType], fmt.Sprintf("%s=%v", metricName, val))
					continue
				}

				mval := fmt.Sprintf("%s=", metricName)

				switch at.Field(i).Type.Kind() {
				case reflect.String:
					s := vt.FieldByName(at.Field(i).Name).Interface().(string)
					if metricType == "tag" {
						s = escapeTagKVFieldK(s)
					} else if metricType == "field" {
						s = escapeFiledV(s)
					}
					mval += s
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					mval += fmt.Sprintf("%vi", vt.FieldByName(at.Field(i).Name))
				case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
					mval += fmt.Sprintf("%vu", vt.FieldByName(at.Field(i).Name))
				default:
					mval += fmt.Sprintf("%v", vt.FieldByName(at.Field(i).Name))
				}
				metric[metricType] = append(metric[metricType], mval)
			}
		}
	}

	if measurement == "" {
		return "error: `influx:\",measurement\"` not found"
	}

	if timestamp.IsZero() {
		return "error: `influx:\",timestamp\"` not found"
	}

	if len(metric["field"]) == 0 {
		return "error: points must have at least one field"
	}

	measurement = escapeMeasurement(measurement)

	if len(metric["tag"]) >= 1 {
		return fmt.Sprintf(
			"%s,%s %s %d", measurement, strings.Join(metric["tag"], ","),
			strings.Join(metric["field"], ","), timestamp.UnixNano(),
		)
	}

	return fmt.Sprintf(
		"%s %s %d", measurement, strings.Join(metric["field"], ","), timestamp.UnixNano(),
	)
}
