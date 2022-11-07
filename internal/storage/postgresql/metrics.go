package postgresql

import (
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
)

type exporter struct {
	db *pgxpool.Pool

	maxOpenConnections *prometheus.Desc

	idleConnections     *prometheus.Desc
	acquiredConnections *prometheus.Desc

	acquireDuration *prometheus.Desc
}

func newExporter(db *pgxpool.Pool, dbName string) *exporter {
	fqName := func(name string) string { return prometheus.BuildFQName("finassist", "database", name) }
	labels := prometheus.Labels{"db_name": dbName}

	return &exporter{
		db: db,
		maxOpenConnections: prometheus.NewDesc(
			fqName("max_open_connections"),
			"Maximum number of open connections to the database.",
			nil,
			labels,
		),
		idleConnections: prometheus.NewDesc(
			fqName("idle_connections"),
			"The number of idle connections.",
			nil,
			labels,
		),
		acquiredConnections: prometheus.NewDesc(
			fqName("acquired_connections"),
			"The number of acquired connections.",
			nil,
			labels,
		),
		acquireDuration: prometheus.NewDesc(
			fqName("wait_duration_seconds"),
			"Waiting duration",
			nil,
			labels,
		),
	}
}

func (e *exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.maxOpenConnections
	ch <- e.idleConnections
	ch <- e.acquiredConnections
	ch <- e.acquireDuration
}

func (e *exporter) Collect(ch chan<- prometheus.Metric) {
	ch <- prometheus.MustNewConstMetric(
		e.maxOpenConnections,
		prometheus.GaugeValue,
		float64(e.db.Stat().MaxConns()),
	)
	ch <- prometheus.MustNewConstMetric(
		e.idleConnections,
		prometheus.GaugeValue,
		float64(e.db.Stat().IdleConns()),
	)
	ch <- prometheus.MustNewConstMetric(
		e.acquiredConnections,
		prometheus.CounterValue,
		float64(e.db.Stat().AcquiredConns()),
	)
	ch <- prometheus.MustNewConstMetric(
		e.acquireDuration,
		prometheus.CounterValue,
		e.db.Stat().AcquireDuration().Seconds(),
	)
}
