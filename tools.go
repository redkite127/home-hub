package main

import (
	"context"
	"log"
	"time"

	"github.com/influxdata/influxdb-client-go/v2/api/write"
)

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
