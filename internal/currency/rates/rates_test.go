package rates

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/config"
	rmocks "gitlab.ozon.dev/almenschhikov/go-course-4/internal/mocks/currency/rates"
	smocks "gitlab.ozon.dev/almenschhikov/go-course-4/internal/mocks/storage"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/utils"
	"go.uber.org/zap"
)

var (
	ctxInterface = reflect.TypeOf((*context.Context)(nil)).Elem()

	today       = utils.TruncateToDate(time.Now())
	yesterday   = today.Add(-24 * time.Hour)
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
		RefreshInterval: 100 * time.Millisecond,
	}
)

type mocksInitializer struct {
	storage func(m *smocks.MockCurrencyRatesStorage)
	gateway func(m *rmocks.Mockgateway)
}

func setupRater(t *testing.T, cfg config.CurrencyConfig, i mocksInitializer) *rater {
	ctrl := gomock.NewController(t)

	storageMock := smocks.NewMockCurrencyRatesStorage(ctrl)
	if i.storage != nil {
		i.storage(storageMock)
	}

	gatewayMock := rmocks.NewMockgateway(ctrl)
	if i.gateway != nil {
		i.gateway(gatewayMock)
	}

	return NewRater(cfg, storageMock, gatewayMock, zap.NewNop())
}

func Test_rater_Run(t *testing.T) {
	t.Run("refresh once", func(t *testing.T) {
		// ARRANGE
		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			<-time.After(time.Millisecond)
			cancel()
		}()

		r := setupRater(t, defaultCfg, mocksInitializer{
			storage: func(m *smocks.MockCurrencyRatesStorage) {
				m.EXPECT().Add(gomock.AssignableToTypeOf(ctxInterface), "USD", today, int64(500000)).Return(nil)
				m.EXPECT().Add(gomock.AssignableToTypeOf(ctxInterface), "EUR", today, int64(550000)).Return(nil)
			},
			gateway: func(m *rmocks.Mockgateway) {
				m.EXPECT().FetchRates(gomock.AssignableToTypeOf(ctxInterface)).Return(map[string]int64{
					"USD": 500000,
					"EUR": 550000,
				}, today, nil)
			},
		})

		// ACT
		_ = r.Run(ctx)

		// ASSERT
		assert.True(t, r.Ready(), "Rater not ready after Run")
	})

	t.Run("refresh twice with errors", func(t *testing.T) {
		// ARRANGE
		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			<-time.After(150 * time.Millisecond)
			cancel()
		}()

		r := setupRater(t, defaultCfg, mocksInitializer{
			storage: func(m *smocks.MockCurrencyRatesStorage) {
				m.EXPECT().Add(gomock.AssignableToTypeOf(ctxInterface), "USD", yesterday, int64(410000)).Return(simpleError)
				m.EXPECT().Add(gomock.AssignableToTypeOf(ctxInterface), "EUR", yesterday, int64(460000)).Return(nil)
			},
			gateway: func(m *rmocks.Mockgateway) {
				m.EXPECT().FetchRates(gomock.AssignableToTypeOf(ctxInterface)).Return(nil, time.Time{}, simpleError)
				m.EXPECT().FetchRates(gomock.AssignableToTypeOf(ctxInterface)).Return(map[string]int64{
					"USD": 410000,
					"EUR": 460000,
				}, yesterday, nil)
			},
		})

		// ACT
		_ = r.Run(ctx)

		// ASSERT
		assert.True(t, r.Ready())
	})
}

func Test_rater_Ready(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		// ARRANGE
		r := setupRater(t, defaultCfg, mocksInitializer{})
		r.ready = true

		// NO ACT

		// ASSERT
		assert.True(t, r.Ready())
	})

	t.Run("false", func(t *testing.T) {
		// ARRANGE
		r := setupRater(t, defaultCfg, mocksInitializer{})
		r.ready = false

		// NO ACT

		// ASSERT
		assert.False(t, r.Ready())
	})
}

