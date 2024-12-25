package influx

import (
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestExtractTagKeyVal(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		Sample   string
		Expected []string
	}{
		{Sample: "key1,key2,key2,field", Expected: []string{"key1,key2,key2", "field"}},
		{Sample: ",measurement", Expected: []string{"", "measurement"}},
		{Sample: "key,field", Expected: []string{"key", "field"}},
		{Sample: "key,", Expected: []string{"key", ""}},
		{Sample: "", Expected: []string{"", ""}},
		{Sample: ",", Expected: []string{"", ""}},
	}

	for _, testCase := range testCases {
		key, val := extractTagKeyVal(testCase.Sample)
		if testCase.Expected[0] != key {
			t.Errorf("expected %s, got: %s", testCase.Expected[0], key)
		}
		if testCase.Expected[1] != val {
			t.Errorf("expected %s, got: %s", testCase.Expected[1], val)
		}
	}
}

func TestConvertToInfluxLineProtocol(t *testing.T) {
	t.Parallel()

	t.Run("escape/measurement/space", func(t *testing.T) {
		ts := time.Now()

		v := struct {
			Ts time.Time `influx:",timestamp"`
			Ms string    `influx:",measurement"`

			Value string `influx:"value,field"`
		}{
			Ms: "escape measurement",
			Ts: ts,

			Value: "long string",
		}

		expected := "escape\\ measurement value=\"long string\" " + strconv.FormatInt(ts.UnixNano(), 10)
		row := ConvertToInfluxLineProtocol(v)
		if expected != row {
			t.Errorf("expected: %s, got: %s", expected, row)
		}
	})

	t.Run("escape/measurement/comma", func(t *testing.T) {
		ts := time.Now()

		v := struct {
			Ts time.Time `influx:",timestamp"`
			Ms string    `influx:",measurement"`

			Value string `influx:"value,field"`
		}{
			Ms: "escape,measurement",
			Ts: ts,

			Value: "long string",
		}

		expected := "escape\\,measurement value=\"long string\" " + strconv.FormatInt(ts.UnixNano(), 10)
		row := ConvertToInfluxLineProtocol(v)
		if expected != row {
			t.Errorf("expected: %s, got: %s", expected, row)
		}
	})

	t.Run("escape/tag/key", func(t *testing.T) {
		ts := time.Now()

		v := struct {
			Ts time.Time `influx:",timestamp"`
			Ms string    `influx:",measurement"`

			Tag1 string `influx:"tag1,key1,tag"`
			Tag2 string `influx:"tag2=key2,tag"`
			Tag3 string `influx:"tag3 key3,tag"`

			Value string `influx:"Field,field"`
		}{
			Ms:    "escape measurement",
			Ts:    ts,
			Tag1:  "other val",
			Tag2:  "ex value2",
			Tag3:  "new value3",
			Value: "long string",
		}

		expected := "escape\\ measurement,tag1\\,key1=other\\ val,tag2\\=key2=ex\\ value2,tag3\\ key3=new\\ value3 Field=\"long string\" " + strconv.FormatInt(ts.UnixNano(), 10)
		row := ConvertToInfluxLineProtocol(v)
		if expected != row {
			t.Errorf("expected: %s, got: %s", expected, row)
		}
	})

	t.Run("escape/tag/value", func(t *testing.T) {
		ts := time.Now()

		v := struct {
			Ts time.Time `influx:",timestamp"`
			Ms string    `influx:",measurement"`

			Tag1 string `influx:"tag1,key1,tag"`
			Tag2 string `influx:"tag2=key2,tag"`
			Tag3 string `influx:"tag3 key3,tag"`

			Value string `influx:"Field,field"`
		}{
			Ms:    "escape measurement",
			Ts:    ts,
			Tag1:  "other=val",
			Tag2:  "ex,value2",
			Tag3:  "new value3",
			Value: "long string",
		}

		expected := "escape\\ measurement,tag1\\,key1=other\\=val,tag2\\=key2=ex\\,value2,tag3\\ key3=new\\ value3 Field=\"long string\" " + strconv.FormatInt(ts.UnixNano(), 10)
		row := ConvertToInfluxLineProtocol(v)
		if expected != row {
			t.Errorf("expected: %s, got: %s", expected, row)
		}
	})

	t.Run("escape/field/key", func(t *testing.T) {
		ts := time.Now()

		v := struct {
			Ts time.Time `influx:",timestamp"`
			Ms string    `influx:",measurement"`

			Field1 string `influx:"field1,key1,field"`
			Field2 string `influx:"field2=key2,field"`
			Field3 string `influx:"field3 key3,field"`

			Value string `influx:"Field,field"`
		}{
			Ms:     "escape measurement",
			Ts:     ts,
			Field1: "val1",
			Field2: "val2",
			Field3: "val3",
			Value:  "long string",
		}

		expected := "escape\\ measurement field1\\,key1=\"val1\",field2\\=key2=\"val2\",field3\\ key3=\"val3\",Field=\"long string\" " + strconv.FormatInt(ts.UnixNano(), 10)
		row := ConvertToInfluxLineProtocol(v)
		if expected != row {
			t.Errorf("expected: %s, got: %s", expected, row)
		}
	})

	t.Run("escape/field/value", func(t *testing.T) {
		ts := time.Now()

		v := struct {
			Ts time.Time `influx:",timestamp"`
			Ms string    `influx:",measurement"`

			Field1 string `influx:"field1,field"`
			Field2 string `influx:"field2,field"`
		}{
			Ms:     "escape measurement",
			Ts:     ts,
			Field1: `va\l1`,
			Field2: `hotel "Queen"`,
		}

		expected := `escape\ measurement field1="va\\\\l1",field2="hotel \\\\\"Queen\\\\\"" ` + strconv.FormatInt(ts.UnixNano(), 10)
		row := ConvertToInfluxLineProtocol(v)
		if expected != row {
			t.Errorf("expected: %s, got: %s", expected, row)
		}
	})

	t.Run("strings", func(t *testing.T) {
		ts := time.Now()

		v := struct {
			Ts time.Time `influx:",timestamp"`
			Ms string    `influx:",measurement"`

			Value string `influx:"value,field"`
		}{
			Ms: "strings",
			Ts: ts,

			Value: "string with spaces",
		}

		expected := "strings value=\"string with spaces\" " + strconv.FormatInt(ts.UnixNano(), 10)
		row := ConvertToInfluxLineProtocol(v)
		if expected != row {
			t.Errorf("expected: %s, got: %s", expected, row)
		}
	})

	t.Run("floats", func(t *testing.T) {
		ts := time.Now()

		v := struct {
			Ts time.Time `influx:",timestamp"`
			Ms string    `influx:",measurement"`

			Float32 float32 `influx:"field1,field"`
			Float64 float64 `influx:"field2,field"`
		}{
			Ms: "floats",
			Ts: ts,

			Float32: 56365.234,
			Float64: 2123423424.34345531,
		}

		expected := "floats field1=56365.234,field2=2.1234234243434553e+09 " + strconv.FormatInt(ts.UnixNano(), 10)

		row := ConvertToInfluxLineProtocol(v)
		if expected != row {
			t.Errorf("expected: %s, got: %s", expected, row)
		}
	})

	t.Run("ints", func(t *testing.T) {
		ts := time.Now()

		v := struct {
			Ts time.Time `influx:",timestamp"`
			Ms string    `influx:",measurement"`

			Int   int   `influx:"field1,field"`
			Int8  int8  `influx:"field2,field"`
			Int16 int16 `influx:"field3,field"`
			Int32 int32 `influx:"field4,field"`
			Int64 int64 `influx:"field5,field"`
		}{
			Ms: "ints",
			Ts: ts,

			Int:   12,
			Int8:  101,
			Int16: 3561,
			Int32: 56365,
			Int64: 2123423424,
		}

		expected := "ints field1=12i,field2=101i,field3=3561i,field4=56365i,field5=2123423424i " + strconv.FormatInt(ts.UnixNano(), 10)

		row := ConvertToInfluxLineProtocol(v)
		if expected != row {
			t.Errorf("expected: %s, got: %s", expected, row)
		}
	})

	t.Run("uints", func(t *testing.T) {
		ts := time.Now()

		v := struct {
			Ts time.Time `influx:",timestamp"`
			Ms string    `influx:",measurement"`

			Uint   uint   `influx:"field1,field"`
			Uint8  uint8  `influx:"field2,field"`
			Uint16 uint16 `influx:"field3,field"`
			Uint32 uint32 `influx:"field4,field"`
			Uint64 uint64 `influx:"field5,field"`
		}{
			Ms: "uints",
			Ts: ts,

			Uint:   12,
			Uint8:  101,
			Uint16: 3561,
			Uint32: 56365,
			Uint64: 2123423424,
		}

		expected := "uints field1=12u,field2=101u,field3=3561u,field4=56365u,field5=2123423424u " + strconv.FormatInt(ts.UnixNano(), 10)

		row := ConvertToInfluxLineProtocol(v)
		if expected != row {
			t.Errorf("expected: %s, got: %s", expected, row)
		}
	})

	t.Run("error/measurement", func(t *testing.T) {
		ts := time.Now()

		v := struct {
			Mode   string `influx:"mode,tag"`
			Worker string `influx:"worker,tag"`

			Timestamp time.Time `influx:",timestamp"`

			Errors int     `influx:"errors,field"`
			Rate   float64 `influx:"rate,field"`

			ExecutionTime Duration `influx:"execTime,field"`
		}{
			Mode:          "aggregate",
			Worker:        "main",
			Timestamp:     ts,
			Errors:        121,
			Rate:          45.678891,
			ExecutionTime: Duration{Value: "39m47.276156216s", To: time.Minute},
		}

		row := ConvertToInfluxLineProtocol(v)
		if !strings.Contains(row, "error: `influx:\",measurement\"`") {
			t.Errorf("expected error, got: %s", row)
		}
	})

	t.Run("error/timestamp", func(t *testing.T) {
		ts := time.Now()

		v := struct {
			Mode   string `influx:",measurement"`
			Worker string `influx:"worker,tag"`

			Timestamp time.Time `influx:"ts,field"`
			Errors    int       `influx:"errors,field"`
			Rate      float64   `influx:"rate,field"`

			ExecutionTime Duration `influx:"execTime,field"`
		}{
			Mode:          "aggregate",
			Worker:        "main",
			Timestamp:     ts,
			Errors:        121,
			Rate:          45.678891,
			ExecutionTime: Duration{Value: "1m17.276156216s", To: time.Second},
		}

		row := ConvertToInfluxLineProtocol(v)
		if !strings.Contains(row, "error: `influx:\",timestamp\"`") {
			t.Errorf("expected error, got: %s", row)
		}
	})

	t.Run("error/field", func(t *testing.T) {
		ts := time.Now()

		v := struct {
			Mode   string `influx:",measurement"`
			Worker string `influx:"worker,tag"`

			Timestamp time.Time `influx:",timestamp"`

			Errors int
			Rate   float64

			ExecutionTime Duration
		}{
			Mode:          "aggregate",
			Worker:        "main",
			Timestamp:     ts,
			Errors:        121,
			Rate:          45.678891,
			ExecutionTime: Duration{Value: "1m17.276156216s", To: time.Second},
		}

		row := ConvertToInfluxLineProtocol(v)
		if !strings.Contains(row, "error: points must have at least one field") {
			t.Errorf("expected error, got: %s", row)
		}
	})

	t.Run("ok", func(t *testing.T) {
		ts := time.Now()

		v := struct {
			Mode   string `influx:",measurement"`
			Worker string `influx:"worker,tag"`

			Timestamp time.Time `influx:",timestamp"`

			Errors int     `influx:"errors,field"`
			Rate   float64 `influx:"rate,field"`

			ExecutionTime Duration `influx:"execTime,field"`
		}{
			Mode:          "aggregate",
			Worker:        "main",
			Timestamp:     ts,
			Errors:        121,
			Rate:          45.678891,
			ExecutionTime: Duration{Value: "39m47.276156216s", To: time.Minute},
		}

		expected := "aggregate,worker=main errors=121i,rate=45.678891,execTime=39.79 " + strconv.FormatInt(ts.UnixNano(), 10)

		row := ConvertToInfluxLineProtocol(v)
		if expected != row {
			t.Errorf("expected: %s, got: %s", expected, row)
		}
	})

	t.Run("duration/second", func(t *testing.T) {
		ts := time.Now()

		v := struct {
			Action        string    `influx:",measurement"`
			Worker        string    `influx:"worker,tag"`
			Timestamp     time.Time `influx:",timestamp"`
			ExecutionTime Duration  `influx:"execTime,field"`
		}{
			Action:        "upload",
			Worker:        "1",
			Timestamp:     ts,
			ExecutionTime: Duration{Value: "1m17.276156216s", To: time.Second},
		}

		expected := "upload,worker=1 execTime=77.28 " + strconv.FormatInt(ts.UnixNano(), 10)

		row := ConvertToInfluxLineProtocol(v)
		if expected != row {
			t.Errorf("expected: %s, got: %s", expected, row)
		}
	})

	t.Run("duration/minute", func(t *testing.T) {
		ts := time.Now()

		v := struct {
			Action        string    `influx:",measurement"`
			Worker        string    `influx:"worker,tag"`
			Timestamp     time.Time `influx:",timestamp"`
			ExecutionTime Duration  `influx:"execTime,field"`
		}{
			Action:        "upload",
			Worker:        "1",
			Timestamp:     ts,
			ExecutionTime: Duration{Value: "17m30.276156216s", To: time.Minute},
		}

		expected := "upload,worker=1 execTime=17.50 " + strconv.FormatInt(ts.UnixNano(), 10)

		row := ConvertToInfluxLineProtocol(v)
		if expected != row {
			t.Errorf("expected: %s, got: %s", expected, row)
		}
	})

}

