package request

import (
	"time"

	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/types"
	"go.uber.org/zap/zapcore"
)

type AddExpense struct {
	User     *types.User
	Date     time.Time
	Amount   int64
	Category string
}

func (r AddExpense) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddInt64("user", int64(*r.User))
	enc.AddTime("date", r.Date)
	enc.AddInt64("amount", r.Amount)
	enc.AddString("category", r.Category)

	return nil
}

type GetReport struct {
	User *types.User
	From time.Time
}

func (r GetReport) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddInt64("user", int64(*r.User))
	enc.AddTime("from", r.From)

	return nil
}
