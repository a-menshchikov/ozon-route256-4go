package request

import (
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/types"
)

type SetLimit struct {
	User     *types.User
	Value    int64
	Category string
}

type ListLimits struct {
	User *types.User
}
