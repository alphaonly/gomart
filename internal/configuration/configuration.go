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

const ServerDefaultJSON = `{"ADDRESS":"localhost:8080","STORE_INTERVAL": "300s","STORE_FILE":"/tmp/devops-metrics-db.json","RESTORE":true,"KEY":""}`
const AgentDefaultJSON = `{"POLL_INTERVAL":"2s","REPORT_INTERVAL":"10s","ADDRESS":"localhost:8080","SCHEME":"http","USE_JSON":1,"KEY":"","RATE_LIMIT":1}`

type AgentConfiguration struct {
	Address        string          `json:"ADDRESS,omitempty"`
	Scheme         string          `json:"SCHEME,omitempty"`
	CompressType   string          `json:"COMPRESS_TYPE,omitempty"`
	PollInterval   schema.Duration `json:"POLL_INTERVAL,omitempty"`
	ReportInterval schema.Duration `json:"REPORT_INTERVAL,omitempty"`
	UseJSON        int             `json:"USE_JSON,omitempty"`
	Key            string          `json:"KEY,omitempty"`
	RateLimit      int             `json:"RATE_LIMIT,omitempty"`
	EnvChanged     map[string]bool
}

type ServerConfiguration struct {
	Address       string          `json:"ADDRESS,omitempty"`
	StoreInterval schema.Duration `json:"STORE_INTERVAL,omitempty"`
	StoreFile     string          `json:"STORE_FILE,omitempty"`
	Restore       bool            `json:"RESTORE,omitempty"`
	Port          string          `json:"PORT,omitempty"` //additionally for listen and serve func
	Key           string          `json:"KEY,omitempty"`
	DatabaseDsn   string          `json:"DATABASE_DSN,omitempty"`
	EnvChanged    map[string]bool
}

type AgentConfigurationOption func(*AgentConfiguration)
type ServerConfigurationOption func(*ServerConfiguration)

func UnMarshalServerDefaults(s string) ServerConfiguration {
	sc := ServerConfiguration{}
	err := json.Unmarshal([]byte(s), &sc)
	if err != nil {
		log.Fatal("cannot unmarshal server configuration")
	}
	return sc

}
func UnMarshalAgentDefaults(s string) AgentConfiguration {
	ac := AgentConfiguration{}
	err := json.Unmarshal([]byte(s), &ac)
	if err != nil {
		log.Fatal("cannot unmarshal server configuration")
	}
	return ac
}

func NewAgentConfiguration() *AgentConfiguration {
	c := UnMarshalAgentDefaults(AgentDefaultJSON)
	c.EnvChanged = make(map[string]bool)

	return &c

}

func NewServerConfiguration() *ServerConfiguration {
	c := UnMarshalServerDefaults(ServerDefaultJSON)
	c.Port = ":" + strings.Split(c.Address, ":")[1]
	c.EnvChanged = make(map[string]bool)
	return &c

}

