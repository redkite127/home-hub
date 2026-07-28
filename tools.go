package main

import (
	"context"
	"log"
	"math"
	"time"

	"github.com/influxdata/influxdb-client-go/v2/api/write"
)

// roundTo rounds at the write boundary, so InfluxDB never stores more precision
// than the sensor actually reports.
func roundTo(v float64, digits int) float64 {
	m := math.Pow(10, float64(digits))

	return math.Round(v*m) / m
}

// addFieldRounded skips missing readings: a nil value means the sensor was
// unavailable this cycle and must not be written as a zero.
func addFieldRounded(p *write.Point, key string, v *float64, digits int) {
	if v == nil {
		return
	}

	p.AddField(key, roundTo(*v, digits))
}

func writePoint(p *write.Point) {
	if dryRun {
		tags := map[string]string{}
		for _, t := range p.TagList() {
			tags[t.Key] = t.Value
		}
		fields := map[string]interface{}{}
		for _, f := range p.FieldList() {
			fields[f.Key] = f.Value
		}
		log.Printf("dry-run: %s tags=%v fields=%v time=%s", p.Name(), tags, fields, p.Time().Format(time.RFC3339))
		return
	}
	influxWriter.WritePoint(p)
}

func flushPoints() {
	if dryRun {
		return
	}
	influxWriter.Flush()
}

func scheduler(ctx context.Context, f func() error, d time.Duration) error {
	// run a first time
	if err := f(); err != nil {
		return err
	}

	// then run every d
	ticker := time.NewTicker(d)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := f(); err != nil {
				return err
			}
		}
	}
}
