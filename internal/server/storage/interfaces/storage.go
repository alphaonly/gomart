package storage

import (
	"context"

	"github.com/alphaonly/gomart/internal/schema"
)

type Storage interface {
	GetUser(ctx context.Context, name string) (u *schema.User, err error)
	SaveUser(ctx context.Context, u *schema.User) (err error)

	GetOrder(ctx context.Context, orderNumber int64) (o *schema.Order, err error)
	SaveOrder(ctx context.Context, o schema.Order) (err error)
	GetOrdersList(ctx context.Context, userName string) (ol schema.Orders, err error)
	GetNewOrdersList(ctx context.Context) (ol schema.Orders, err error)
	SaveWithdrawal(ctx context.Context, w schema.Withdrawal) (err error)
	GetWithdrawalsList(ctx context.Context, userName string) (wl *schema.Withdrawals, err error)
}
