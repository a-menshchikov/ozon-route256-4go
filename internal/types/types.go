package types

import (
	"strconv"
	"time"

	"go.uber.org/zap/zapcore"
)

type User int64

func (u *User) String() string {
	return strconv.Itoa(int(*u))
}

type ExpenseItem struct {
	Date     time.Time
	Amount   int64
	Currency string
}

type Report struct {
	Data    map[string]int64
	Success bool
	Error   string
}

type LimitItem struct {
	Total    int64
	Remains  int64
	Currency string
}

func (l LimitItem) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddInt64("total", l.Total)
	enc.AddInt64("remains", l.Remains)
	enc.AddString("currency", l.Currency)

	return nil
}
