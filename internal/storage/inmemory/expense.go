package inmemory

import (
	"time"

	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/types"
)

type inMemoryExpenseStorage struct {
	data map[*types.User][]*expensesGroup
}

type expensesGroup struct {
	category string
	expenses []types.ExpenseItem
}

func (s *inMemoryExpenseStorage) Add(user *types.User, item types.ExpenseItem, category string) error {
	if _, ok := s.data[user]; !ok {
		s.data[user] = []*expensesGroup{{
			category: category,
			expenses: []types.ExpenseItem{item},
		}}
		return nil
	}

	for _, group := range s.data[user] {
		if group.category == category {
			group.expenses = append(group.expenses, item)
			return nil
		}
	}

	s.data[user] = append(s.data[user], &expensesGroup{
		category: category,
		expenses: []types.ExpenseItem{item},
	})

	return nil
}

func (s *inMemoryExpenseStorage) List(user *types.User, from time.Time) (map[string][]types.ExpenseItem, error) {
	if _, ok := s.data[user]; !ok {
		return nil, nil
	}

	result := make(map[string][]types.ExpenseItem)

	for _, group := range s.data[user] {
		result[group.category] = make([]types.ExpenseItem, 0)
		for _, item := range group.expenses {
			if item.Date.After(from) {
				result[group.category] = append(result[group.category], item)
			}
		}
	}

	return result, nil
}
