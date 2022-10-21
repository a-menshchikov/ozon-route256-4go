package request

import (
	"time"

	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/types"
)

type AddExpense struct {
	User     *types.User
	Date     time.Time
	Amount   int64
	Category string
}

type GetReport struct {
	User *types.User
	From time.Time
}
