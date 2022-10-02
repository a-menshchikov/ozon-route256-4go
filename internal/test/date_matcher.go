package test

import (
	"fmt"
	"time"
)

type DateMatcher struct {
	date time.Time
}

func SameDate(date time.Time) DateMatcher {
	return DateMatcher{date.Truncate(24 * time.Hour)}
}

func (m DateMatcher) Matches(x interface{}) bool {
	d, ok := x.(time.Time)
	if !ok {
		return false
	}

	return d.Truncate(24 * time.Hour).Equal(m.date)
}

func (m DateMatcher) String() string {
	return fmt.Sprintf("equals %v (%T)", m.date, m.date)
}
