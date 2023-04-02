package accrual

import (
	"context"
	"errors"
	"github.com/alphaonly/gomart/internal/schema"
	storage "github.com/alphaonly/gomart/internal/server/storage/interfaces"
	"github.com/go-resty/resty/v2"
	"log"
	"net/http"
	"strconv"
	"time"
)

//Periodically checking orders' accrual from remote service

type Configuration struct {
}

func NewChecker(serviceAddress string, requestTime int64, storage storage.Storage) (c *Checker) {
	return &Checker{
		serviceAddress: serviceAddress,
		requestTime:    time.Duration(requestTime) * time.Millisecond,
		storage:        storage,
	}
}

type Checker struct {
	serviceAddress string
	requestTime    time.Duration //200 * time.Millisecond
	storage        storage.Storage
}
type Response struct {
	order   string
	status  string
	accrual float64
}

func (c Checker) Run(ctx context.Context) {
	ticker := time.NewTicker(c.requestTime)

	//resty customization
	errRedirectBlocked := errors.New("HTTP redirect blocked")
	redirPolicy := resty.RedirectPolicyFunc(func(_ *http.Request, _ []*http.Request) error {
		return errRedirectBlocked
	})
	httpc := resty.New().
		SetBaseURL(c.serviceAddress).
		SetRedirectPolicy(redirPolicy)

doItAGain:
	for {
		select {
		case <-ticker.C:
			//Getting New unprocessed orders to make a request to accrual system
			oList, err := c.storage.GetNewOrdersList(ctx)
			if err != nil {
				log.Fatal("can not get new orders list")
			}

			for orderNumber, data := range oList {

				orderNumberStr := strconv.Itoa(int(orderNumber))
				req := httpc.R().
					SetHeader("Content-Type", "application/json")

				response := Response{}
				resp, err := req.
					SetResult(response).
					Get("api/orders/" + orderNumberStr)
				if err != nil {
					log.Printf("order %v response error: %v", orderNumber, resp)
					continue
				}
				if response.status != "PROCESSED" {
					continue
				}

				data.Accrual = response.accrual
				data.Status = schema.OrderStatus["PROCESSED"]

				err = c.storage.SaveOrder(ctx, data)
				if err != nil {
					log.Fatal("unable to save order")
				}
			}

		case <-ctx.Done():
			break doItAGain
		}
	}

}
