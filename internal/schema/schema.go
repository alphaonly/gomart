package schema

import (
	"encoding/json"
	"errors"
	"time"
)

type PreviousBytes []byte
type ContextKey int

const PKey1 ContextKey = 123455

type duplicateChecker interface{
	check(make(map[string]MetricsJSON))	
}

type MetricsJSONSlice []MetricsJSON
func (s *MetricsJSONSlice)check(m map[string]MetricsJSON){

}

// Если двойные записи в counter - суммируем в одну, gauge - оставляем последнюю
func (s *MetricsJSONSlice) EnhancedDistinct() error {
	m := make(map[string]MetricsJSON)
	for _, e := range *s {
		
		if e.MType == "counter"{
			c, exists := m[e.ID]
			if exists {
				sum := int64(*e.Delta + *c.Delta)
				m[e.ID] = MetricsJSON{e.ID, e.MType, &sum, e.Value, ""}
				continue
			}
		}
		m[e.ID] = e
	}
	*s = MetricsJSONSlice{}
	for _, v := range m {
		*s = append(*s, v)
	}
	return nil
}

type MetricsJSON struct {
	ID    string   `json:"id"`              // имя метрикИ
	MType string   `json:"type"`            // параметр, пID    string   `json:"id"`              // имя метрикиринимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
	Hash  string   `json:"hash,omitempty"`  // значение хеш-функции
}

func (m MetricsJSON) Equals(m2 MetricsJSON) bool {
	return m.ID == m2.ID
}

type Duration time.Duration

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(d).String())
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch value := v.(type) {
	case float64:
		*d = Duration(time.Duration(value))
		return nil
	case string:
		tmp, err := time.ParseDuration(value)
		if err != nil {
			return err
		}
		*d = Duration(tmp)
		return nil
	default:
		return errors.New("invalid duration")
	}
}
