package limit

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	mocks "gitlab.ozon.dev/almenschhikov/go-course-4/internal/mocks/storage"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/types"
)

var (
	testUser    = &([]types.User{types.User(int64(123))}[0])
	simpleError = errors.New("error")
)

type mocksInitializer struct {
	storage func(*mocks.MockExpenseLimitStorage)
}

func setupLimiter(t *testing.T, i mocksInitializer) *limiter {
	ctrl := gomock.NewController(t)

	storageMock := mocks.NewMockExpenseLimitStorage(ctrl)
	if i.storage != nil {
		i.storage(storageMock)
	}

	return NewLimiter(storageMock)
}

func Test_limiter_Get(t *testing.T) {
	t.Run("error", func(t *testing.T) {
		// ARRANGE
		l := setupLimiter(t, mocksInitializer{
			storage: func(m *mocks.MockExpenseLimitStorage) {
				m.EXPECT().Get(testUser, "taxi").Return(types.LimitItem{}, false, simpleError)
			},
		})

		// ACT
		item, err := l.Get(testUser, "taxi")

		// ASSERT
		assert.Error(t, err)
		assert.Empty(t, item)
	})

	t.Run("limit not found", func(t *testing.T) {
		// ARRANGE
		l := setupLimiter(t, mocksInitializer{
			storage: func(m *mocks.MockExpenseLimitStorage) {
				m.EXPECT().Get(testUser, "").Return(types.LimitItem{}, false, nil)
			},
		})

		// ACT
		item, err := l.Get(testUser, "")

		// ASSERT
		assert.NoError(t, err)
		assert.Empty(t, item)
	})

	t.Run("found", func(t *testing.T) {
		// ARRANGE
		l := setupLimiter(t, mocksInitializer{
			storage: func(m *mocks.MockExpenseLimitStorage) {
				m.EXPECT().Get(testUser, "coffee").Return(types.LimitItem{
					Total:    1000000,
					Remains:  150000,
					Currency: "USD",
				}, true, nil)
			},
		})

		// ACT
		item, err := l.Get(testUser, "coffee")

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, types.LimitItem{
			Total:    1000000,
			Remains:  150000,
			Currency: "USD",
		}, item)
	})
}

func Test_limiter_Set(t *testing.T) {
	t.Run("negative limit", func(t *testing.T) {
		// ARRANGE
		l := setupLimiter(t, mocksInitializer{})

		// ACT
		err := l.Set(
			testUser,
			int64(-10000), // limit
			"RUB",         // currency
			"",            // category
		)

		// ASSERT
		assert.Error(t, err)
	})

	t.Run("error", func(t *testing.T) {
		// ARRANGE
		l := setupLimiter(t, mocksInitializer{
			storage: func(m *mocks.MockExpenseLimitStorage) {
				m.EXPECT().Set(testUser, int64(1000000), "RUB", "coffee").Return(simpleError)
			},
		})

		// ACT
		err := l.Set(
			testUser,
			int64(1000000), // limit
			"RUB",          // currency
			"coffee",       // category
		)

		// ASSERT
		assert.Error(t, err)
	})

	t.Run("success", func(t *testing.T) {
		// ARRANGE
		l := setupLimiter(t, mocksInitializer{
			storage: func(m *mocks.MockExpenseLimitStorage) {
				m.EXPECT().Set(testUser, int64(1200000), "RUB", "coffee").Return(nil)
			},
		})

		// ACT
		err := l.Set(
			testUser,
			int64(1200000), // limit
			"RUB",          // currency
			"coffee",       // category
		)

		// ASSERT
		assert.NoError(t, err)
	})
}

func Test_limiter_Decrease(t *testing.T) {
	t.Run("error", func(t *testing.T) {
		// ARRANGE
		l := setupLimiter(t, mocksInitializer{
			storage: func(m *mocks.MockExpenseLimitStorage) {
				m.EXPECT().Decrease(testUser, int64(100000), "taxi").Return(false, simpleError)
			},
		})

		// ACT
		ok, err := l.Decrease(
			testUser,
			int64(100000), // value
			"taxi",        // category
		)

		// ASSERT
		assert.Error(t, err)
		assert.False(t, ok)
	})

	t.Run("success", func(t *testing.T) {
		// ARRANGE
		l := setupLimiter(t, mocksInitializer{
			storage: func(m *mocks.MockExpenseLimitStorage) {
				m.EXPECT().Decrease(testUser, int64(200000), "coffee").Return(true, nil)
			},
		})

		// ACT
		ok, err := l.Decrease(
			testUser,
			int64(200000), // value
			"coffee",      // category
		)

		// ASSERT
		assert.NoError(t, err)
		assert.True(t, ok)
	})
}

func Test_limiter_Unset(t *testing.T) {
	t.Run("error", func(t *testing.T) {
		// ARRANGE
		l := setupLimiter(t, mocksInitializer{
			storage: func(m *mocks.MockExpenseLimitStorage) {
				m.EXPECT().Unset(testUser, "").Return(simpleError)
			},
		})

		// ACT
		err := l.Unset(
			testUser,
			"", // category
		)

		// ASSERT
		assert.Error(t, err)
	})

	t.Run("success", func(t *testing.T) {
		// ARRANGE
		l := setupLimiter(t, mocksInitializer{
			storage: func(m *mocks.MockExpenseLimitStorage) {
				m.EXPECT().Unset(testUser, "taxi").Return(nil)
			},
		})

		// ACT
		err := l.Unset(
			testUser,
			"taxi", // category
		)

		// ASSERT
		assert.NoError(t, err)
	})
}

func Test_limiter_List(t *testing.T) {
	t.Run("error", func(t *testing.T) {
		// ARRANGE
		l := setupLimiter(t, mocksInitializer{
			storage: func(m *mocks.MockExpenseLimitStorage) {
				m.EXPECT().List(testUser).Return(nil, false, simpleError)
			},
		})

		// ACT
		list, err := l.List(testUser)

		// ASSERT
		assert.Error(t, err)
		assert.Empty(t, list)
	})

	t.Run("not found", func(t *testing.T) {
		// ARRANGE
		l := setupLimiter(t, mocksInitializer{
			storage: func(m *mocks.MockExpenseLimitStorage) {
				m.EXPECT().List(testUser).Return(nil, false, nil)
			},
		})

		// ACT
		list, err := l.List(testUser)

		// ASSERT
		assert.NoError(t, err)
		assert.Empty(t, list)
	})

	t.Run("success", func(t *testing.T) {
		// ARRANGE
		l := setupLimiter(t, mocksInitializer{
			storage: func(m *mocks.MockExpenseLimitStorage) {
				m.EXPECT().List(testUser).Return(map[string]types.LimitItem{
					"taxi": {
						Total:    15000000,
						Remains:  6000000,
						Currency: "RUB",
					},
					"": {
						Total:    1000000,
						Remains:  600000,
						Currency: "EUR",
					},
				}, true, nil)
			},
		})

		// ACT
		list, err := l.List(testUser)

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, map[string]types.LimitItem{
			"taxi": {
				Total:    15000000,
				Remains:  6000000,
				Currency: "RUB",
			},
			"": {
				Total:    1000000,
				Remains:  600000,
				Currency: "EUR",
			},
		}, list)
	})
}
