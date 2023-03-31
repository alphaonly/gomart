package handlers

import (
	"context"
	"errors"
	"fmt"

	"github.com/alphaonly/gomart/internal/schema"
	stor "github.com/alphaonly/gomart/internal/server/storage/interfaces"
	"github.com/theplant/luhn"
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
	err = eh.Storage.SaveUser(ctx, *u)
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
		return errors.New("401 login or	password is unknown")
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

func (eh EntityHandler) ValidateOrderNumber(order int) (ok bool, err error) {
	// data validation
	if order == 0 {
		return false, errors.New("400 no order number")
	}
	// Check if username authorized
	return luhn.Valid(order), nil

}
