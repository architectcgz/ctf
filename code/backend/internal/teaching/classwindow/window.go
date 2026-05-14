package classwindow

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	DateLayout  = "2006-01-02"
	DefaultDays = 7
	MaxDays     = 31
)

var (
	errMissingPair = errors.New("from_date 和 to_date 必须同时传入")
	errRangeOrder  = errors.New("to_date 不能早于 from_date")
	errRangeTooBig = fmt.Errorf("时间范围不能超过 %d 天", MaxDays)
)

type Range struct {
	FromDate     string
	ToDate       string
	Days         int
	Since        time.Time
	StartOfDay   time.Time
	EndExclusive time.Time
}

func Parse(now time.Time, fromDate, toDate string) (Range, error) {
	current := now.UTC()
	if current.IsZero() {
		current = time.Now().UTC()
	}

	normalizedFrom := strings.TrimSpace(fromDate)
	normalizedTo := strings.TrimSpace(toDate)
	if (normalizedFrom == "") != (normalizedTo == "") {
		return Range{}, errMissingPair
	}

	if normalizedFrom == "" && normalizedTo == "" {
		end := startOfDate(current)
		start := end.AddDate(0, 0, -(DefaultDays - 1))
		return newRange(start, end), nil
	}

	start, err := time.ParseInLocation(DateLayout, normalizedFrom, time.UTC)
	if err != nil {
		return Range{}, fmt.Errorf("from_date 必须是 %s 格式", DateLayout)
	}
	end, err := time.ParseInLocation(DateLayout, normalizedTo, time.UTC)
	if err != nil {
		return Range{}, fmt.Errorf("to_date 必须是 %s 格式", DateLayout)
	}
	if end.Before(start) {
		return Range{}, errRangeOrder
	}

	days := inclusiveDays(start, end)
	if days > MaxDays {
		return Range{}, errRangeTooBig
	}
	return newRange(start, end), nil
}

func newRange(start, end time.Time) Range {
	startUTC := startOfDate(start)
	endUTC := startOfDate(end)
	return Range{
		FromDate:     startUTC.Format(DateLayout),
		ToDate:       endUTC.Format(DateLayout),
		Days:         inclusiveDays(startUTC, endUTC),
		Since:        startUTC,
		StartOfDay:   startUTC,
		EndExclusive: endUTC.AddDate(0, 0, 1),
	}
}

func startOfDate(value time.Time) time.Time {
	utc := value.UTC()
	return time.Date(utc.Year(), utc.Month(), utc.Day(), 0, 0, 0, 0, time.UTC)
}

func inclusiveDays(start, end time.Time) int {
	return int(end.Sub(start)/(24*time.Hour)) + 1
}
