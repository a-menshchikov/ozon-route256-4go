package expense

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	mmocks "gitlab.ozon.dev/almenschhikov/go-course-4/internal/mocks/model"
	smocks "gitlab.ozon.dev/almenschhikov/go-course-4/internal/mocks/storage"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/model"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/test"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/types"
	"go.uber.org/zap"
)

type reporterMocksInitializer struct {
	storage func(m *smocks.MockExpenseStorage)
	rater   func(m *mmocks.MockRater)
}

func setupReporter(t *testing.T, i reporterMocksInitializer) *reporter {
	ctrl := gomock.NewController(t)

	storageMock := smocks.NewMockExpenseStorage(ctrl)
	if i.storage != nil {
		i.storage(storageMock)
	}

	raterMock := mmocks.NewMockRater(ctrl)
	if i.rater != nil {
		i.rater(raterMock)
	}

	return NewReporter(storageMock, raterMock, zap.NewNop())
}

func Test_reporter_GetReport(t *testing.T) {
	t.Run("not ready", func(t *testing.T) {
		// ARRANGE
		r := setupReporter(t, reporterMocksInitializer{
			rater: func(m *mmocks.MockRater) {
				m.EXPECT().TryAcquireExchange().Return(false)
			},
		})

		// ACT
		data, err := r.GetReport(context.Background(), test.User, test.Today, "RUB")

		// ASSERT
		assert.Equal(t, model.ErrNotReady, err)
		assert.Empty(t, data)
	})

	t.Run("storage error", func(t *testing.T) {
		// ARRANGE
		r := setupReporter(t, reporterMocksInitializer{
			storage: func(m *smocks.MockExpenseStorage) {
				m.EXPECT().List(gomock.AssignableToTypeOf(test.CtxInterface), test.User, test.Today).Return(nil, test.SimpleError)
			},
			rater: func(m *mmocks.MockRater) {
				m.EXPECT().TryAcquireExchange().Return(true)
				m.EXPECT().ReleaseExchange()
			},
		})

		// ACT
		data, err := r.GetReport(context.Background(), test.User, test.Today, "RUB")

		// ASSERT
		assert.Error(t, err)
		assert.Empty(t, data)
	})

	t.Run("rater error", func(t *testing.T) {
		// ARRANGE
		r := setupReporter(t, reporterMocksInitializer{
			storage: func(m *smocks.MockExpenseStorage) {
				m.EXPECT().List(gomock.AssignableToTypeOf(test.CtxInterface), test.User, test.Today).Return(map[string][]types.ExpenseItem{
					"coffee": {
						{
							Date:     test.Today,
							Amount:   20000,
							Currency: "USD",
						},
						{
							Date:     test.Today,
							Amount:   30000,
							Currency: "EUR",
						},
					},
				}, nil)
			},
			rater: func(m *mmocks.MockRater) {
				m.EXPECT().TryAcquireExchange().Return(true)
				m.EXPECT().ReleaseExchange()
				m.EXPECT().Exchange(gomock.AssignableToTypeOf(test.CtxInterface), int64(20000), "USD", "RUB", test.Today).Return(int64(1000000), nil)
				m.EXPECT().Exchange(gomock.AssignableToTypeOf(test.CtxInterface), int64(30000), "EUR", "RUB", test.Today).Return(int64(0), test.SimpleError)
			},
		})

		// ACT
		data, err := r.GetReport(context.Background(), test.User, test.Today, "RUB")

		// ASSERT
		assert.Error(t, err)
		assert.Empty(t, data)
	})

	t.Run("success", func(t *testing.T) {
		// ARRANGE
		r := setupReporter(t, reporterMocksInitializer{
			storage: func(m *smocks.MockExpenseStorage) {
				m.EXPECT().List(gomock.AssignableToTypeOf(test.CtxInterface), test.User, test.Today).Return(map[string][]types.ExpenseItem{
					"taxi": {
						{
							Date:     test.Today,
							Amount:   100000,
							Currency: "USD",
						},
						{
							Date:     test.Today,
							Amount:   120000,
							Currency: "EUR",
						},
					},
					"coffee": {
						{
							Date:     test.Today,
							Amount:   1200000,
							Currency: "RUB",
						},
					},
				}, nil)
			},
			rater: func(m *mmocks.MockRater) {
				m.EXPECT().TryAcquireExchange().Return(true)
				m.EXPECT().ReleaseExchange()
				m.EXPECT().Exchange(gomock.AssignableToTypeOf(test.CtxInterface), int64(100000), "USD", "RUB", test.Today).Return(int64(5000000), nil)
				m.EXPECT().Exchange(gomock.AssignableToTypeOf(test.CtxInterface), int64(120000), "EUR", "RUB", test.Today).Return(int64(6600000), nil)
				m.EXPECT().Exchange(gomock.AssignableToTypeOf(test.CtxInterface), int64(1200000), "RUB", "RUB", test.Today).Return(int64(1200000), nil)
			},
		})

		// ACT
		data, err := r.GetReport(context.Background(), test.User, test.Today, "RUB")

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, map[string]int64{
			"taxi":   11600000,
			"coffee": 1200000,
		}, data)
	})
}
