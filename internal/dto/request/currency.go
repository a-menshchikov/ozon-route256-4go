package request

import (
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/types"
	"go.uber.org/zap/zapcore"
)

type SetCurrency struct {
	User *types.User
	Code string
}

func (r SetCurrency) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddInt64("user", int64(*r.User))
	enc.AddString("code", r.Code)

	return nil
}

type ListCurrencies struct {
	User *types.User
}