type TestMeasurement struct {
	Worker        string    `influx:"worker,tag"`
	Timestamp     time.Time `influx:",timestamp"`
	ExecutionTime Duration  `influx:"execTime,field"`
}

func (m TestMeasurement) InfluxMeasurement() string { return "download" }

type TestMeasurementOverride struct {
	Name          string    `influx:",measurement"`
	Worker        string    `influx:"worker,tag"`
	Timestamp     time.Time `influx:",timestamp"`
	ExecutionTime Duration  `influx:"execTime,field"`
}

func (m TestMeasurementOverride) InfluxMeasurement() string { return "download" }

func TestInfluxMeasurement(t *testing.T) {
	t.Parallel()

	t.Run("func", func(t *testing.T) {
		ts := time.Now()

		v := TestMeasurement{
			Worker:        "1",
			Timestamp:     ts,
			ExecutionTime: Duration{Value: "17m30.276156216s", To: time.Minute},
		}

		expected := "download,worker=1 execTime=17.50 " + strconv.FormatInt(ts.UnixNano(), 10)
		row := ConvertToInfluxLineProtocol(v)
		if expected != row {
			t.Errorf("expected: %s, got: %s", expected, row)
		}
	})

	t.Run("override", func(t *testing.T) {
		ts := time.Now()

		v := TestMeasurementOverride{
			Name:          "upload",
			Worker:        "1",
			Timestamp:     ts,
			ExecutionTime: Duration{Value: "17m30.276156216s", To: time.Minute},
		}

		expected := "upload,worker=1 execTime=17.50 " + strconv.FormatInt(ts.UnixNano(), 10)
		row := ConvertToInfluxLineProtocol(v)
		if expected != row {
			t.Errorf("expected: %s, got: %s", expected, row)
		}
	})
}

