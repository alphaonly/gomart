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
	"time"

	storage "github.com/alphaonly/gomart/internal/server/storage/interfaces"

	"github.com/alphaonly/gomart/internal/configuration"
	"github.com/alphaonly/gomart/internal/schema"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
)

type Handlers struct {
	Storage       storage.Storage
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

	r.Route("/", func(r chi.Router) {
		// r.Get("/", getListCompressed)
		r.Get("/ping", h.HandlePing)
		r.Get("/ping/", h.HandlePing)
		r.Get("/check/", h.HandleCheckHealth)
		r.Post("/api/user/register", h.PostValidation(h.HandlePostUserRegister(nil)))
		r.Post("/api/user/login", h.PostValidation(h.HandlePostUserLogin(nil)))
		r.Post("/api/user/orders", h.PostValidation(h.BasicUserAuthorization(h.HandlePostUserOrders(nil))))
		r.Post("/api/user/balance/withdraw", h.PostValidation(h.BasicUserAuthorization(h.HandlePostUserBalanceWithdraw(nil))))
		r.Get("/api/user/orders", h.GetValidation(h.BasicUserAuthorization(h.HandleGetUserOrders(nil))))
		r.Get("/api/user/balance", h.GetValidation(h.BasicUserAuthorization(h.HandleGetUserBalance(nil))))
		r.Get("/api/user/withdrawals", h.GetValidation(h.BasicUserAuthorization(h.HandleGetUserWithdrawals(nil))))

		//Mock for accrual system (in case similar addresses) returns +5
		r.Get("/api/orders/{number}", h.GetValidation(h.HandleGetUserWithdrawals(nil)))
		//GET /api/orders/{number}
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

		orderNumber, err := h.EntityHandler.ValidateOrderNumber(r.Context(), string(requestByteData), string(user))
		if err != nil {
			if strings.Contains(err.Error(), "400") {
				httpErrorW(w, fmt.Sprintf("order number  %v insufficient format %w", orderNumber), err, http.StatusBadRequest)
				return
			}
			if strings.Contains(err.Error(), "422") {
				httpErrorW(w, fmt.Sprintf("order %v insufficient format %w", orderNumber), err, http.StatusUnprocessableEntity)
				return
			}
			if strings.Contains(err.Error(), "409") {
				httpErrorW(w, fmt.Sprintf("order %v exists %w", orderNumber), err, http.StatusConflict)
				return
			}
			if strings.Contains(err.Error(), "200") {
				log.Printf("order %v exists: %v", orderNumber, err.Error())
				w.WriteHeader(http.StatusOK)
				return
			}
		}
		//Create object order
		o := schema.Order{
			Order:   orderNumber,
			User:    string(user),
			Status:  schema.OrderStatus["NEW"],
			Created: schema.CreatedTime(time.Now()),
		}
		err = h.Storage.SaveOrder(r.Context(), o)
		if err != nil {
			httpErrorW(w, fmt.Sprintf("order's number %v not saved", orderNumber), err, http.StatusInternalServerError)
			return
		}
		//Response
		w.WriteHeader(http.StatusAccepted)
	}
}
func (h *Handlers) HandleGetUserOrders(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("HandleGetUserOrders invoked")

		//Get parameters from previous handler
		userName, err := getPreviousParameter[schema.CtxUName, schema.ContextKey](r, schema.CtxKeyUName)
		if err != nil {
			httpError(w, fmt.Errorf("cannot get userName from context %w", err), http.StatusInternalServerError)
			return
		}
		//Handling
		orderList, err := h.EntityHandler.GetUsersOrders(r.Context(), string(userName))
		if strings.Contains(err.Error(), "204") {
			httpErrorW(w, fmt.Sprintf("No orders for user %v", userName), err, http.StatusNoContent)
			return
		}
		//Response
		bytes, err := json.Marshal(orderList)
		if err != nil {
			httpErrorW(w, fmt.Sprintf("user %v order list json marshal error", userName), err, http.StatusInternalServerError)
			return
		}
		_, err = w.Write(bytes)
		if err != nil {
			httpErrorW(w, fmt.Sprintf("user %v HandleGetUserOrders write response error", userName), err, http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	}
}
func (h *Handlers) HandleGetUserBalance(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("HandleGetUserBalance invoked")
		//Get parameters from previous handler
		userName, err := getPreviousParameter[schema.CtxUName, schema.ContextKey](r, schema.CtxKeyUName)
		if err != nil {
			httpError(w, fmt.Errorf("cannot get userName from context %w", err), http.StatusInternalServerError)
			return
		}
		//Handling
		balance, err := h.EntityHandler.GetUserBalance(r.Context(), string(userName))
		if err != nil {
			httpError(w, fmt.Errorf("cannot get user data by userName %v from context %w", userName, err), http.StatusInternalServerError)
			return
		}
		//Response
		bytes, err := json.Marshal(balance)
		if err != nil {
			httpErrorW(w, fmt.Sprintf("user %v balance json marshal error", userName), err, http.StatusInternalServerError)
			return
		}
		_, err = w.Write(bytes)
		if err != nil {
			httpErrorW(w, fmt.Sprintf("user %v balance write response error", userName), err, http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	}
}
func (h *Handlers) HandlePostUserBalanceWithdraw(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("HandlePostUserBalanceWithdraw invoked")
		//Get parameters from previous handler
		userName, err := getPreviousParameter[schema.CtxUName, schema.ContextKey](r, schema.CtxKeyUName)
		if err != nil {
			httpError(w, fmt.Errorf("cannot get userName from context %w", err), http.StatusInternalServerError)
			return
		}
		//Handling
		requestByteData, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Unrecognized json request ", http.StatusBadRequest)
			return
		}
		userWithdrawalRequest := UserWithdrawalRequest{}
		err = json.Unmarshal(requestByteData, &userWithdrawalRequest)
		if err != nil {
			http.Error(w, "Error json-marshal request data", http.StatusBadRequest)
			return
		}
		err = h.EntityHandler.MakeUserWithdrawal(r.Context(), string(userName), userWithdrawalRequest)
		if err != nil {
			if strings.Contains(err.Error(), "402") {
				httpErrorW(w, "make withdrawal error", err, http.StatusPaymentRequired)
				return
			}
			if strings.Contains(err.Error(), "422") {
				httpErrorW(w, "order number invalid", err, http.StatusUnprocessableEntity)
				return
			}
			if strings.Contains(err.Error(), "500") {
				httpErrorW(w, "internal error", err, http.StatusInternalServerError)
				return
			}
		}
		//Response
		w.WriteHeader(http.StatusOK)
	}
}
func (h *Handlers) HandleGetUserWithdrawals(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("HandleGetUserWithdrawals invoked")
		//Get parameters from previous handler
		userName, err := getPreviousParameter[schema.CtxUName, schema.ContextKey](r, schema.CtxKeyUName)
		if err != nil {
			httpError(w, fmt.Errorf("can not get userName from context %w", err), http.StatusInternalServerError)
			return
		}
		//Handling
		wList, err := h.EntityHandler.GetUsersWithdrawals(r.Context(), string(userName))
		if err != nil {
			if strings.Contains(err.Error(), "500") {
				httpErrorW(w, "internal error", err, http.StatusInternalServerError)
				return
			}
			if strings.Contains(err.Error(), "204") {
				httpErrorW(w, "no withdrawals", err, http.StatusNoContent)
				return
			}
		}
		//Response
		bytes, err := json.Marshal(wList)
		if err != nil {
			httpErrorW(w, fmt.Sprintf("user %v withdrawals list json marshal error", userName), err, http.StatusInternalServerError)
			return
		}
		_, err = w.Write(bytes)
		if err != nil {
			httpErrorW(w, fmt.Sprintf("user %v withdrawals list write response error", userName), err, http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	}
}

func (h *Handlers) HandleGetOrderAccrual(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("HandleGetOrderAccrual invoked")
		//Handling
		orderNumberStr := chi.URLParam(r, "number")
		if orderNumberStr == "" {
			httpError(w, fmt.Errorf("order number  %v is empty", orderNumberStr), http.StatusBadRequest)
			return
		}

		orderNumber, err := strconv.ParseInt(orderNumberStr, 10, 64)
		if err != nil {
			httpError(w, fmt.Errorf("order number  %v is bad format", orderNumberStr), http.StatusBadRequest)
			return
		}

		accrual := 5.3

		OrderAcrualResponse := struct {
			order   int64
			status  string
			accrual float64
		}{
			order:   orderNumber,
			status:  "PROCESSED",
			accrual: accrual,
		}

		//Response
		bytes, err := json.Marshal(OrderAcrualResponse)
		if err != nil {
			httpErrorW(w, fmt.Sprintf("order occrual response json marshal error"), err, http.StatusInternalServerError)
			return
		}
		_, err = w.Write(bytes)
		if err != nil {
			httpErrorW(w, fmt.Sprintf("order occrual response write response error"), err, http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
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
