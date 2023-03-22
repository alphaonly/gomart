package metricsjson

import (
	"encoding/json"
	"errors"

	"github.com/alphaonly/gomart/internal/schema"
	mVal "github.com/alphaonly/gomart/internal/server/metricvaluei"
)

type MetricsMapType map[string]mVal.MetricValue

func (m MetricsMapType) MarshalJSON() ([]byte, error) {
	mjArray := make([]schema.MetricsJSON, len(m))
	i := 0
	for k, v := range m {
		switch value := v.(type) {
		case *mVal.GaugeValue:
			{
				v := value.GetInternalValue().(float64)
				mjArray[i] = schema.MetricsJSON{
					ID:    k,
					MType: "gauge",
					Value: &v,
				}
			}
		case *mVal.CounterValue:
			{
				v := value.GetInternalValue().(int64)
				mjArray[i] = schema.MetricsJSON{
					ID:    k,
					MType: "counter",
					Delta: &v,
				}
			}
		default:
			return nil, errors.New("undefined type in type switch metricValue")
		}
		i++
	}
	mBytes, err := json.Marshal(&mjArray)
	if err != nil {
		return nil, err
	}

	return mBytes, nil
}

func (m MetricsMapType) UnmarshalJSON(b []byte) error {
	var mjArray []schema.MetricsJSON
	if err := json.Unmarshal(b, &mjArray); err != nil {
		return err
	}
	for _, v := range mjArray {

		switch v.MType {
		case "gauge":
			if v.Value == nil {
				return errors.New("v.Value null pointer")
			}
			m[v.ID] = mVal.NewFloat(*v.Value)
		case "counter":
			if v.Delta == nil {
				return errors.New("v.delta null pointer")
			}
			m[v.ID] = mVal.NewInt(*v.Delta)
		default:
			return errors.New("unknown type in decoding metricsJSONArray")
		}
	}

	return nil
}
