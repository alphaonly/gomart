package handlers

import (
	"bytes"
	"fmt"

	"net/url"

	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alphaonly/gomart/internal/server/storage/implementations/mapstorage"
)

func TestHandleMetric(t *testing.T) {

	type want struct {
		code        int
		response    string
		contentType string
	}
	type requestParams struct {
		method string
		url    string
		want   want
	}

	data := url.Values{}

	metricsRequestsParam := make(map[string]requestParams)

	contentType := "text/plain"
	urlPrefix := ""
	//Check Url Ok
	urlStr := urlPrefix + "/update/gauge/Alloc/2.36912E+05"
	r1 := requestParams{method: http.MethodPost, url: urlStr,
		want: want{code: http.StatusOK, response: `{"status":"ok"}`, contentType: contentType}}

	//Check Url bad unknown metric
	urlStr = urlPrefix + "/update/gauge/Alerrorloc/2.36912E+05"
	r3 := requestParams{method: http.MethodPost, url: urlStr,
		want: want{code: http.StatusOK, response: ``, contentType: contentType}}
	//Check Url bad method
	urlStr = urlPrefix + "/update/gauge/Alloc/2.36912E+05"
	r5 := requestParams{method: http.MethodGet, url: urlStr,
		want: want{code: http.StatusMethodNotAllowed, response: `{"status":"ok"}`, contentType: contentType}}
	//Check Url empty metric
	urlStr = urlPrefix + "/update/gauge//2.36912E+05"
	r6 := requestParams{method: http.MethodPost, url: urlStr,
		want: want{code: http.StatusNotFound, response: `{"status":"ok"}`, contentType: contentType}}
	//Check Url empty metric value
	urlStr = urlPrefix + "/update/gauge/counter/"
	r7 := requestParams{method: http.MethodPost, url: urlStr,
		want: want{code: http.StatusNotFound, response: `{"status":"ok"}`, contentType: contentType}}

	//var r4 requestParams

	metricsRequestsParam["r1"] = r1

	metricsRequestsParam["r3"] = r3

	metricsRequestsParam["r5"] = r5
	metricsRequestsParam["r6"] = r6
	metricsRequestsParam["r7"] = r7

	tests := []struct {
		name string
		ID   string
		want want
	}{
		{
			name: "test#1 positive",
			ID:   "r1",
			want: metricsRequestsParam["r1"].want,
		},

		{
			name: "test#3 negative",
			ID:   "r3",
			want: metricsRequestsParam["r3"].want,
		},

		{
			name: "test#5 negative",
			ID:   "r5",
			want: metricsRequestsParam["r5"].want,
		},
		{
			name: "test#6 negative",
			ID:   "r6",
			want: metricsRequestsParam["r6"].want,
		},
		{
			name: "test#7 negative",
			ID:   "r7",
			want: metricsRequestsParam["r7"].want,
		},
	}
	fmt.Println("start!")

	s := mapstorage.New()
	h := Handlers{Storage: s}

	r := h.NewRouter()

	ts := httptest.NewServer(r)
	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fmt.Println("url from test:" + metricsRequestsParam[tt.ID].url)

			request := httptest.NewRequest(metricsRequestsParam[tt.ID].method, metricsRequestsParam[tt.ID].url, bytes.NewBufferString(data.Encode()))

			w := httptest.NewRecorder()

			//h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			//
			//	handlers := Handlers{}
			//	handlers.HandlePostMetric(w, r)
			//})

			r.ServeHTTP(w, request)

			response := w.Result()
			if response.StatusCode != tt.want.code {
				t.Errorf("error code %v want %v", response.StatusCode, tt.want.code)
				fmt.Println(response)
				fmt.Println(w.Body.String())

			}

			if (response.StatusCode == http.StatusOK) &&
				(response.Header.Get("Content-type") != tt.want.contentType) {
				t.Errorf("error contentType %v want %v", response.Header.Get("Content-type"), tt.want.contentType)
			}
			err := response.Body.Close()
			if err != nil {
				t.Errorf("response body close error: %v response", response.Body)
			}
		})

	}

}
