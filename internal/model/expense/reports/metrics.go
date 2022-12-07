package reports

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"google.golang.org/grpc"
)

var (
	producedCount = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "finassis",
			Subsystem: "reporter",
			Name:      "produced_messages_total",
			Help:      "FinAssist report messages produced.",
		},
		[]string{
			"currency",
		},
	)

	consumedCount = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "finassis",
			Subsystem: "reporter",
			Name:      "consumed_messages_total",
			Help:      "FinAssist report messages consumed.",
		},
		[]string{
			"result",
		},
	)

	consumeTime = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "finassis",
			Subsystem: "reporter",
			Name:      "consume_time_seconds",
			Help:      "FinAssist reports consume time.",
			Buckets:   []float64{0.0005, 0.001, 0.005, 0.015, 0.05, 0.1, 0.5, 1},
		},
		[]string{
			"method",
		},
	)
)

func metricsInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()
	defer consumeTime.WithLabelValues(
		info.FullMethod, // method
	).Observe(time.Since(start).Seconds())

	return handler(ctx, req)
}