func Test_rater_Exchange(t *testing.T) {
	t.Run("equal currencies", func(t *testing.T) {
		// ARRANGE
		r := setupRater(t, defaultCfg, mocksInitializer{})

		// ACT
		value, err := r.Exchange(
			context.Background(),
			int64(10000), // value
			"EUR",        // from
			"EUR",        // to
			today,        // date
		)

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, int64(10000), value)
	})

	t.Run("RUB to USD base error", func(t *testing.T) {
		// ARRANGE
		r := setupRater(t, defaultCfg, mocksInitializer{
			storage: func(m *smocks.MockCurrencyRatesStorage) {
				m.EXPECT().Get(gomock.AssignableToTypeOf(ctxInterface), "RUB", today).Return(int64(0), false, simpleError)
			},
		})

		// ACT
		value, err := r.Exchange(
			context.Background(),
			int64(1000000), // value
			"RUB",          // from
			"USD",          // to
			today,          // date
		)

		// ASSERT
		assert.Error(t, err)
		assert.Equal(t, int64(0), value)
	})

	t.Run("RUB to USD base no rates", func(t *testing.T) {
		// ARRANGE
		r := setupRater(t, defaultCfg, mocksInitializer{
			storage: func(m *smocks.MockCurrencyRatesStorage) {
				m.EXPECT().Get(gomock.AssignableToTypeOf(ctxInterface), "RUB", today).Return(int64(0), false, nil)
			},
		})

		// ACT
		value, err := r.Exchange(
			context.Background(),
			int64(1000000), // value
			"RUB",          // from
			"USD",          // to
			today,          // date
		)

		// ASSERT
		assert.Error(t, err)
		assert.Equal(t, int64(0), value)
	})

	t.Run("RUB to EUR target error", func(t *testing.T) {
		// ARRANGE
		r := setupRater(t, defaultCfg, mocksInitializer{
			storage: func(m *smocks.MockCurrencyRatesStorage) {
				m.EXPECT().Get(gomock.AssignableToTypeOf(ctxInterface), "RUB", today).Return(int64(200), true, nil)
				m.EXPECT().Get(gomock.AssignableToTypeOf(ctxInterface), "EUR", today).Return(int64(0), false, simpleError)
			},
		})

		// ACT
		value, err := r.Exchange(
			context.Background(),
			int64(1000000), // value
			"RUB",          // from
			"EUR",          // to
			today,          // date
		)

		// ASSERT
		assert.Error(t, err)
		assert.Equal(t, int64(0), value)
	})

	t.Run("RUB to EUR target no rates", func(t *testing.T) {
		// ARRANGE
		r := setupRater(t, defaultCfg, mocksInitializer{
			storage: func(m *smocks.MockCurrencyRatesStorage) {
				m.EXPECT().Get(gomock.AssignableToTypeOf(ctxInterface), "RUB", today).Return(int64(200), true, nil)
				m.EXPECT().Get(gomock.AssignableToTypeOf(ctxInterface), "EUR", today).Return(int64(0), false, nil)
			},
		})

		// ACT
		value, err := r.Exchange(
			context.Background(),
			int64(1000000), // value
			"RUB",          // from
			"EUR",          // to
			today,          // date
		)

		// ASSERT
		assert.Error(t, err)
		assert.Equal(t, int64(0), value)
	})

	t.Run("RUB to USD success", func(t *testing.T) {
		// ARRANGE
		r := setupRater(t, defaultCfg, mocksInitializer{
			storage: func(m *smocks.MockCurrencyRatesStorage) {
				m.EXPECT().Get(gomock.AssignableToTypeOf(ctxInterface), "RUB", today).Return(int64(200), true, nil)
			},
		})

		// ACT
		value, err := r.Exchange(
			context.Background(),
			int64(1500000), // value
			"RUB",          // from
			"USD",          // to
			today,          // date
		)

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, int64(30000), value)
	})
}
