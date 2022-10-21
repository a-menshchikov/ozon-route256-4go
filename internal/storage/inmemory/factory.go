package inmemory

import (
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/storage"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/types"
)

type factory struct{}

func NewFactory() *factory {
	return &factory{}
}

func (f *factory) CreateTelegramUserStorage() storage.TelegramUserStorage {
	return &inMemoryTelegramUserStorage{
		data: make(map[int64]*types.User),
	}
}

func (f *factory) CreateExpenseStorage() storage.ExpenseStorage {
	return &inMemoryExpenseStorage{
		data: make(map[*types.User][]*expensesGroup),
	}
}

func (f *factory) CreateLimitStorage() storage.ExpenseLimitStorage {
	return &inMemoryExpenseLimitStorage{
		data: make(map[*types.User]map[string]types.LimitItem),
	}
}

func (f *factory) CreateCurrencyStorage() storage.CurrencyStorage {
	return &inMemoryCurrencyStorage{
		data: make(map[*types.User]string),
	}
}

func (f *factory) CreateRatesStorage() storage.CurrencyRatesStorage {
	return &inMemoryCurrencyRatesStorage{
		data: make(map[string]int64),
	}
}
