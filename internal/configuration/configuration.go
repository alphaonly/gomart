package configuration

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/alphaonly/gomart/internal/schema"
)

const ServerDefaultJSON = `{
"RUN_ADDRESS":"localhost:8080",
"DATABASE_URI": "postgres://postgres:mypassword@localhost:5432/yandex",
"ACCRUAL_SYSTEM_ADDRESS":"localhost:8080",
"RESTORE":true,"KEY":"",
"ACCRUAL_TIME":200
}`

type ServerConfiguration struct {
	RunAddress           string `json:"RUN_ADDRESS,omitempty"`
	Port                 string `json:"PORT,omitempty"`
	DatabaseURI          string `json:"DATABASE_URI,omitempty"`
	AccrualSystemAddress string `json:"ACCRUAL_SYSTEM_ADDRESS,omitempty"`
	AccrualTime          int64  `json:"ACCRUAL_TIME,omitempty"`
	EnvChanged           map[string]bool
}

type ServerConfigurationOption func(*ServerConfiguration)

func UnMarshalServerDefaults(s string) ServerConfiguration {
	sc := ServerConfiguration{}
	err := json.Unmarshal([]byte(s), &sc)
	if err != nil {
		log.Fatal("cannot unmarshal server configuration")
	}
	return sc

}

func NewServerConfiguration() *ServerConfiguration {
	c := UnMarshalServerDefaults(ServerDefaultJSON)
	c.Port = ":" + strings.Split(c.RunAddress, ":")[1]
	c.EnvChanged = make(map[string]bool)
	return &c

}

func NewServerConf(options ...ServerConfigurationOption) *ServerConfiguration {
	c := UnMarshalServerDefaults(ServerDefaultJSON)
	c.EnvChanged = make(map[string]bool)
	for _, option := range options {
		option(&c)
	}
	return &c
}

func UpdateSCFromEnvironment(c *ServerConfiguration) {
	c.RunAddress = getEnv("RUN_ADDRESS", &StrValue{c.RunAddress}, c.EnvChanged).(string)
	c.AccrualSystemAddress = getEnv("ACCRUAL_SYSTEM_ADDRESS", &StrValue{c.AccrualSystemAddress}, c.EnvChanged).(string)
	//PORT is derived from ADDRESS
	c.Port = ":" + strings.Split(c.RunAddress, ":")[1]
	c.DatabaseURI = getEnv("DATABASE_URI", &StrValue{c.DatabaseURI}, c.EnvChanged).(string)
}

func UpdateSCFromFlags(c *ServerConfiguration) {

	dc := NewServerConfiguration()

	var (
		a = flag.String("a", dc.RunAddress, "Domain name and :port")
		r = flag.String("r", dc.AccrualSystemAddress, "Restore from external storage:true/false")
		d = flag.String("d", dc.DatabaseURI, "database destination string")
	)
	flag.Parse()

	message := "variable %v  updated from flags, value %v"
	//Если значение из переменных равно значению по умолчанию, тогда берем из flagS
	if !c.EnvChanged["ADDRESS"] {
		c.RunAddress = *a
		c.Port = ":" + strings.Split(c.RunAddress, ":")[1]
		log.Printf(message, "ADDRESS", c.RunAddress)
		log.Printf(message, "PORT", c.Port)
	}
	if !c.EnvChanged["ACCRUAL_SYSTEM_ADDRESS"] {
		c.AccrualSystemAddress = *r
		log.Printf(message, "ACCRUAL_SYSTEM_ADDRESS", c.AccrualSystemAddress)
	}
	if !c.EnvChanged["DATABASE_URI"] {
		c.DatabaseURI = *d
		log.Printf(message, "DATABASE_URI", c.DatabaseURI)
	}
}

type VariableValue interface {
	Get() interface{}
	Set(string)
}
type StrValue struct {
	value string
}

func (v *StrValue) Get() interface{} {
	return v.value
}
func NewStrValue(s string) VariableValue {
	return &StrValue{value: s}
}
func (v *StrValue) Set(s string) {
	v.value = s
}

type IntValue struct {
	value int
}

func (v IntValue) Get() interface{} {
	return v.value
}
func (v *IntValue) Set(s string) {
	var err error
	v.value, err = strconv.Atoi(s)
	if err != nil {
		log.Fatal("Int Parse error")
	}
}

func NewIntValue(s string) VariableValue {
	changedValue, err := strconv.Atoi(s)
	if err != nil {
		log.Fatal("Int64 Parse error")
	}
	return &IntValue{value: changedValue}
}

type BoolValue struct {
	value bool
}

func (v BoolValue) Get() interface{} {
	return v.value
}
func (v *BoolValue) Set(s string) {
	var err error
	v.value, err = strconv.ParseBool(s)
	if err != nil {
		log.Fatal("Bool Parse error")
	}
}
func NewBoolValue(s string) VariableValue {
	changedValue, err := strconv.ParseBool(s)
	if err != nil {
		log.Fatal("Bool Parse error")
	}
	return &BoolValue{value: changedValue}
}

type DurValue struct {
	value schema.Duration
}

func (v DurValue) Get() interface{} {
	return v.value
}
func (v *DurValue) Set(s string) {
	var err error
	interval, err := time.ParseDuration(s)
	if err != nil {
		log.Fatal("Duration Parse error")
	}
	v.value = schema.Duration(interval)
}

func NewDurValue(s string) VariableValue {
	interval, err := time.ParseDuration(s)
	if err != nil {
		log.Fatal("Duration Parse error")
	}
	return &DurValue{value: schema.Duration(interval)}
}

func getEnv(variableName string, variableValue VariableValue, changed map[string]bool) (changedValue interface{}) {
	var stringVal string

	if variableValue == nil {
		log.Fatal("nil pointer in getEnv")
	}
	var exists bool
	stringVal, exists = os.LookupEnv(variableName)
	if !exists {
		log.Printf("variable "+variableName+" not presented in environment, remains default:%v", variableValue.Get())
		changed[variableName] = false
		return variableValue.Get()
	}
	variableValue.Set(stringVal)
	changed[variableName] = true
	log.Println("variable " + variableName + " presented in environment, value: " + stringVal)

	return variableValue.Get()
}
