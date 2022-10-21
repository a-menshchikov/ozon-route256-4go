package types

import (
	"time"
)

type User int64

type ExpenseItem struct {
	Date     time.Time
	Amount   int64
	Currency string
}

type LimitItem struct {
	Total    int64
	Remains  int64
	Currency string
}
