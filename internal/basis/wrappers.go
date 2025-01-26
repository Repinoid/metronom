package basis

import (
	"context"
	"time"
)

var AttemptDelays = []int{1, 3, 5}

func CommonMetricWrapper(origFunc func(ctx context.Context, metr *Metrics,
	metras *[]Metrics) error) func(ctx context.Context, metr *Metrics, metras *[]Metrics) error { // ---1
	wrappedFunc := func(ctx context.Context, metr *Metrics, metras *[]Metrics) error { // ---2
		err := origFunc(ctx, metr, metras)
		if err != nil { // ---3
			for _, delay := range AttemptDelays {
				time.Sleep(time.Duration(delay) * time.Second)
				err = origFunc(ctx, metr, metras)
				if err == nil {
					break
				}
			}
		} // ---3
		return err
	} // ---2
	return wrappedFunc
} // ---1
