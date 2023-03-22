package mapstorage

import (
	"context"
	"errors"
	"sync"

	metricsJSON "github.com/alphaonly/gomart/internal/server/metricsJSON"
	metricvalueI "github.com/alphaonly/gomart/internal/server/metricvaluei"
	stor "github.com/alphaonly/gomart/internal/server/storage/interfaces"
)

type MapStorage struct {
	mutex      *sync.Mutex
	metricsMap *metricsJSON.MetricsMapType
}

func NewStorage() (sr stor.Storage) {
	map_ := make(metricsJSON.MetricsMapType)
	mapStorage := stor.Storage(MapStorage{
		mutex:      &sync.Mutex{},
		metricsMap: &map_,
	})
	return mapStorage
}

func New() (storage *MapStorage) {
	map_ := make(metricsJSON.MetricsMapType)
	mapStorage := MapStorage{
		mutex:      &sync.Mutex{},
		metricsMap: &map_,
	}
	return &mapStorage
}

//Имплементация интерфейса storage

// type Storage interface {
// 	GetMetric(ctx context.Context, name string) (mv *interfaces.MetricValue, err error)
// 	SaveMetric(ctx context.Context, name string, mv *interfaces.MetricValue) (err error)
// 	GetAllMetrics(ctx context.Context) (mvList *map[string]interfaces.MetricValue, err error)
// }

func (m MapStorage) GetMetric(ctx context.Context, name string, MType string) (mv metricvalueI.MetricValue, err error) {
	_map := *m.metricsMap
	if m.metricsMap == nil || len(_map) == 0 {
		return nil, errors.New("404 - not found")
	}
	if _map[name] == nil {
		return nil, errors.New("404 - not found")
	}
	return _map[name], nil
}
func (m MapStorage) SaveMetric(ctx context.Context, name string, mv *metricvalueI.MetricValue) (r error) {

	(*m.metricsMap)[name] = *mv

	return nil
}

func (m MapStorage) GetAllMetrics(ctx context.Context) (mvList *metricsJSON.MetricsMapType, err error) {

	if m.metricsMap == nil {
		return nil, errors.New("map was not initialized")
	}

	return m.metricsMap, nil

}

func (m MapStorage) SaveAllMetrics(ctx context.Context, mvList *metricsJSON.MetricsMapType) (err error) {
	(*m.metricsMap) = *mvList

	return nil
}
