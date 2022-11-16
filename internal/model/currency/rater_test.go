//go:build unit

package currency

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/config"
	cmocks "gitlab.ozon.dev/almenschhikov/go-course-4/internal/mocks/model/currency"
	smocks "gitlab.ozon.dev/almenschhikov/go-course-4/internal/mocks/storage"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/test"
	"go.uber.org/zap"
)

type raterMocksInitializer struct {
	storage func(m *smocks.MockCurrencyRatesStorage)
	gateway func(m *cmocks.Mockgateway)
}

func setupRater(t *testing.T, cfg config.CurrencyConfig, i raterMocksInitializer) *rater {
	ctrl := gomock.NewController(t)

	storageMock := smocks.NewMockCurrencyRatesStorage(ctrl)
	if i.storage != nil {
		i.storage(storageMock)
	}

	gatewayMock := cmocks.NewMockgateway(ctrl)
	if i.gateway != nil {
		i.gateway(gatewayMock)
	}

	return NewRater(cfg, storageMock, gatewayMock, zap.NewNop())
}

func Test_rater_Run(t *testing.T) {
	t.Run("refresh once and check lock", func(t *testing.T) {
		// ARRANGE
		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			<-time.After(time.Millisecond)
			cancel()
		}()

		var r *rater
		r = setupRater(t, test.DefaultCurrencyCfg, raterMocksInitializer{
			storage: func(m *smocks.MockCurrencyRatesStorage) {
				m.EXPECT().Add(gomock.AssignableToTypeOf(test.CtxInterface), "USD", test.Today, int64(500000)).Return(nil)
				m.EXPECT().Add(gomock.AssignableToTypeOf(test.CtxInterface), "EUR", test.Today, int64(550000)).DoAndReturn(func(_ context.Context, _ string, _ time.Time, _ int64) any {
					assert.False(t, r.TryAcquireExchange())
					return nil
				})
			},
			gateway: func(m *cmocks.Mockgateway) {
				m.EXPECT().FetchRates(gomock.AssignableToTypeOf(test.CtxInterface)).Return(map[string]int64{
					"USD": 500000,
					"EUR": 550000,
				}, test.Today, nil)
			},
		})

		// ACT & ASSERT
		assert.True(t, r.TryAcquireExchange(), "Rater not ready")
		r.ReleaseExchange()

		_ = r.Run(ctx)

		// ASSERT
		assert.True(t, r.TryAcquireExchange(), "Rater not ready after Run")
		r.ReleaseExchange()
	})

	t.Run("refresh twice with errors", func(t *testing.T) {
		// ARRANGE
		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			<-time.After(150 * time.Millisecond)
			cancel()
		}()

		r := setupRater(t, test.DefaultCurrencyCfg, raterMocksInitializer{
			storage: func(m *smocks.MockCurrencyRatesStorage) {
				m.EXPECT().Add(gomock.AssignableToTypeOf(test.CtxInterface), "USD", test.Yesterday, int64(410000)).Return(test.SimpleError)
				m.EXPECT().Add(gomock.AssignableToTypeOf(test.CtxInterface), "EUR", test.Yesterday, int64(460000)).Return(nil)
			},
			gateway: func(m *cmocks.Mockgateway) {
				m.EXPECT().FetchRates(gomock.AssignableToTypeOf(test.CtxInterface)).Return(nil, time.Time{}, test.SimpleError)
				m.EXPECT().FetchRates(gomock.AssignableToTypeOf(test.CtxInterface)).Return(map[string]int64{
					"USD": 410000,
					"EUR": 460000,
				}, test.Yesterday, nil)
			},
		})

		// ACT
		_ = r.Run(ctx)

		// ASSERT
		assert.True(t, r.ready)
	})
}

