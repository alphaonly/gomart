package main

import (
	"errors"
	"log"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/alphaonly/gomart/internal/schema"
	"github.com/go-resty/resty/v2"
)

// create HTTP client without redirects support

func CheckCounterGzipHandlers() {
	// create HTTP client without redirects support
	errRedirectBlocked := errors.New("HTTP redirect blocked")
	redirPolicy := resty.RedirectPolicyFunc(func(_ *http.Request, _ []*http.Request) error {
		return errRedirectBlocked
	})
	httpc := resty.New().
		SetBaseURL("http://127.0.0.1:8080").
		SetRedirectPolicy(redirPolicy)

	id := "GetSetZip" + strconv.Itoa(rand.Intn(256))

	//value1, value2 := int64(rand.Int31()), int64(rand.Int31())
	req := httpc.R().
		SetHeader("Accept-Encoding", "gzip").
		SetHeader("Content-Type", "application/json")

	var result schema.MetricsJSON
	resp, err := req.
		SetBody(&schema.MetricsJSON{
			ID:    id,
			MType: "counter"}).
		SetResult(&result).
		Post("value/")
	if err != nil {
		log.Println(resp)
	}
}
func ff2() {

	CheckCounterGzipHandlers()

}
