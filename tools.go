package main

import (
	"context"
	"time"
)

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
