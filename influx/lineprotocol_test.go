package influx

import (
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestConvertToInfluxLineProtocol(t *testing.T) {
	t.Parallel()

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
