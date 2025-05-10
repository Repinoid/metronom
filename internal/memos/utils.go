package memos

import (
	"errors"
	"strconv"
	"strings"
)

// takeInterval возвращает значение интервала в секундах и ошибку
func TakeInterval(s string) (t int, err error) {
	if s == "" {
		return 0, nil
	}
	sec, isSec := strings.CutSuffix(s, "s")
	if isSec {
		hm, err := strconv.Atoi(sec)
		if err != nil {
			return 0, err
		}
		return hm, nil
	}
	min, isMin := strings.CutSuffix(s, "m") // на всяк случай минуты
	if isMin {
		hm, err := strconv.Atoi(min)
		if err != nil {
			return 0, err
		}
		return hm * 60, nil
	}
	return 0, errors.New("bad Interval format")
}
