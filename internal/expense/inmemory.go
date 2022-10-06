package expense

import (
	"time"

	"github.com/pkg/errors"
)

type Storage struct {
	data map[int64][]*expensesGroup
}

type expensesGroup struct {
	category string
	expenses []*expenseItem
}

type expenseItem struct {
	amount int64
	date   time.Time
}

func NewInmemoryStorage() *Storage {
	return &Storage{
		data: map[int64][]*expensesGroup{},
	}
}

func (s *Storage) Init(userID int64) {
	s.data[userID] = []*expensesGroup{}
}

func (s *Storage) Add(userID int64, date time.Time, amount int64, category string) error {
	expense, err := newExpense(amount, date)
	if err != nil {
		return err
	}

	if _, ok := s.data[userID]; !ok {
		s.data[userID] = []*expensesGroup{{
			category: category,
			expenses: []*expenseItem{},
		}}
	}

	for _, group := range s.data[userID] {
		if group.category != category {
			continue
		}

		group.expenses = append(group.expenses, expense)

		return nil
	}

	s.data[userID] = append(s.data[userID], &expensesGroup{
		category: category,
		expenses: []*expenseItem{expense},
	})

	return nil
}

func newExpense(amount int64, date time.Time) (*expenseItem, error) {
	if amount < 0 {
		return nil, errors.New("сумма трат должна быть положительным числом")
	}

	if date.After(time.Now()) {
		return nil, errors.New("траты из будущего не поддерживаются")
	}

	return &expenseItem{
		amount: amount,
		date:   date,
	}, nil
}

func (s *Storage) List(userID int64, from time.Time) map[string]int64 {
	if _, ok := s.data[userID]; !ok {
		return nil
	}

	result := make(map[string]int64)

	for _, group := range s.data[userID] {
		result[group.category] = int64(0)

		for _, expense := range group.expenses {
			if expense.date.After(from) {
				result[group.category] += expense.amount
			}
		}
	}

	return result
}
