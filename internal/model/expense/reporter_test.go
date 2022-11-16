//go:build unit

package expense

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	mocks "gitlab.ozon.dev/almenschhikov/go-course-4/internal/mocks/model/expense"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/model"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/test"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/types"
	"go.uber.org/zap"
)

type reporterMocksInitializer struct {
	producer func(m *mocks.Mockproducer)
	listener func(m *mocks.Mocklistener)
}

func setupReporter(t *testing.T, timeout time.Duration, i reporterMocksInitializer) *reporter {
	ctrl := gomock.NewController(t)

	producerMock := mocks.NewMockproducer(ctrl)
	if i.producer != nil {
		i.producer(producerMock)
	}

	listenerMock := mocks.NewMocklistener(ctrl)
	if i.listener != nil {
		i.listener(listenerMock)
	}

	return NewReporter(timeout, producerMock, listenerMock, zap.NewNop())
}

func Test_reporter_GetReport(t *testing.T) {
	t.Run("producer error", func(t *testing.T) {
		// ARRANGE
		r := setupReporter(t, 0, reporterMocksInitializer{
			producer: func(m *mocks.Mockproducer) {
				m.EXPECT().Send(gomock.AssignableToTypeOf(test.CtxInterface), test.User, test.Yesterday, "RUB").Return(test.SimpleError)
			},
			listener: func(m *mocks.Mocklistener) {
				m.EXPECT().Subscribe(test.User).Return(make(<-chan types.Report))
				m.EXPECT().Unsubscribe(test.User)
			},
		})

		// ACT
		data, err := r.GetReport(context.Background(), test.User, test.Yesterday, "RUB")

		// ASSERT
		assert.Error(t, err)
		assert.Empty(t, data)
	})

	t.Run("timeout", func(t *testing.T) {
		// ARRANGE
		r := setupReporter(t, time.Millisecond, reporterMocksInitializer{
			producer: func(m *mocks.Mockproducer) {
				m.EXPECT().Send(gomock.AssignableToTypeOf(test.CtxInterface), test.User, test.Yesterday, "USD").Return(nil)
			},
			listener: func(m *mocks.Mocklistener) {
				m.EXPECT().Subscribe(test.User).Return(make(<-chan types.Report))
				m.EXPECT().Unsubscribe(test.User)
			},
		})

		// ACT
		data, err := r.GetReport(context.Background(), test.User, test.Yesterday, "USD")

		// ASSERT
		assert.Error(t, err)
		assert.Equal(t, "context deadline exceeded", err.Error())
		assert.Empty(t, data)
	})

	t.Run("not ready", func(t *testing.T) {
		// ARRANGE
		r := setupReporter(t, time.Second, reporterMocksInitializer{
			producer: func(m *mocks.Mockproducer) {
				m.EXPECT().Send(gomock.AssignableToTypeOf(test.CtxInterface), test.User, test.Today, "EUR").Return(nil)
			},
			listener: func(m *mocks.Mocklistener) {
				reportCh := make(chan types.Report)
				go func() {
					reportCh <- types.Report{
						Success: false,
						Error:   "not ready",
					}
				}()
				m.EXPECT().Subscribe(test.User).Return(func(ch <-chan types.Report) <-chan types.Report { return ch }(reportCh))
				m.EXPECT().Unsubscribe(test.User)
			},
		})

		// ACT
		data, err := r.GetReport(context.Background(), test.User, test.Today, "EUR")

		// ASSERT
		assert.Equal(t, model.ErrNotReady, err)
		assert.Empty(t, data)
	})

	t.Run("error", func(t *testing.T) {
		// ARRANGE
		r := setupReporter(t, time.Second, reporterMocksInitializer{
			producer: func(m *mocks.Mockproducer) {
				m.EXPECT().Send(gomock.AssignableToTypeOf(test.CtxInterface), test.User, test.Today, "RUB").Return(nil)
			},
			listener: func(m *mocks.Mocklistener) {
				reportCh := make(chan types.Report)
				go func() {
					reportCh <- types.Report{
						Success: false,
						Error:   "general error",
					}
				}()
				m.EXPECT().Subscribe(test.User).Return(func(ch <-chan types.Report) <-chan types.Report { return ch }(reportCh))
				m.EXPECT().Unsubscribe(test.User)
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
		r := setupReporter(t, time.Second, reporterMocksInitializer{
			producer: func(m *mocks.Mockproducer) {
				m.EXPECT().Send(gomock.AssignableToTypeOf(test.CtxInterface), test.User, test.Yesterday, "USD").Return(nil)
			},
			listener: func(m *mocks.Mocklistener) {
				reportCh := make(chan types.Report)
				go func() {
					reportCh <- types.Report{
						Data: map[string]int64{
							"taxi":   220000,
							"coffee": 250000,
						},
						Success: true,
					}
				}()
				m.EXPECT().Subscribe(test.User).Return(func(ch <-chan types.Report) <-chan types.Report { return ch }(reportCh))
				m.EXPECT().Unsubscribe(test.User)
			},
		})

		// ACT
		data, err := r.GetReport(context.Background(), test.User, test.Yesterday, "USD")

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, map[string]int64{
			"taxi":   220000,
			"coffee": 250000,
		}, data)
	})
}
