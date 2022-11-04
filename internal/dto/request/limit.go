package request

import (
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/types"
	"go.uber.org/zap/zapcore"
)

type SetLimit struct {
	User     *types.User
	Value    int64
	Category string
}

func (r SetLimit) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddInt64("user", int64(*r.User))
	enc.AddInt64("value", r.Value)
	enc.AddString("category", r.Category)

	return nil
}

type ListLimits struct {
	User *types.User
}

func (r ListLimits) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddInt64("user", int64(*r.User))

	return nil
}