func Test_rater_Exchange(t *testing.T) {
	t.Run("equal currencies", func(t *testing.T) {
		// ARRANGE
		r := setupRater(t, test.DefaultCurrencyCfg, raterMocksInitializer{})

		// ACT
		value, err := r.Exchange(
			context.Background(),
			int64(10000), // value
			"EUR",        // from
			"EUR",        // to
			test.Today,   // date
		)

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, int64(10000), value)
	})

	t.Run("RUB to USD base storage error", func(t *testing.T) {
		// ARRANGE
		r := setupRater(t, test.DefaultCurrencyCfg, raterMocksInitializer{
			storage: func(m *smocks.MockCurrencyRatesStorage) {
				m.EXPECT().Get(gomock.AssignableToTypeOf(test.CtxInterface), "RUB", test.Today).Return(int64(0), false, test.SimpleError)
			},
		})

		// ACT
		value, err := r.Exchange(
			context.Background(),
			int64(1000000), // value
			"RUB",          // from
			"USD",          // to
			test.Today,     // date
		)

		// ASSERT
		assert.Error(t, err)
		assert.Equal(t, int64(0), value)
	})

	t.Run("RUB to USD base no rates", func(t *testing.T) {
		// ARRANGE
		r := setupRater(t, test.DefaultCurrencyCfg, raterMocksInitializer{
			storage: func(m *smocks.MockCurrencyRatesStorage) {
				m.EXPECT().Get(gomock.AssignableToTypeOf(test.CtxInterface), "RUB", test.Today).Return(int64(0), false, nil)
			},
		})

		// ACT
		value, err := r.Exchange(
			context.Background(),
			int64(1000000), // value
			"RUB",          // from
			"USD",          // to
			test.Today,     // date
		)

		// ASSERT
		assert.Error(t, err)
		assert.Equal(t, int64(0), value)
	})

	t.Run("RUB to EUR target storage error", func(t *testing.T) {
		// ARRANGE
		r := setupRater(t, test.DefaultCurrencyCfg, raterMocksInitializer{
			storage: func(m *smocks.MockCurrencyRatesStorage) {
				m.EXPECT().Get(gomock.AssignableToTypeOf(test.CtxInterface), "EUR", test.Today).Return(int64(0), false, test.SimpleError)
			},
		})

		// ACT
		value, err := r.Exchange(
			context.Background(),
			int64(1000000), // value
			"USD",          // from
			"EUR",          // to
			test.Today,     // date
		)

		// ASSERT
		assert.Error(t, err)
		assert.Equal(t, int64(0), value)
	})

	t.Run("RUB to EUR target no rates", func(t *testing.T) {
		// ARRANGE
		r := setupRater(t, test.DefaultCurrencyCfg, raterMocksInitializer{
			storage: func(m *smocks.MockCurrencyRatesStorage) {
				m.EXPECT().Get(gomock.AssignableToTypeOf(test.CtxInterface), "RUB", test.Today).Return(int64(200), true, nil)
				m.EXPECT().Get(gomock.AssignableToTypeOf(test.CtxInterface), "EUR", test.Today).Return(int64(0), false, nil)
			},
		})

		// ACT
		value, err := r.Exchange(
			context.Background(),
			int64(1000000), // value
			"RUB",          // from
			"EUR",          // to
			test.Today,     // date
		)

		// ASSERT
		assert.Error(t, err)
		assert.Equal(t, int64(0), value)
	})

	t.Run("RUB to USD success", func(t *testing.T) {
		// ARRANGE
		r := setupRater(t, test.DefaultCurrencyCfg, raterMocksInitializer{
			storage: func(m *smocks.MockCurrencyRatesStorage) {
				m.EXPECT().Get(gomock.AssignableToTypeOf(test.CtxInterface), "RUB", test.Today).Return(int64(200), true, nil)
			},
		})

		// ACT
		value, err := r.Exchange(
			context.Background(),
			int64(1500000), // value
			"RUB",          // from
			"USD",          // to
			test.Today,     // date
		)

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, int64(30000), value)
	})
}
