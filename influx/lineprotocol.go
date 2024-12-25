package influx

import (
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"
)

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
			params := strings.Split(tag, ",")
			metricName, metricType := params[0], params[1]

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
						log.Println(err)
						continue
					}
					metric[metricType] = append(
						metric[metricType], fmt.Sprintf("%s=%v", metricName, val))
					continue
				}

				// Default marshaling, based on primitive types:
				mval := fmt.Sprintf("%s=%v", metricName, vt.FieldByName(at.Field(i).Name))
				switch at.Field(i).Type.Kind() {
				case reflect.Int:
					mval += "i"
				case reflect.Uint64:
					mval += "u"
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

	return fmt.Sprintf("%s,%s %s %d", measurement, strings.Join(metric["tag"], ","),
		strings.Join(metric["field"], ","), timestamp.UnixNano())
}
