//go:build unit

package expense

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	mocks "gitlab.ozon.dev/almenschhikov/go-course-4/internal/mocks/storage"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/test"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/types"
)

type expenserMocksInitializer struct {
	storage func(m *mocks.MockExpenseStorage)
}

func setupExpenser(t *testing.T, i expenserMocksInitializer) *expenser {
	ctrl := gomock.NewController(t)

	storageMock := mocks.NewMockExpenseStorage(ctrl)
	if i.storage != nil {
		i.storage(storageMock)
	}

	return NewExpenser(storageMock)
}

func Test_expenser_AddExpenser(t *testing.T) {
	t.Run("negative amount", func(t *testing.T) {
		// ARRANGE
		e := setupExpenser(t, expenserMocksInitializer{})

		// ACT
		err := e.AddExpense(
			context.Background(),
			test.User,
			test.Today,    // date
			int64(-10000), // amount
			"RUB",         // currency
			"",            // category
		)

		// ASSERT
		assert.Error(t, err)
	})

	t.Run("feature expense", func(t *testing.T) {
		// ARRANGE
		e := setupExpenser(t, expenserMocksInitializer{})

		// ACT
		err := e.AddExpense(
			context.Background(),
			test.User,
			test.Tomorrow, // date
			int64(10000),  // amount
			"RUB",         // currency
			"",            // category
		)

		// ASSERT
		assert.Error(t, err)
	})

	t.Run("error", func(t *testing.T) {
		// ARRANGE
		e := setupExpenser(t, expenserMocksInitializer{
			storage: func(m *mocks.MockExpenseStorage) {
				m.EXPECT().Add(gomock.AssignableToTypeOf(test.CtxInterface), test.User, types.ExpenseItem{
					Date:     test.Today,
					Amount:   20000,
					Currency: "RUB",
				}, "taxi").Return(test.SimpleError)
			},
		})

		// ACT
		err := e.AddExpense(
			context.Background(),
			test.User,
			test.Today,   // date
			int64(20000), // amount
			"RUB",        // currency
			"taxi",       // category
		)

		// ASSERT
		assert.Error(t, err)
	})

	t.Run("success", func(t *testing.T) {
		// ARRANGE
		e := setupExpenser(t, expenserMocksInitializer{
			storage: func(m *mocks.MockExpenseStorage) {
				m.EXPECT().Add(gomock.AssignableToTypeOf(test.CtxInterface), test.User, types.ExpenseItem{
					Date:     test.Today,
					Amount:   150000,
					Currency: "RUB",
				}, "coffee").Return(nil)
			},
		})

		// ACT
		err := e.AddExpense(
			context.Background(),
			test.User,
			test.Today,    // date
			int64(150000), // amount
			"RUB",         // currency
			"coffee",      // category
		)

		// ASSERT
		assert.NoError(t, err)
	})
}