func NewAgentConf(options ...AgentConfigurationOption) *AgentConfiguration {
	c := UnMarshalAgentDefaults(AgentDefaultJSON)
	c.EnvChanged = make(map[string]bool)
	for _, option := range options {
		option(&c)
	}
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

func UpdateACFromEnvironment(c *AgentConfiguration) {

	c.Address = getEnv("ADDRESS", &StrValue{c.Address}, c.EnvChanged).(string)
	c.CompressType = getEnv("COMPRESS_TYPE", &StrValue{c.CompressType}, c.EnvChanged).(string)
	c.PollInterval = getEnv("POLL_INTERVAL", &DurValue{c.PollInterval}, c.EnvChanged).(schema.Duration)
	c.ReportInterval = getEnv("REPORT_INTERVAL", &DurValue{c.ReportInterval}, c.EnvChanged).(schema.Duration)
	c.Scheme = getEnv("SCHEME", &StrValue{c.Scheme}, c.EnvChanged).(string)
	c.UseJSON = getEnv("USE_JSON", &IntValue{c.UseJSON}, c.EnvChanged).(int)
	c.Key = getEnv("KEY", &StrValue{c.Key}, c.EnvChanged).(string)
	c.RateLimit = getEnv("RATE_LIMIT", &IntValue{c.RateLimit}, c.EnvChanged).(int)
}

func UpdateSCFromEnvironment(c *ServerConfiguration) {
	c.Address = getEnv("ADDRESS", &StrValue{c.Address}, c.EnvChanged).(string)
	c.Restore = getEnv("RESTORE", &BoolValue{c.Restore}, c.EnvChanged).(bool)
	c.StoreFile = getEnv("STORE_FILE", &StrValue{c.StoreFile}, c.EnvChanged).(string)
	c.StoreInterval = getEnv("STORE_INTERVAL", &DurValue{c.StoreInterval}, c.EnvChanged).(schema.Duration)
	c.Key = getEnv("KEY", &StrValue{c.Key}, c.EnvChanged).(string)
	//PORT is derived from ADDRESS
	c.Port = ":" + strings.Split(c.Address, ":")[1]
	c.DatabaseDsn = getEnv("DATABASE_DSN", &StrValue{c.DatabaseDsn}, c.EnvChanged).(string)
}

func UpdateACFromFlags(c *AgentConfiguration) {
	dc := NewAgentConfiguration()
	var (
		a = flag.String("a", dc.Address, "Domain name and :port")
		p = flag.Duration("p", time.Duration(dc.PollInterval), "Poll interval")
		r = flag.Duration("r", time.Duration(dc.ReportInterval), "Report interval")
		j = flag.Int("j", dc.UseJSON, "Use JSON 0-No JSON,1- JSON, 2-JSON Batch")
		t = flag.String("t", dc.CompressType, "Compress type: \"deflate\" supported")
		k = flag.String("k", dc.Key, "string key for hash signing")
		l = flag.Int("l", dc.RateLimit, "Number of parallel inbound requests ")
	)

	flag.Parse()

	//Если значение параметра из переменных окружения равно по умолчанию, то обновляем из флагов

	message := "variable %v  updated from flags, value %v"
	if !c.EnvChanged["ADDRESS"] {
		c.Address = *a
		log.Printf(message, "ADDRESS", c.Address)
	}
	if !c.EnvChanged["POLL_INTERVAL"] {
		c.PollInterval = schema.Duration(*p)
		log.Printf(message, "POLL_INTERVAL", c.PollInterval)
	}

	if !c.EnvChanged["REPORT_INTERVAL"] {
		c.ReportInterval = schema.Duration(*r)
		log.Printf(message, "REPORT_INTERVAL", c.ReportInterval)
	}

	if !c.EnvChanged["USE_JSON"] {
		c.UseJSON = *j
		log.Printf(message, "USE_JSON", c.UseJSON)
	}

	if !c.EnvChanged["COMPRESS_TYPE"] {
		c.CompressType = *t
		log.Printf(message, "COMPRESS_TYPE", c.CompressType)
	}
	if !c.EnvChanged["KEY"] {
		c.Key = *k
		log.Printf(message, "KEY", c.Key)
	}
	if !c.EnvChanged["RATE_LIMIT"] {
		c.RateLimit = *l
		log.Printf(message, "RATE_LIMIT", c.RateLimit)
	}
}

func UpdateSCFromFlags(c *ServerConfiguration) {

	dc := NewServerConfiguration()

	var (
		a = flag.String("a", dc.Address, "Domain name and :port")
		i = flag.Duration("i", time.Duration(dc.StoreInterval), "Store interval")
		f = flag.String("f", dc.StoreFile, "Store file full path")
		r = flag.Bool("r", dc.Restore, "Restore from external storage:true/false")
		k = flag.String("k", dc.Key, "string key for hash signing")
		d = flag.String("d", dc.DatabaseDsn, "database destination string")
	)
	flag.Parse()

	message := "variable %v  updated from flags, value %v"
	//Если значение из переменных равно значению по умолчанию, тогда берем из flagS
	if !c.EnvChanged["ADDRESS"] {
		c.Address = *a
		c.Port = ":" + strings.Split(c.Address, ":")[1]
		log.Printf(message, "ADDRESS", c.Address)
		log.Printf(message, "PORT", c.Port)
	}
	if !c.EnvChanged["STORE_INTERVAL"] {
		c.StoreInterval = schema.Duration(*i)
		log.Printf(message, "STORE_INTERVAL", c.StoreInterval)
	}
	if !c.EnvChanged["STORE_FILE"] {
		c.StoreFile = *f
		log.Printf(message, "STORE_FILE", c.StoreFile)
	}
	if !c.EnvChanged["RESTORE"] {
		c.Restore = *r
		log.Printf(message, "RESTORE", c.Restore)
	}
	if !c.EnvChanged["KEY"] {
		c.Key = *k
		log.Printf(message, "KEY", c.Key)
	}
	if !c.EnvChanged["DATABASE_DSN"] {
		c.DatabaseDsn = *d
		log.Printf(message, "DATABASE_DSN", c.DatabaseDsn)
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
