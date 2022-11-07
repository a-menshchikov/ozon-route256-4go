package telegram

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	_commandResponseTime = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "finassist",
			Subsystem: "telegram",
			Name:      "response_time_seconds",
			Help:      "Telegram Bot response time.",
			Buckets:   []float64{0.001, 0.005, 0.015, 0.05, 0.1, 0.5, 1, 2, 5},
		},
		[]string{
			"command",
		},
	)

	_commandCount = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "finassis",
			Subsystem: "telegram",
			Name:      "commands_total",
			Help:      "Telegram Bot commands handled.",
		},
		[]string{
			"command",
		},
	)
)
