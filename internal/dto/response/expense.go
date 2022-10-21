package response

import (
	"time"
)

type AddExpense struct {
	Ready        bool
	LimitReached bool
	Success      bool
}

type GetReport struct {
	From     time.Time
	Ready    bool
	Currency string
	Data     map[string]int64
	Success  bool
}
