package storage

import (
	"context"

	metricsJSON "github.com/alphaonly/gomart/internal/server/metricsJSON"
	metricValueInt "github.com/alphaonly/gomart/internal/server/metricvaluei"
)

type Gauge float64
type Counter int64
type Storage interface {
	GetMetric(ctx context.Context, name string, MType string) (mv metricValueInt.MetricValue, err error)
	SaveMetric(ctx context.Context, name string, mv *metricValueInt.MetricValue) (err error)
	GetAllMetrics(ctx context.Context) (mvList *metricsJSON.MetricsMapType, err error)
	SaveAllMetrics(ctx context.Context, mvList *metricsJSON.MetricsMapType) (err error)
}
