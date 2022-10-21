package request

import (
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/types"
)

type SetCurrency struct {
	User *types.User
	Code string
}

type ListCurrencies struct {
	User *types.User
}
