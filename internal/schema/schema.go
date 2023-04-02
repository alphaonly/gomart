package schema

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

type PreviousBytes []byte
type CtxUName string
type ContextKey int

const PKey1 ContextKey = 123455
const CtxKeyUName ContextKey = 1343456

type orderType map[string]int64

var OrderStatus = orderType{
	"NEW":        1,
	"PROCESSING": 2,
	"INVALID":    3,
	"PROCESSED":  4,
}

type User struct {
	User       string  `json:"login"`
	Password   string  `json:"password"`
	Accrual    float64 `json:"current,omitempty"`
	Withdrawal float64 `json:"withdrawn,omitempty"`
}

func (u User) CheckIdentity(u2 *User) (ok bool) {
	if u.User == u2.User && u.Password == u2.Password {
		return true
	}
	return false
}

type CreatedTime time.Time

func (t CreatedTime) MarshalJSON() ([]byte, error) {
	value := time.Time(t)
	//created, err := time.Parse(time.RFC3339, d.created_at.String)

	bytes, err := json.Marshal(value.Format(time.RFC3339))
	if err != nil {
		return nil, fmt.Errorf("error marshal  CreatedTime %v", value)
	}
	return bytes, nil
}
func (t *CreatedTime) UnmarshalJSON(b []byte) error {
	var createdTimeString string
	err := json.Unmarshal(b, &createdTimeString)
	if err != nil {
		return fmt.Errorf("error unmarshal  CreatedTime %v", b)
	}
	createdTime, err := time.Parse(time.RFC3339, createdTimeString)
	if err != nil {
		return fmt.Errorf("error parse  CreatedTime to RFC3339 %v", b)
	}
	*t = CreatedTime(createdTime)
	return nil
}

type Order struct {
	Order   int64       `json:"number"`
	User    string      `json:"user"`
	Status  int64       `json:"status,omitempty"`
	Accrual float64     `json:"accrual,omitempty"`
	Created CreatedTime `json:"uploaded_at"`
}

type Withdrawal struct {
	User       string      `json:"user"`
	Processed  CreatedTime `json:"processed_at"`
	Withdrawal float64     `json:"sum,omitempty"`
}

type ByTimeDescending Withdrawals

func (a ByTimeDescending) Len() int      { return len(a) }
func (a ByTimeDescending) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByTimeDescending) Less(i, j int) bool {
	return time.Time(a[i].Processed).Before(time.Time(a[i].Processed))
}

type Withdrawals []Withdrawal

type Orders map[int64]Order

func (o Orders) MarshalJSON() ([]byte, error) {

	oArray := make([]Order, len(o))
	i := 0
	for k, v := range o {
		oArray[i] = Order{
			Order:   k,
			Status:  v.Status,
			Accrual: v.Accrual,
			Created: v.Created,
		}
		i++
	}
	bytes, err := json.Marshal(&oArray)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}
func (o Orders) UnmarshalJSON(b []byte) error {
	var oArray []Order
	if err := json.Unmarshal(b, &oArray); err != nil {
		return err
	}
	for _, v := range oArray {
		o[v.Order] = v
	}
	return nil
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
