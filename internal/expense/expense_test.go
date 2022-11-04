package expense

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	mocks "gitlab.ozon.dev/almenschhikov/go-course-4/internal/mocks/storage"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/types"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/utils"
)

var (
	testUser    = &([]types.User{types.User(int64(123))}[0])
	today       = utils.TruncateToDate(time.Now())
	tomorrow    = today.Add(24 * time.Hour)
	simpleError = errors.New("error")
)

type mocksInitializer struct {
	storage func(*mocks.MockExpenseStorage)
}

func setupExpenser(t *testing.T, i mocksInitializer) *expenser {
	ctrl := gomock.NewController(t)

	storageMock := mocks.NewMockExpenseStorage(ctrl)
	if i.storage != nil {
		i.storage(storageMock)
	}

	return NewExpenser(storageMock)
}

func Test_expenser_Add(t *testing.T) {
	t.Run("negative amount", func(t *testing.T) {
		// ARRANGE
		e := setupExpenser(t, mocksInitializer{})

		// ACT
		err := e.Add(
			testUser,
			today,         // date
			int64(-10000), // amount
			"RUB",         // currency
			"",            // category
		)

		// ASSERT
		assert.Error(t, err)
	})

	t.Run("feature expense", func(t *testing.T) {
		// ARRANGE
		e := setupExpenser(t, mocksInitializer{})

		// ACT
		err := e.Add(
			testUser,
			tomorrow,     // date
			int64(10000), // amount
			"RUB",        // currency
			"",           // category
		)

		// ASSERT
		assert.Error(t, err)
	})

	t.Run("error", func(t *testing.T) {
		// ARRANGE
		e := setupExpenser(t, mocksInitializer{
			storage: func(m *mocks.MockExpenseStorage) {
				m.EXPECT().Add(testUser, types.ExpenseItem{
					Date:     today,
					Amount:   20000,
					Currency: "RUB",
				}, "taxi").Return(simpleError)
			},
		})

		// ACT
		err := e.Add(
			testUser,
			today,        // date
			int64(20000), // amount
			"RUB",        // currency
			"taxi",       // category
		)

		// ASSERT
		assert.Error(t, err)
	})

	t.Run("success", func(t *testing.T) {
		// ARRANGE
		e := setupExpenser(t, mocksInitializer{
			storage: func(m *mocks.MockExpenseStorage) {
				m.EXPECT().Add(testUser, types.ExpenseItem{
					Date:     today,
					Amount:   150000,
					Currency: "RUB",
				}, "coffee").Return(nil)
			},
		})

		// ACT
		err := e.Add(
			testUser,
			today,         // date
			int64(150000), // amount
			"RUB",         // currency
			"coffee",      // category
		)

		// ASSERT
		assert.NoError(t, err)
	})
}

func Test_expenser_Report(t *testing.T) {
	t.Run("error", func(t *testing.T) {
		// ARRANGE
		e := setupExpenser(t, mocksInitializer{
			storage: func(m *mocks.MockExpenseStorage) {
				m.EXPECT().List(testUser, today).Return(nil, simpleError)
			},
		})

		// ACT
		data, err := e.Report(testUser, today)

		// ASSERT
		assert.Error(t, err)
		assert.Empty(t, data)
	})

	t.Run("success", func(t *testing.T) {
		// ARRANGE
		e := setupExpenser(t, mocksInitializer{
			storage: func(m *mocks.MockExpenseStorage) {
				m.EXPECT().List(testUser, today).Return(map[string][]types.ExpenseItem{
					"taxi": {
						{
							Date:     today,
							Amount:   100000,
							Currency: "USD",
						},
						{
							Date:     today,
							Amount:   120000,
							Currency: "EUR",
						},
					},
					"coffee": {
						{
							Date:     today,
							Amount:   1200000,
							Currency: "RUB",
						},
					},
				}, nil)
			},
		})

		// ACT
		data, err := e.Report(testUser, today)

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, map[string][]types.ExpenseItem{
			"taxi": {
				{
					Date:     today,
					Amount:   100000,
					Currency: "USD",
				},
				{
					Date:     today,
					Amount:   120000,
					Currency: "EUR",
				},
			},
			"coffee": {
				{
					Date:     today,
					Amount:   1200000,
					Currency: "RUB",
				},
			},
		}, data)
	})
}
