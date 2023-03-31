package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	storage "github.com/alphaonly/gomart/internal/server/storage/interfaces"

	"github.com/alphaonly/gomart/internal/configuration"
	"github.com/alphaonly/gomart/internal/schema"
	"github.com/alphaonly/gomart/internal/signchecker"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
)

type Handlers struct {
	Storage       storage.Storage
	Signer        signchecker.Signer
	Conf          configuration.ServerConfiguration
	EntityHandler *EntityHandler
}

func (h *Handlers) WriteResponseBodyHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("WriteResponseBodyHandler invoked")

		//read body
		var bytesData []byte
		var err error
		var prev schema.PreviousBytes

		if p := r.Context().Value(schema.PKey1); p != nil {
			prev = p.(schema.PreviousBytes)
		}
		if prev != nil {
			//body from previous handler
			bytesData = prev
			log.Printf("got body from previous handler:%v", string(bytesData))
		} else {
			//body from request if there is no previous handler
			bytesData, err = io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusNotImplemented)
				return
			}
			log.Printf("got body from request:%v", string(bytesData))
		}
		//Set flag in case compressed data
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			w.Header().Set("Content-Encoding", "gzip")
		}
		//Set Response Header
		w.WriteHeader(http.StatusOK)
		//write Response Body
		_, err = w.Write(bytesData)
		if err != nil {
			log.Println("byteData writing error")
			http.Error(w, "byteData writing error", http.StatusInternalServerError)
			return
		}
	}

}

func (h *Handlers) HandlePing(w http.ResponseWriter, r *http.Request) {
	log.Println("HandlePing invoked")
	log.Println("server:HandlePing:database string:" + h.Conf.DatabaseURI)
	conn, err := pgx.Connect(r.Context(), h.Conf.DatabaseURI)
	if err != nil {
		httpError(w, errors.New("server: ping handler: Unable to connect to database:"+err.Error()), http.StatusInternalServerError)
		return
	}
	defer conn.Close(context.Background())
	log.Println("server: ping handler: connection established, 200 OK ")
	w.Write([]byte("200 OK"))
	w.WriteHeader(http.StatusOK)
}

func (h *Handlers) NewRouter() chi.Router {

	var (
	// writePost = h.WriteResponseBodyHandler
	//writeList = h.WriteResponseBodyHandler

	// compressPost = compression.GZipCompressionHandler
	//compressList = compression.GZipCompressionHandler

	// handlePost      = h.HandlePostMetricJSON
	// handlePostBatch = h.HandlePostMetricJSONBatch
	//handleList = h.HandleGetMetricFieldList
	//handleList = h.HandleGetMetricFieldList

	//The sequence for post JSON and respond compressed JSON if no value
	// postJSONAndGetCompressed = handlePost(compressPost(writePost()))
	//The sequence for post JSON and respond compressed JSON if no value receiving data in batch
	// postJSONAndGetCompressedBatch = handlePostBatch(compressPost(writePost()))

	//The sequence for get compressed metrics html list
	//getListCompressed = handleList(compressList(writeList()))
	// getListCompressed = h.HandleGetMetricFieldListSimple(nil)
	)
	r := chi.NewRouter()
	//

	// var p PingHandler
	r.Route("/", func(r chi.Router) {
		// r.Get("/", getListCompressed)
		r.Get("/ping", h.HandlePing)
		r.Get("/ping/", h.HandlePing)
		r.Get("/check/", h.HandleCheckHealth)
		// r.Get("/value/{TYPE}/{NAME}", h.HandleGetMetricValue)
		// r.Post("/value", postJSONAndGetCompressed)
		// // r.Post("/value/", postJSONAndGetCompressed)
		// r.Post("/update", postJSONAndGetCompressed)
		// r.Post("/update/", postJSONAndGetCompressed)
		// r.Post("/updates", postJSONAndGetCompressedBatch)
		// r.Post("/updates/", postJSONAndGetCompressedBatch)
		// r.Post("/update/{TYPE}/{NAME}/{VALUE}", h.HandlePostMetric)
		// r.Post("/update/{TYPE}/{NAME}/", h.HandlePostErrorPattern)
		// r.Post("/update/{TYPE}/", h.HandlePostErrorPatternNoName)

		r.Post("/api/user/register", h.PostValidation(h.HandlePostUserRegister(nil)))
		r.Post("/api/{USER}/login", h.HandlePostUserLogin(nil))
		r.Post("/api/{USER}/orders", h.HandlePostUserOrders(nil))
		r.Post("/api/{USER}/balance/withdraw", h.HandlePostUserBalanceWithdraw(nil))
		r.Get("/api/{USER}/orders", h.HandleGetUserOrders(nil))
		r.Get("/api/{USER}/balance", h.HandleGetUserBalance(nil))
		r.Get("/api/{USER}/withdrawals", h.HandleGetUserWithdrawals(nil))

	})

	return r
}

func logFatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func (h *Handlers) HandleCheckHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {

		w.WriteHeader(http.StatusOK)

	}
}

func (h *Handlers) GetValidation(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("HandleGetValidation invoked")
		//Validation
		if r.Method != http.MethodGet {
			http.Error(w, "Only GET is allowed", http.StatusMethodNotAllowed)
			return
		}
		if next != nil {
			//call further handler with context parameters
			next.ServeHTTP(w, r)
			return
		}
	}
}
func (h *Handlers) PostValidation(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("HandlePostValidation invoked")
		//Validation
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST is allowed", http.StatusMethodNotAllowed)
			return
		}
		if next != nil {
			//call further handler with context parameters
			next.ServeHTTP(w, r)
			return
		}
	}
}
func (h *Handlers) HandlePostUserRegister(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("HandlePostUserRegister invoked")

		// //Basic authentication
		// userBA, passwordBA, ok := r.BasicAuth()
		// if !ok {
		// 	httpError(w, "basic authentication is not ok", http.StatusInternalServerError)
		// }

		//Handling body
		requestByteData, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Unrecognized json request ", http.StatusBadRequest)
			return
		}
		u := new(schema.User)
		err = json.Unmarshal(requestByteData, u)
		if err != nil {
			http.Error(w, "Error json-marshal request data", http.StatusBadRequest)
			return
		}
		//Logic
		err = h.EntityHandler.RegisterUser(r.Context(), u)
		if err != nil {
			if strings.Contains(err.Error(), "400") {
				http.Error(w, "login "+u.User+": bad request", http.StatusBadRequest)
				return
			}
			if strings.Contains(err.Error(), "409") {
				http.Error(w, "login "+u.User+"is occupied", http.StatusConflict)
				return
			}
			http.Error(w, "login "+u.User+"register internal error", http.StatusInternalServerError)
			return
		}
		//Response
		w.WriteHeader(http.StatusOK)

	}
}
func (h *Handlers) HandlePostUserLogin(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("HandlePostUserLogin invoked")

		//Handling body
		requestByteData, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Unrecognized json request ", http.StatusBadRequest)
			return
		}
		u := new(schema.User)
		err = json.Unmarshal(requestByteData, u)
		if err != nil {
			http.Error(w, "Error json-marshal request data", http.StatusBadRequest)
			return
		}
		//Logic
		err = h.EntityHandler.AuthenticateUser(r.Context(), u)
		if err != nil {
			if strings.Contains(err.Error(), "400") {
				http.Error(w, "login "+u.User+": bad request", http.StatusBadRequest)
				return
			}
			if strings.Contains(err.Error(), "409") {
				http.Error(w, "login "+u.User+"is occupied", http.StatusConflict)
				return
			}
			http.Error(w, "login "+u.User+"register internal error", http.StatusInternalServerError)
			return
		}
		//Response
		w.WriteHeader(http.StatusOK)
	}
}

