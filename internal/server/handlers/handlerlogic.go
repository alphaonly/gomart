package handlers

import (
	"context"
	"errors"
	"fmt"
	"github.com/alphaonly/gomart/internal/schema"
	stor "github.com/alphaonly/gomart/internal/server/storage/interfaces"
	"github.com/theplant/luhn"
	"sort"
	"strconv"
	"time"
)

type EntityHandler struct {
	Storage         stor.Storage
	AuthorizedUsers map[string]bool
}

func (eh EntityHandler) RegisterUser(ctx context.Context, u *schema.User) (err error) {
	// data validation
	if u.User == "" || u.Password == "" {
		return errors.New("400 user or password is empty")
	}
	// Check if username exists
	userChk, err := eh.Storage.GetUser(ctx, u.User)

	if err != nil {
		return fmt.Errorf("cannot get user from storage", err)
	}
	if userChk != nil {
		//login has already been occupied
		return errors.New("409 login " + userChk.User + " is occupied")
	}
	err = eh.Storage.SaveUser(ctx, u)
	if err != nil {
		return fmt.Errorf("cannot save user in storage", err)
	}
	return nil
}

func (eh EntityHandler) AuthenticateUser(ctx context.Context, u *schema.User) (err error) {
	// data validation
	if u.User == "" || u.Password == "" {
		return errors.New("400 user or password is empty")
	}
	// Check if username exists
	userInStorage, err := eh.Storage.GetUser(ctx, u.User)
	if !u.CheckIdentity(userInStorage) {
		return errors.New("401 login or password is unknown")
	}
	eh.AuthorizedUsers[u.User] = true

	return nil
}

func (eh EntityHandler) CheckIfUserAuthorized(user string) (ok bool, err error) {
	// data validation
	if user == "" {
		return false, errors.New("400 login is empty")
	}
	// Check if username authorized
	return eh.AuthorizedUsers[user], nil

}

func (eh EntityHandler) ValidateOrderNumber(ctx context.Context, orderNumberStr string, user string) (orderNum int64, err error) {
	orderNumber, err := strconv.Atoi(orderNumberStr)
	if err != nil {
		return 0, fmt.Errorf("400 order number bad number value %w", err)
	}
	// order number format check
	if orderNumber <= 0 {
		return int64(orderNumber), fmt.Errorf("400 no order number zero or less:%v", orderNumber)
	}
	// orderNumber number validation according Luhn algorithm
	if !luhn.Valid(orderNumber) {
		return int64(orderNumber), fmt.Errorf("422 no order number with Luhn: %v", orderNumber)
	}
	// Check if orderNumber had already existed
	orderChk, err := eh.Storage.GetOrder(ctx, int64(orderNumber))
	if err != nil {
		return int64(orderNumber), nil
	}
	//Order exists, check user
	if user == orderChk.User {
		return int64(orderNumber), fmt.Errorf("200 order %v exists with user %v", orderNumber, user)
	}
	return int64(orderNumber), fmt.Errorf("409 order %v exists with another user %v", orderNumber, orderChk.User)
}

func (eh EntityHandler) GetUsersOrders(ctx context.Context, userName string) (orders schema.Orders, err error) {
	// data validation
	if userName == "" {
		return nil, fmt.Errorf("400 user %v is empty", userName)
	}
	//getOrders
	orderslist, err := eh.Storage.GetOrdersList(ctx, userName)
	if err != nil {
		return nil, fmt.Errorf("204 no orders for user %v %w", userName, err)
	}
	return orderslist, nil
}

type UserBalanceResponse struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

func (eh EntityHandler) GetUserBalance(ctx context.Context, userName string) (response *UserBalanceResponse, err error) {
	// data validation
	if userName == "" {
		return nil, fmt.Errorf("400 user %v is empty", userName)
	}
	//getUser
	user, err := eh.Storage.GetUser(ctx, userName)
	if err != nil {
		return nil, err
	}
	return &UserBalanceResponse{user.Accrual, user.Withdrawal}, nil
}

type UserWithdrawalRequest struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}

func (eh EntityHandler) MakeUserWithdrawal(ctx context.Context, userName string, request UserWithdrawalRequest) (err error) {
	// data validation
	if userName == "" {
		return fmt.Errorf("400 user %v is empty", userName)
	}
	//check order number
	orderNumber, err := eh.ValidateOrderNumber(ctx, request.Order, userName)
	if err != nil {
		return fmt.Errorf("422 order number invalid %v %w", orderNumber, err)
	}
	//getUser
	user, err := eh.Storage.GetUser(ctx, userName)
	if err != nil {
		return err
	}
	//Calculate new accrual
	newAccrual := user.Accrual - request.Sum
	if newAccrual < 0 {
		return fmt.Errorf("402 insufficient funds for withdrawal")
	}
	//Update Users data
	user.Accrual = newAccrual
	user.Withdrawal += request.Sum
	err = eh.Storage.SaveUser(ctx, user)
	if err != nil {
		return fmt.Errorf("500 can not update data of user %v after withrawal attempt on order %v %w", userName, orderNumber, err)
	}
	//Add withdrawal
	w := schema.Withdrawal{
		User:       userName,
		Processed:  schema.CreatedTime(time.Now()),
		Withdrawal: request.Sum,
	}
	err = eh.Storage.SaveWithdrawal(ctx, w)
	if err != nil {
		return fmt.Errorf("500 can not create withdrawal data of user %v after withrawal attempt on order %v %w", userName, orderNumber, err)
	}
	return nil
}
func (eh EntityHandler) GetUsersWithdrawals(ctx context.Context, userName string) (withdrawals *schema.Withdrawals, err error) {
	// data validation
	if userName == "" {
		return nil, fmt.Errorf("400 user %v is empty", userName)
	}
	//getOrders
	wList, err := eh.Storage.GetWithdrawalsList(ctx, userName)
	if err != nil {
		return nil, fmt.Errorf("500 internal error on getting withdrawals for user %v %w", userName, err)
	}
	if len(*wList) == 0 {
		return nil, fmt.Errorf("204 no withdrawals for user %v %w", userName, err)
	}

	sort.Sort(schema.ByTimeDescending(*wList))
	return wList, nil
}