type TestTimestamp struct {
	Name          string   `influx:",measurement"`
	ExecutionTime Duration `influx:"execTime,field"`
}

func (m TestTimestamp) InfluxTimestamp() time.Time {
	return time.Date(2024, time.April, 10, 23, 23, 23, 0, time.UTC)
}

type TestTimestampOverride struct {
	Name          string    `influx:",measurement"`
	Timestamp     time.Time `influx:",timestamp"`
	ExecutionTime Duration  `influx:"execTime,field"`
}

func (m TestTimestampOverride) InfluxTimestamp() time.Time {
	return time.Date(2024, time.April, 10, 23, 23, 23, 0, time.UTC)
}

func TestInfluxTimestamp(t *testing.T) {
	t.Parallel()

	t.Run("func", func(t *testing.T) {
		v := TestTimestamp{
			Name:          "backup",
			ExecutionTime: Duration{Value: "45m", To: time.Minute},
		}
		expected := "backup execTime=45.00 " + strconv.FormatInt(time.Date(2024, time.April, 10, 23, 23, 23, 0, time.UTC).UnixNano(), 10)
		row := ConvertToInfluxLineProtocol(v)
		if expected != row {
			t.Errorf("expected: %s, got: %s", expected, row)
		}
	})

	t.Run("override", func(t *testing.T) {
		ts := time.Date(2024, time.November, 15, 13, 15, 17, 0, time.UTC)
		v := TestTimestampOverride{
			Name:          "backup",
			Timestamp:     ts,
			ExecutionTime: Duration{Value: "45m", To: time.Minute},
		}

		expected := "backup execTime=45.00 " + strconv.FormatInt(ts.UnixNano(), 10)
		row := ConvertToInfluxLineProtocol(v)
		if expected != row {
			t.Errorf("expected: %s, got: %s", expected, row)
		}
	})
}