func (h *Handlers) BasicUserAuthorization(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("BasicUserAuthorization invoked")
		//Basic authentication
		userBA, _, ok := r.BasicAuth()
		if !ok {
			http.Error(w, "basic authentication is not ok", http.StatusInternalServerError)
			return
		}
		var err error
		ok, err = h.EntityHandler.CheckIfUserAuthorized(userBA)
		if err != nil {
			if strings.Contains(err.Error(), "400") {
				httpError(w, fmt.Errorf("login %v: bad request %w", userBA, err), http.StatusBadRequest)
				return
			}
		}
		if !ok {
			httpError(w, errors.New("login "+userBA+" not authorized"), http.StatusBadRequest)
			return
		}

		if next == nil {
			log.Fatal("handler requires next handler not nil")
		}
		//call further handler with context parameters
		ctx := context.WithValue(r.Context(), schema.PKey1, schema.CtxUName(userBA))
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
func (h *Handlers) HandlePostUserOrders(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("HandlePostUserOrders invoked")
		//Get parameters from previous handler
		user, err := getPreviousParameter[schema.CtxUName, schema.ContextKey](r, schema.CtxKeyUName)
		if err != nil {
			httpError(w, fmt.Errorf("cannot get userName from context %w", err), http.StatusInternalServerError)
			return
		}
		//Handling
		requestByteData, err := io.ReadAll(r.Body)
		if err != nil {
			httpError(w, fmt.Errorf("unrecognized request body %w", err), http.StatusBadRequest)
			return
		}
		orderNumber, err := strconv.Atoi(string(requestByteData))
		if err != nil {
			httpError(w, fmt.Errorf("unrecognized order number %w", err), http.StatusBadRequest)
			return
		}
		ok, err := h.EntityHandler.ValidateOrderNumber(orderNumber)
		if err != nil {
			httpError(w, fmt.Errorf("order %v Luhn check  internal error %w", orderNumber, err), http.StatusInternalServerError)
			return
		}
		if !ok {
			httpError(w, fmt.Errorf("order's number %v not valid", orderNumber), http.StatusInternalServerError)
			return
		}

		orderList, err := h.Storage.GetOrdersList(r.Context(), schema.User{User: string(user)})
		if err != nil {
			httpError(w, fmt.Errorf("cannot get orders list by user  %w", err), http.StatusInternalServerError)
			return
		}
		val, ok := orderList[int64(orderNumber)]
		if ok {
			if val.User == string(user) {
				log.Printf("order %v has already been created by user %v", orderNumber, user)
				w.WriteHeader(http.StatusOK)
			} else {
				log.Printf("order %v has already been created by different %v", orderNumber, user)
				w.WriteHeader(http.StatusOK)
			}

		}
		//TODO: GetOrder needed by number to compare user
		// order := schema.Order{Order: int64(orderNumber)}
		//Response

	}
}
func (h *Handlers) HandleGetUserOrders(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("HandleGetUserOrders invoked")
		//Get Parameters

		//Handling
		//Response
	}
}
func (h *Handlers) HandleGetUserBalance(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("HandleGetUserBalance invoked")
		//Get Parameters

		//Handling
		//Response
	}
}
func (h *Handlers) HandlePostUserBalanceWithdraw(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("HandlePostUserBalanceWithdraw invoked")
		//Get Parameters

		//Handling
		//Response
	}
}
func (h *Handlers) HandleGetUserWithdrawals(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("HandleGetUserWithdrawals invoked")
		//Get Parameters

		//Handling
		//Response
	}
}

func httpErrorW(w http.ResponseWriter, eStr string, err error, status int) {
	if err != nil {
		newE := fmt.Errorf(eStr+" %w", err)
		httpError(w, newE, status)
		log.Println("server:" + newE.Error())
	}
}

func httpError(w http.ResponseWriter, err error, status int) {
	if err != nil {
		http.Error(w, err.Error(), status)
		log.Println("server:" + err.Error())
	}
}

func getPreviousParameter[T any, V any](r *http.Request, key V) (data T, err error) {
	var prev T
	var p any

	if p = r.Context().Value(key); p == nil {
		log.Fatal("got nil data from previous handler")
		return prev, errors.New("got nil data from previous handler")
	}

	return p.(T), nil

}
