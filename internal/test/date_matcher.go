package test

import (
	"fmt"
	"time"
)

type DateMatcher struct {
	date time.Time
}

func SameDate(date time.Time) DateMatcher {
	year, month, day := date.Date()

	return DateMatcher{time.Date(year, month, day, 0, 0, 0, 0, time.UTC)}
}

func (m DateMatcher) Matches(x interface{}) bool {
	date, ok := x.(time.Time)
	if !ok {
		return false
	}

	year, month, day := date.Date()

	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC).Equal(m.date)
}

func (m DateMatcher) String() string {
	return fmt.Sprintf("equals %v (%T)", m.date, m.date)
}
