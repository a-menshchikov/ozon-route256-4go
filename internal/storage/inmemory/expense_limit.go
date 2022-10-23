package inmemory

import (
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/types"
)

type inMemoryExpenseLimitStorage struct {
	data map[*types.User]map[string]types.LimitItem
}

func (s *inMemoryExpenseLimitStorage) Get(user *types.User, category string) (types.LimitItem, bool, error) {
	if limits, found := s.data[user]; found {
		if limit, found := limits[category]; found {
			return limit, found, nil
		}

		if limit, found := limits[""]; found {
			return limit, found, nil
		}
	}

	return types.LimitItem{}, false, nil
}

func (s *inMemoryExpenseLimitStorage) Set(user *types.User, total int64, currency, category string) error {
	item := types.LimitItem{
		Total:    total,
		Remains:  total,
		Currency: currency,
	}

	if _, ok := s.data[user]; ok {
		s.data[user][category] = item
	} else {
		s.data[user] = map[string]types.LimitItem{category: item}
	}

	return nil
}

func (s *inMemoryExpenseLimitStorage) Decrease(user *types.User, value int64, category string) (bool, error) {
	if _, ok := s.data[user]; !ok {
		return true, nil
	}

	item, ok := s.data[user][category]
	if !ok {
		category = ""
		item, ok = s.data[user][category]
	}
	if !ok {
		return true, nil
	}

	item.Remains -= value
	if item.Remains < 0 {
		item.Remains = 0
	}

	s.data[user][category] = item

	return item.Remains == 0, nil
}

func (s *inMemoryExpenseLimitStorage) Unset(user *types.User, category string) error {
	if _, ok := s.data[user]; ok {
		delete(s.data[user], category)
	}

	return nil
}

func (s *inMemoryExpenseLimitStorage) List(user *types.User) (map[string]types.LimitItem, bool, error) {
	limits, found := s.data[user]

	return limits, found, nil
}
