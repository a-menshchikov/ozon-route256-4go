//go:build unit

package currency

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/config"
	mocks "gitlab.ozon.dev/almenschhikov/go-course-4/internal/mocks/storage"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/test"
)

type currencyManagerMocksInitializer struct {
	storage func(m *mocks.MockCurrencyStorage)
}

func setupCurrencyManager(t *testing.T, cfg config.CurrencyConfig, i currencyManagerMocksInitializer) *currencyManager {
	ctrl := gomock.NewController(t)

	storageMock := mocks.NewMockCurrencyStorage(ctrl)
	if i.storage != nil {
		i.storage(storageMock)
	}

	return NewCurrencyManager(cfg, storageMock)
}

func Test_manager_Get(t *testing.T) {
	t.Run("error", func(t *testing.T) {
		// ARRANGE
		m := setupCurrencyManager(t, test.DefaultCurrencyCfg, currencyManagerMocksInitializer{
			storage: func(m *mocks.MockCurrencyStorage) {
				m.EXPECT().Get(gomock.AssignableToTypeOf(test.CtxInterface), test.User).Return("", false, test.SimpleError)
			},
		})

		// ACT
		currency, err := m.Get(context.Background(), test.User)

		// ASSERT
		assert.Error(t, err)
		assert.Empty(t, currency)
	})

	t.Run("success", func(t *testing.T) {
		// ARRANGE
		m := setupCurrencyManager(t, test.DefaultCurrencyCfg, currencyManagerMocksInitializer{
			storage: func(m *mocks.MockCurrencyStorage) {
				m.EXPECT().Get(gomock.AssignableToTypeOf(test.CtxInterface), test.User).Return("EUR", true, nil)
			},
		})

		// ACT
		currency, err := m.Get(context.Background(), test.User)

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, "EUR", currency)
	})

	t.Run("default", func(t *testing.T) {
		// ARRANGE
		m := setupCurrencyManager(t, test.DefaultCurrencyCfg, currencyManagerMocksInitializer{
			storage: func(m *mocks.MockCurrencyStorage) {
				m.EXPECT().Get(gomock.AssignableToTypeOf(test.CtxInterface), test.User).Return("", false, nil)
			},
		})

		// ACT
		currency, err := m.Get(context.Background(), test.User)

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, "USD", currency)
	})
}

func Test_manager_Set(t *testing.T) {
	t.Run("unknown", func(t *testing.T) {
		// ARRANGE
		m := setupCurrencyManager(t, test.DefaultCurrencyCfg, currencyManagerMocksInitializer{})

		// ACT
		err := m.Set(context.Background(), test.User, "RUB")

		// ASSERT
		assert.Error(t, err)
	})

	t.Run("error", func(t *testing.T) {
		// ARRANGE
		m := setupCurrencyManager(t, test.DefaultCurrencyCfg, currencyManagerMocksInitializer{
			storage: func(m *mocks.MockCurrencyStorage) {
				m.EXPECT().Set(gomock.AssignableToTypeOf(test.CtxInterface), test.User, "EUR").Return(test.SimpleError)
			},
		})

		// ACT
		err := m.Set(context.Background(), test.User, "EUR")

		// ASSERT
		assert.Error(t, err)
	})

	t.Run("success", func(t *testing.T) {
		// ARRANGE
		m := setupCurrencyManager(t, test.DefaultCurrencyCfg, currencyManagerMocksInitializer{
			storage: func(m *mocks.MockCurrencyStorage) {
				m.EXPECT().Set(gomock.AssignableToTypeOf(test.CtxInterface), test.User, "USD").Return(nil)
			},
		})

		// ACT
		err := m.Set(context.Background(), test.User, "USD")

		// ASSERT
		assert.NoError(t, err)
	})
}

func Test_manager_ListCurrenciesCodesWithFlags(t *testing.T) {
	t.Run("list", func(t *testing.T) {
		// ARRANGE
		m := setupCurrencyManager(t, test.DefaultCurrencyCfg, currencyManagerMocksInitializer{
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
