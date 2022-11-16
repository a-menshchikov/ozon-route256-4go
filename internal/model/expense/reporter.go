package expense

import (
	"context"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/model"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/types"
	"go.uber.org/zap"
)

type (
	producer interface {
		Send(ctx context.Context, user *types.User, from time.Time, currency string) error
	}

	listener interface {
		Subscribe(user *types.User) <-chan types.Report
		Unsubscribe(user *types.User)
	}
)

type reporter struct {
	timeout  time.Duration
	producer producer
	listener listener
	logger   *zap.Logger
}

func NewReporter(timeout time.Duration, p producer, ls listener, l *zap.Logger) *reporter {
	return &reporter{
		timeout:  timeout,
		producer: p,
		listener: ls,
		logger:   l,
	}
}

func (r *reporter) GetReport(ctx context.Context, user *types.User, from time.Time, currency string) (map[string]int64, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "reporter.GetReport", opentracing.Tags{
		"user":     *user,
		"from":     from,
		"currency": currency,
	})
	defer span.Finish()

	reportCh := r.listener.Subscribe(user)
	defer r.listener.Unsubscribe(user)

	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	if err := r.producer.Send(ctx, user, from, currency); err != nil {
		return nil, err
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case report := <-reportCh:
		if report.Success {
			return report.Data, nil
		}

		if report.Error == model.ErrNotReady.Error() {
			return nil, model.ErrNotReady
		}

		return nil, errors.New(report.Error)
	}
}
