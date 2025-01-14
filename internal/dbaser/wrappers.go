package dbaser

import "time"

var AttemptDelays = []int{1, 3, 5}

type MetricValueTypes interface {
	int64 | float64
}

func TableMetricWrapper(origFunc func(MetricBaseStruct *Struct4db, metr *Metrics) error) func(MetricBaseStruct *Struct4db, metr *Metrics) error {
	wrappedFunc := func(MetricBaseStruct *Struct4db, metr *Metrics) error {
		err := origFunc(MetricBaseStruct, metr)
		if err != nil {
			for _, delay := range AttemptDelays {
				time.Sleep(time.Duration(delay) * time.Second)
				if err = origFunc(MetricBaseStruct, metr); err == nil {
					break
				}
				//				fmt.Println(delay, " MetricWrapper !")
			}
		}
		return err
	}
	return wrappedFunc
}

func TableBuncherWrapper(origFunc func(MetricBaseStruct *Struct4db, metrArray []Metrics) error) func(MetricBaseStruct *Struct4db, metrArray []Metrics) error {
	wrappedFunc := func(MetricBaseStruct *Struct4db, metrArray []Metrics) error {
		err := origFunc(MetricBaseStruct, metrArray)
		if err != nil {
			for _, delay := range AttemptDelays {
				time.Sleep(time.Duration(delay) * time.Second)
				if err = origFunc(MetricBaseStruct, metrArray); err == nil {
					break
				}
				//				fmt.Println(delay, " BUNCHWrapper !")
			}
		}
		return err
	}
	return wrappedFunc
}

func TableGetAllsWrapper[MV MetricValueTypes](origFunc func(MetricBaseStruct *Struct4db, mappa *map[string]MV) error) func(MetricBaseStruct *Struct4db,
	mappa *map[string]MV) error {
	wrappedFunc := func(MetricBaseStruct *Struct4db, mappa *map[string]MV) error {
		err := origFunc(MetricBaseStruct, mappa)
		if err != nil {
			for _, delay := range AttemptDelays {
				time.Sleep(time.Duration(delay) * time.Second)
				if err = origFunc(MetricBaseStruct, mappa); err == nil {
					break
				}
				//				fmt.Println(delay, "TableGetAllsWrapper !")
			}
		}
		return err
	}
	return wrappedFunc
}
