package influx

import (
	"strconv"
	"time"
)

type Duration struct {
	Value string
	To    time.Duration
}

func (d Duration) MarshalInflux() (string, error) {
	duration, err := time.ParseDuration(string(d.Value))
	if err != nil {
		return "", err
	}
	switch d.To {
	case time.Nanosecond:
		return strconv.FormatInt(duration.Nanoseconds(), 10) + "i", nil
	case time.Microsecond:
		return strconv.FormatInt(duration.Microseconds(), 10) + "i", nil
	case time.Millisecond:
		return strconv.FormatInt(duration.Milliseconds(), 10) + "i", nil
	case time.Second:
		return strconv.FormatFloat(duration.Seconds(), 'f', 2, 64), nil
	case time.Minute:
		return strconv.FormatFloat(duration.Minutes(), 'f', 2, 64), nil
	case time.Hour:
		return strconv.FormatFloat(duration.Hours(), 'f', 2, 64), nil
	default:
		return d.Value, nil // return the value as is
	}
}
