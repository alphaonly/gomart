package main_test

import (
	"context"
	"log"
	"testing"
	"time"

	db "github.com/alphaonly/gomart/internal/server/storage/implementations/dbstorage"
	stor "github.com/alphaonly/gomart/internal/server/storage/interfaces"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"

	conf "github.com/alphaonly/gomart/internal/configuration"
	"github.com/alphaonly/gomart/internal/server"
	"github.com/alphaonly/gomart/internal/server/handlers"
)

func TestRun(t *testing.T) {

	tests := []struct {
		name string
		URL  string
		want string
	}{
		{
			name: "test#1 - Positive: server accessible",
			URL:  "http://localhost:8080/check/",
			want: "200 OK",
		},
		{
			name: "test#2 - Negative: server do not respond",
			URL:  "http://localhost:8080/chek/",
			want: "404 Not Found",
		},
	}

	sc := conf.NewServerConf(conf.UpdateSCFromEnvironment, conf.UpdateSCFromFlags)

	for _, tt := range tests {

		t.Run(tt.name, func(tst *testing.T) {
			//Up server for 3 seconds
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			var storage stor.Storage
			if sc.DatabaseURI == "" {
				log.Fatal("no database url")
			} else {
				storage = db.NewDBStorage(ctx, sc.DatabaseURI)
			}
			handlers := &handlers.Handlers{}

			server := server.New(sc, storage, handlers)

			go func() {
				err := server.Run(ctx)
				if err != nil {
					return
				}

			}()

			//wait for server is up
			time.Sleep(time.Second * 2)

			keys := make(map[string]string)
			keys["Content-Type"] = "plain/text"
			keys["Accept"] = "plain/text"

			client := resty.New()

			r := client.R().
				SetHeaders(keys)
			resp, err := r.Get(tt.URL)
			if err != nil {
				t.Logf("send new request error:%v", err)
			}
			t.Logf("get returned status:%v", resp.Status())
			if !assert.Equal(t, tt.want, resp.Status()) {
				t.Error("Server responded unexpectedly")

			}
			err = server.Shutdown(ctx)
			if err != nil {
				t.Fatal(err)
			}
		})
	}

}
