package basis

import (
	"context"
	"time"
)

var AttemptDelays = []int{1, 3, 5}

func GetMetricWrapper(origFunc func(ctx context.Context, metr *Metrics) (Metrics,
	error)) func(ctx context.Context, metr *Metrics) (Metrics, error) {
	wrappedFunc := func(ctx context.Context, metr *Metrics) (Metrics, error) {
		metrix, err := origFunc(ctx, metr)
		if err != nil {
			for _, delay := range AttemptDelays {
				time.Sleep(time.Duration(delay) * time.Second)
				metrix, err = origFunc(ctx, metr)
				if err == nil {
					break
				}
			}
		}
		return metrix, err
	}
	return wrappedFunc
}
func PutMetricWrapper(origFunc func(ctx context.Context, metr *Metrics) error) func(ctx context.Context, metr *Metrics) error {
	wrappedFunc := func(ctx context.Context, metr *Metrics) error {
		err := origFunc(ctx, metr)
		if err != nil {
			for _, delay := range AttemptDelays {
				time.Sleep(time.Duration(delay) * time.Second)
				err := origFunc(ctx, metr)
				if err == nil {
					break
				}
			}
		}
		return err
	}
	return wrappedFunc
}

func GetAllMetricsWrapper(origFunc func(ctx context.Context) (*[]Metrics, error)) func(ctx context.Context) (*[]Metrics, error) {
	wrappedFunc := func(ctx context.Context) (*[]Metrics, error) {
		metras, err := origFunc(ctx)
		if err != nil {
			for _, delay := range AttemptDelays {
				time.Sleep(time.Duration(delay) * time.Second)
				metras, err = origFunc(ctx)
				if err == nil {
					break
				}
			}
		}
		return metras, err
	}
	return wrappedFunc
}

// PutAllMetrics(ctx context.Context, metras *[]Metrics) error
func PutAllMetricsWrapper(origFunc func(ctx context.Context, metras *[]Metrics) error) func(ctx context.Context, metras *[]Metrics) error {
	wrappedFunc := func(ctx context.Context, metras *[]Metrics) error {
		err := origFunc(ctx, metras)
		if err != nil {
			for _, delay := range AttemptDelays {
				time.Sleep(time.Duration(delay) * time.Second)
				err := origFunc(ctx, metras)
				if err == nil {
					break
				}
			}
		}
		return err
	}
	return wrappedFunc
}
