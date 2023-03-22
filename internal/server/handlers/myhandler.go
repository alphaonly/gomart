package handlers

import "net/http"

type HandlerData struct {
	body []byte
}
type MyHandler interface {
	// http.Handler
	validate() bool
	handle() (ok bool)
	respond() (ok bool)
	Run(data *HandlerData, next MyHandler)
}
type PingHandler struct {
	F    func(w http.ResponseWriter, r *http.Request)
	data *HandlerData
}

func (m PingHandler) validate() bool     { return true }
func (m PingHandler) handle() (ok bool)  { return true }
func (m PingHandler) respond() (ok bool) { return true }
func (m PingHandler) Run(data *HandlerData, next MyHandler) {
	m.data = data
	m.F = func(w http.ResponseWriter, r *http.Request) {
		if !m.validate() {
			return
		}
		if !m.handle() {
			return
		}
		if !m.respond() {
			return
		}

		if next != nil {
			next.Run(data, nil) //!!!!
		}

	}
}

var pp PingHandler
var mh MyHandler = pp
