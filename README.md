# Custom Go struct tags

[![codecov](https://codecov.io/gh/1buran/custom-tags/graph/badge.svg?token=QH2B55P9AD)](https://codecov.io/gh/1buran/custom-tags)
[![Go Reference](https://pkg.go.dev/badge/github.com/1buran/custom-tags.svg)](https://pkg.go.dev/github.com/1buran/custom-tags)
[![goreportcard](https://goreportcard.com/badge/github.com/1buran/custom-tags)](https://goreportcard.com/report/github.com/1buran/custom-tags)

This is lib for dump Go structs to different exotic formats (currently supported only influxdb line protocol).

## Getting Started

> [!CAUTION]
> This is alpha stage, so library api may change!

Convert Go structs to influxdb line protocol format.

The structs should describe their reflection to influx line protocol format with struct tags:
- tag name represents the name of measurement, `tag` or `field `of line protocol row,
- tag value represents the type of data: `measurement`, `tag`, `field` or `timestamp`.

The only one measurement and timestamp should be, for these types of data the name of tag
may be omitted cos it will not be used.

Example:
```go
import (
  "time"
  "fmt"
  "github.com/1buran/custom-tags/influx"
)

type Node struct {
  Operation     string        `influx:",measurement"` // name is omitted cos will not used
  DataCenter    string        `influx:"dc,tag"`
  CloudProvider string        `influx:"cloud,tag"`
  Errors        int           `influx:"errors,field"`
  ExecutionTime time.Duration `influx:"time,field"`
  Timestamp     time.Time     `influx:",timestamp"` // name is omitted cos will not used
}

func (n Node) String() string { return influx.ConvertToInfluxLineProtocol(n) }

var backupDuration time.Duration = ... // calculate or get the time of backup creation
v := Node{
  Operation: "backup", DataCenter: "east-1", CloudProvider: "AWS",
  ExecutionTime: backupDuration, Timestamp: time.Now(),
}

// expected result:
v.String() == "backup,dc=east-1,cloud=AWS errors=0,time=45m30.233456987s 1735137974129911864"

// you may easy write metrics to file:
fmt.Fprintf(&fileMetrics, v)
```
Did you notice? The value of time duration `time=45m12.233456987s` is not what we are wanted,
this is just a plain string which is useless to show it as value on the graphs,
unfortunately influxdb does not support `time.Duration` so you have to convert this to
meaningful value: seconds, minutes, hours etc what is more suitable for you.

You may just store nanoseconds as golang time duration with maximum of precision,
but this is not handy in further representation on graphs.

That is why `custom-tags` has own handy type `Duration` which store duration
with preferred time fractions and maximum 2 digits of float fraction.

So let's rewrite struct:
```go
type Node struct {
  ...
  ExecutionTime influx.Duration `influx:"time,field"`
  ...
}

var backupDuration time.Duration = ... // calculate or get the time of backup creation
v := Node{
  Operation: "backup", DataCenter: "east-1", CloudProvider: "AWS",
  ExecutionTime: influx.Duration{Value: backupDuration.String(), To: time.Minute}, Timestamp: time.Now(),
}

// expected result:
v.String() == "backup,dc=east-1,cloud=AWS errors=0,time=45.50 1735137974129911864"
```
now it looks much pretty and became much easy to show on the graphs.

https://docs.influxdata.com/influxdb/v2/reference/syntax/line-protocol/

## Custom methods for getting measurement and timestamp of data

Typically you always have the field of struct which can be used as `measurement`,
in terms of influxdb, this is the name of measurement.

But if you have no such field or do not want store extra data in your struct,
you may define `InfluxMeasurement() string` method:

```go
type UploadMetrics struct {
	Ts    time.Time       `influx:",timestamp"`
	Time  time.Duration   `influx:"time,field"`
	Speed float64         `influx:"speed,field"`
}

func (u UploadMetrics) InfluxMeasurement() string { return "upload" }
```

The same thing is correct for another one special required influx line protocol row field - timestamp.
You may define it as `InfluxTimestamp() time.Time` method:
```go
type UploadMetrics struct {
	Time  time.Duration `influx:"time,field"`
	Speed float64       `influx:"speed,field"`
}

func (u UploadMetrics) InfluxMeasurement() string { return "upload" }
func (u UploadMetrics) InfluxTimestamp() time.Time { return time.Now() }
```

and then use it in code:

```go
func main() {
  stats := UploadMetrics{}
  ...
  t := time.Now()
  Upload(file)
  stats.Time = time.Since(t)
  stats.Speed = file.size / stats.Time.Seconds()
  fmt.Fprint(&metricsFile, stats)
}
```

or more shorter variant:

```go
func main() {
  ...
  t := time.Now()
  Upload(file)
  fmt.Fprint(&metricsFile, UploadMetrics{Time: time.Since(t), Speed: file.size / stats.Time.Seconds()})
}
```
it works fine cos the metrics dumps to disk at the moment of their creation,
in real world the mentioned immplementation `InfluxTimestamp() time.Time { return time.Now() }`
is not really what do you want, but having the ability of dynamic construction of measurement timestamp
may be very useful in some situations.
