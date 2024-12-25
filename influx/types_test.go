package influx

import (
	"strings"
	"testing"
	"time"
)

func TestDuration(t *testing.T) {
	t.Parallel()
	t.Run("default", func(t *testing.T) {
		d := Duration{Value: "13h23m12.234444234s"}
		influxRepr, err := d.MarshalInflux()
		if err != nil {
			t.Error(err)
		}
		if "13h23m12.234444234s" != influxRepr {
			t.Errorf("expected 13h23m12.234444234s, got: %s", influxRepr)
		}
	})
	t.Run("error", func(t *testing.T) {
		d := Duration{Value: "345.12"}
		influxRepr, err := d.MarshalInflux()
		if err == nil {
			t.Error("expected error not found")
		}
		if !strings.Contains(err.Error(), "missing unit in duration") {
			t.Errorf("expected 'missing unit in duration' error, got: %s", err)
		}
		if "" != influxRepr {
			t.Errorf("expected empty, got: %s", influxRepr)
		}
	})
	t.Run("int/nano", func(t *testing.T) {
		d := Duration{Value: "12ns", To: time.Nanosecond}
		influxRepr, err := d.MarshalInflux()
		if err != nil {
			t.Error(err)
		}
		if "12i" != influxRepr {
			t.Errorf("expected 12i, got: %s", influxRepr)
		}
	})
	t.Run("int/micro", func(t *testing.T) {
		d := Duration{Value: "12µs", To: time.Microsecond}
		influxRepr, err := d.MarshalInflux()
		if err != nil {
			t.Error(err)
		}
		if "12i" != influxRepr {
			t.Errorf("expected 12i, got: %s", influxRepr)
		}
	})
	t.Run("int/mili", func(t *testing.T) {
		d := Duration{Value: "12ms", To: time.Millisecond}
		influxRepr, err := d.MarshalInflux()
		if err != nil {
			t.Error(err)
		}
		if "12i" != influxRepr {
			t.Errorf("expected 12i, got: %s", influxRepr)
		}
	})
	t.Run("float/seconds", func(t *testing.T) {
		d := Duration{Value: "12.223434567s", To: time.Second}
		influxRepr, err := d.MarshalInflux()
		if err != nil {
			t.Error(err)
		}
		if "12.22" != influxRepr {
			t.Errorf("expected 12.22, got: %s", influxRepr)
		}
	})
	t.Run("float/minutes", func(t *testing.T) {
		d := Duration{Value: "12m30.223434567s", To: time.Minute}
		influxRepr, err := d.MarshalInflux()
		if err != nil {
			t.Error(err)
		}
		if "12.50" != influxRepr {
			t.Errorf("expected 12.50, got: %s", influxRepr)
		}
	})
	t.Run("float/hours", func(t *testing.T) {
		d := Duration{Value: "12h30m11.223434567s", To: time.Hour}
		influxRepr, err := d.MarshalInflux()
		if err != nil {
			t.Error(err)
		}
		if "12.50" != influxRepr {
			t.Errorf("expected 12.50, got: %s", influxRepr)
		}
	})
}
