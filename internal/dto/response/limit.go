package response

import (
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/types"
)

type SetLimit bool

type ListLimits struct {
	Ready           bool
	CurrentCurrency string
	List            map[string]LimitItem
	Success         bool
}

type LimitItem struct {
	Total   int64
	Remains int64

	Origin types.LimitItem
}
