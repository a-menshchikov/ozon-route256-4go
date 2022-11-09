package currency

import (
	"context"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/config"
	mocks "gitlab.ozon.dev/almenschhikov/go-course-4/internal/mocks/storage"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/types"
)

var (
	ctxInterface = reflect.TypeOf((*context.Context)(nil)).Elem()

	testUser    = &([]types.User{types.User(int64(123))}[0])
	simpleError = errors.New("error")
	defaultCfg  = config.CurrencyConfig{
		Base: "USD",
		Available: []config.Currency{
			{
				Code: "USD",
				Flag: "$",
			},
			{
				Code: "EUR",
				Flag: "¢",
			},
		},
	}
)

type mocksInitializer struct {
	storage func(m *mocks.MockCurrencyStorage)
}

func setupManager(t *testing.T, cfg config.CurrencyConfig, i mocksInitializer) *manager {
	ctrl := gomock.NewController(t)

	storageMock := mocks.NewMockCurrencyStorage(ctrl)
	if i.storage != nil {
		i.storage(storageMock)
	}

	return NewManager(cfg, storageMock)
}

func Test_manager_Get(t *testing.T) {
	t.Run("error", func(t *testing.T) {
		// ARRANGE
		m := setupManager(t, defaultCfg, mocksInitializer{
			storage: func(m *mocks.MockCurrencyStorage) {
				m.EXPECT().Get(gomock.AssignableToTypeOf(ctxInterface), testUser).Return("", false, simpleError)
			},
		})

		// ACT
		currency, err := m.Get(context.Background(), testUser)

		// ASSERT
		assert.Error(t, err)
		assert.Empty(t, currency)
	})

	t.Run("success", func(t *testing.T) {
		// ARRANGE
		m := setupManager(t, defaultCfg, mocksInitializer{
			storage: func(m *mocks.MockCurrencyStorage) {
				m.EXPECT().Get(gomock.AssignableToTypeOf(ctxInterface), testUser).Return("EUR", true, nil)
			},
		})

		// ACT
		currency, err := m.Get(context.Background(), testUser)

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, "EUR", currency)
	})

	t.Run("default", func(t *testing.T) {
		// ARRANGE
		m := setupManager(t, defaultCfg, mocksInitializer{
			storage: func(m *mocks.MockCurrencyStorage) {
				m.EXPECT().Get(gomock.AssignableToTypeOf(ctxInterface), testUser).Return("", false, nil)
			},
		})

		// ACT
		currency, err := m.Get(context.Background(), testUser)

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, "USD", currency)
	})
}

func Test_manager_Set(t *testing.T) {
	t.Run("unknown", func(t *testing.T) {
		// ARRANGE
		m := setupManager(t, defaultCfg, mocksInitializer{})

		// ACT
		err := m.Set(context.Background(), testUser, "RUB")

		// ASSERT
		assert.Error(t, err)
	})

	t.Run("error", func(t *testing.T) {
		// ARRANGE
		m := setupManager(t, defaultCfg, mocksInitializer{
			storage: func(m *mocks.MockCurrencyStorage) {
				m.EXPECT().Set(gomock.AssignableToTypeOf(ctxInterface), testUser, "EUR").Return(simpleError)
			},
		})

		// ACT
		err := m.Set(context.Background(), testUser, "EUR")

		// ASSERT
		assert.Error(t, err)
	})

	t.Run("success", func(t *testing.T) {
		// ARRANGE
		m := setupManager(t, defaultCfg, mocksInitializer{
			storage: func(m *mocks.MockCurrencyStorage) {
				m.EXPECT().Set(gomock.AssignableToTypeOf(ctxInterface), testUser, "USD").Return(nil)
			},
		})

		// ACT
		err := m.Set(context.Background(), testUser, "USD")

		// ASSERT
		assert.NoError(t, err)
	})
}

func Test_manager_ListCurrenciesCodesWithFlags(t *testing.T) {
	t.Run("list", func(t *testing.T) {
		// ARRANGE
		m := setupManager(t, defaultCfg, mocksInitializer{
			storage: func(m *mocks.MockCurrencyStorage) {},
		})

		// ACT
		list := m.ListCurrenciesCodesWithFlags()

		// ASSERT
		assert.Equal(t, []string{
			"USD $",
			"EUR ¢",
		}, list)
	})
}
