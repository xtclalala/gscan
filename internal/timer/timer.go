package timer

import (
	"context"
	"time"
)

func RunTime(c context.Context, count uint, timing time.Duration, f func()) {
	go func(c context.Context, count uint, timing time.Duration, f func()) {
		t := time.NewTicker(timing)
		var runCount uint = 0
		for {
			select {
			case <-c.Done():
				return
			case <-t.C:
				runCount += 1
				f()
				if runCount == count {
					return
				}
			}
		}
	}(c, count, timing, f)
}
