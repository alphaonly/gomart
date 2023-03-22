package metricvaluei

import (
	"encoding/json"
	"errors"
	"strconv"
)

type Gauge float64
type Counter int64

type MetricValue interface {
	SetValue(MetricValue)
	GetValue() MetricValue         //Gauge or Counter
	GetInternalValue() interface{} //float64 or int64
	GetString() string
	AddValue(MetricValue) MetricValue // increment  current and return incremented( for counter only)

	MarshalJSON() ([]byte, error)
	UnmarshalJSON(b []byte) error
}

type GaugeValueJSON struct {
	Value float64 `json:"value"`
}

type GaugeValue struct {
	valueFloat float64
}

func NewFloat(g float64) *GaugeValue {
	return &GaugeValue{

		valueFloat: g,
	}
}
func (v *GaugeValue) New(g Gauge) *GaugeValue {
	v.valueFloat = float64(g)
	return v
}
func (v *GaugeValue) SetValue(gauge MetricValue) {
	v.valueFloat = gauge.(*GaugeValue).valueFloat
}
func (v *GaugeValue) GetValue() MetricValue {
	return v
}
func (v *GaugeValue) GetInternalValue() interface{} {
	return v.valueFloat
}
func (v *GaugeValue) GetString() string {
	return strconv.FormatFloat(v.valueFloat, 'f', -1, 64)
}
func (v *GaugeValue) AddValue(v1 MetricValue) MetricValue {
	return v //Mocked, as it's needed for counter only
}
func (v *GaugeValue) MarshalJSON() ([]byte, error) {
	cj := GaugeValueJSON{Value: v.valueFloat}

	return json.Marshal(cj)
}

func (v *GaugeValue) UnmarshalJSON(data []byte) error {
	cj := &GaugeValueJSON{}

	err := json.Unmarshal(data, cj)
	if err != nil {
		return errors.New("CounterValue unmarshal error")
	}
	v.valueFloat = cj.Value

	return nil

}

// check
//var m interfaces.MetricValue = &GaugeValue{}

type CounterValueJSON struct {
	Value string `json:"value"`
}

type CounterValue struct {
	valueInt int64
}

func NewCounterValue() MetricValue {
	return MetricValue(&CounterValue{})
}

func (v *CounterValue) New(c Counter) *CounterValue {
	v.valueInt = int64(c)
	return v
}
func NewInt(c int64) *CounterValue {
	return &CounterValue{
		valueInt: c,
	}
}

func (v *CounterValue) SetValue(counter MetricValue) {
	v.valueInt = counter.(*CounterValue).valueInt
}
func (v *CounterValue) GetValue() MetricValue {
	return v
}
func (v *CounterValue) GetInternalValue() interface{} {
	return v.valueInt
}
func (v *CounterValue) GetString() string {
	return strconv.FormatUint(uint64(v.valueInt), 10)
}
func (v *CounterValue) AddValue(v1 MetricValue) MetricValue {
	sumVal := v1.(*CounterValue).valueInt + v.valueInt
	return &CounterValue{valueInt: sumVal}
}

func (v *CounterValue) MarshalJSON() ([]byte, error) {
	cj := CounterValueJSON{Value: v.GetString()}

	return json.Marshal(cj)
}

func (v *CounterValue) UnmarshalJSON(data []byte) error {
	cj := &CounterValueJSON{}

	err := json.Unmarshal(data, cj)
	if err != nil {
		return errors.New("CounterValue unmarshal error")
	}
	v.valueInt, err = strconv.ParseInt(cj.Value, 10, 64)
	if err != nil {
		return errors.New("unmarshalled CounterValue parse error")
	}

	return nil

}

// check
//  m:= MetricValue(&CounterValue{})
//  g:= MetricValue(&GaugeValue{})
//
