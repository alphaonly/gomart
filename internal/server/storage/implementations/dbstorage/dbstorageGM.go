package storage

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/alphaonly/gomart/internal/schema"
	"github.com/jackc/pgx/v5/pgxpool"
)

// GetUser(ctx context.Context, name string) (u *schema.User, err error)
// SaveUser(ctx context.Context, u schema.User) (err error)

// SaveOrder(ctx context.Context, o schema.Order) (err error)
// GetOrdersList(ctx context.Context, u schema.User) (wl schema.Orders, err error)

// SaveWithdrawal(ctx context.Context, w schema.Withdrawal) (err error)
// GetWithdrawalsList(ctx context.Context, u schema.User) (wl schema.Withdrawals, err error)

// -d=postgres://postgres:mypassword@localhost:5432/yandex
const (
	selectLineUsersTable            = `SELECT user_id, password, accrual, withdrawal FROM public.users WHERE user_id=$1;`
	selectLineOrdersTable           = `SELECT order_id, user_id, status, accrual, uploaded_at FROM public.orders WHERE order_id=$1;`
	selectAllOrdersTableByUser      = `SELECT order_id, user_id,status, accrual, uploaded_at FROM public.orders WHERE user_id = $1;`
	selectAllOrdersTableByStatus    = `SELECT order_id, user_id, status,accrual, uploaded_at  FROM public.orders WHERE status = $1;`
	selectAllWithdrawalsTableByUser = `SELECT user_id,  uploaded_at, withdrawal FROM public.withdrawals WHERE user_id = $1;`

	createOrUpdateIfExistsUsersTable = `
	INSERT INTO public.users (user_id, password, accrual, withdrawal) 
	VALUES ($1, $2, $3, $4)
	ON CONFLICT (user_id) DO UPDATE 
  	SET password 	= $2,
	  	accrual 	= $3,
		withdrawal 	= $4; 
  	`
	createOrUpdateIfExistsOrdersTable = `
	  INSERT INTO public.orders (order_id, user_id, status,accrual) 
	  VALUES ($1, $2, $3,$4)
	  ON CONFLICT (order_id,user_id) DO UPDATE 
		SET password = $2,
			status = $3,
		    accrual = $4; 
		`
	createOrUpdateIfExistsWithdrawalsTable = `
		INSERT INTO public.withdrawals (user_id, uploaded_at, withdrawal) 
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id,uploaded_at) DO UPDATE 
		  SET withdrawn = $3; 
		  `
	createUsersTable = `create table public.users
	(	user_id varchar(40) not null primary key,
		password  TEXT not null,
		accrual double precision,
		withdrawal double precision 
	);`
	createOrdersTable = `create table public.orders
	(	order_id integer not null, 
		user_id varchar(40) not null,
		status integer,		
		accrual double precision,
		uploaded_at TEXT not null, 
		primary key (order_id,user_id)
	);`
	createWithdrawalsTable = `create table public.withdrawals
	(	user_id 		varchar(40) primary key,
		uploaded_at 	TEXT 		unique not null,
		withdrawal 		double precision 	not null	
	);`

	checkIfUsersTableExists       = `SELECT 'public.users'::regclass;`
	checkIfOrdersTableExists      = `SELECT 'public.orders'::regclass;`
	checkIfWithdrawalsTableExists = `SELECT 'public.withdrawals'::regclass;`
)

var message = []string{
	0: "DBStorage:unable to connect to database",
	1: "DBStorage:%v table has created",
	2: "DBStorage:unable to create %v table",
	3: "DBStorage:createOrUpdateIfExistsUsersTable error",
	4: "DBStorage:QueryRow failed: %v\n",
	5: "DBStorage:RowScan error",
	6: "DBStorage:time cannot be parsed",
}

type dbUsers struct {
	user_id    sql.NullString
	password   sql.NullString
	accural    sql.NullFloat64
	withdrawal sql.NullFloat64
}

type dbOrders struct {
	order_id   sql.NullInt64
	user_id    sql.NullString
	status     sql.NullInt64
	accrual    sql.NullFloat64
	created_at sql.NullString
}

type dbWithdrawals struct {
	user_id    sql.NullString
	created_at sql.NullString
	withdrawal sql.NullFloat64
}

type DBStorage struct {
	dataBaseURL string
	pool        *pgxpool.Pool
	conn        *pgxpool.Conn
}

func createTable(ctx context.Context, s DBStorage, checkSql string, createSql string) error {

	resp, err := s.pool.Exec(ctx, checkSql)
	if err != nil {
		log.Println(message[2] + err.Error())
		//create Table
		resp, err = s.pool.Exec(ctx, createSql)
		if err != nil {
			log.Fatal(err)
		}
		log.Println(message[1] + resp.String())
	} else {
		log.Println(message[2] + resp.String())
	}

	return err
}

func NewDBStorage(ctx context.Context, dataBaseURL string) *DBStorage {
	//get params
	s := DBStorage{dataBaseURL: dataBaseURL}
	//connect db
	var err error
	//s.conn, err = pgx.Connect(ctx, s.dataBaseURL)
	s.pool, err = pgxpool.New(ctx, s.dataBaseURL)
	if err != nil {
		logFatalf(message[0], err)
		return nil
	}
	// check users table exists
	err = createTable(ctx, s, checkIfUsersTableExists, createUsersTable)
	logFatalf("error:", err)
	// check orders table exists
	err = createTable(ctx, s, checkIfOrdersTableExists, createOrdersTable)
	logFatalf("error:", err)
	// check withdrawals table exists
	err = createTable(ctx, s, checkIfWithdrawalsTableExists, createWithdrawalsTable)
	logFatalf("error:", err)

	return &s
}

func logFatalf(mess string, err error) {
	if err != nil {
		log.Fatalf(mess+": %v\n", err)
	}
}
func (s *DBStorage) connectDB(ctx context.Context) (ok bool) {
	ok = false
	var err error

	if s.pool == nil {
		s.pool, err = pgxpool.New(ctx, s.dataBaseURL)
		logFatalf(message[0], err)
	}
	for i := 0; i < 10; i++ {
		s.conn, err = s.pool.Acquire(ctx)
		if err != nil {
			log.Println(message[12] + " " + err.Error())
			time.Sleep(time.Millisecond * 200)
			continue
		}
		break
	}

	err = s.conn.Ping(ctx)
	if err != nil {
		logFatalf(message[0], err)
	}

	ok = true
	return ok
}

//GetUser(ctx context.Context, name string) (u *schema.User, err error)
//SaveUser(ctx context.Context, u *schema.User) (err error)
//
//GetOrder(ctx context.Context, orderNumber int) (o *schema.Order, err error)
//SaveOrder(ctx context.Context, o schema.Order) (err error)
//GetOrdersList(ctx context.Context, userName string) (wl schema.Orders, err error)
//
//SaveWithdrawal(ctx context.Context, w schema.Withdrawal) (err error)
//GetWithdrawalsList(ctx context.Context, u schema.User) (wl *schema.Withdrawals, err error)

func (s DBStorage) GetUser(ctx context.Context, name string) (u *schema.User, err error) {
	if !s.connectDB(ctx) {
		return nil, errors.New(message[0])
	}
	defer s.conn.Release()
	d := dbUsers{user_id: sql.NullString{String: name, Valid: true}}
	row := s.conn.QueryRow(ctx, selectLineUsersTable, &d.user_id)
	err = row.Scan(&d.user_id, &d.password, &d.accural, &d.withdrawal)
	if err != nil {
		log.Printf("QueryRow failed: %v\n", err)
		return nil, err
	}
	return &schema.User{
		User:       d.user_id.String,
		Password:   d.password.String,
		Accrual:    d.accural.Float64,
		Withdrawal: d.withdrawal.Float64,
	}, nil
}

func (s DBStorage) SaveUser(ctx context.Context, u *schema.User) (err error) {
	if !s.connectDB(ctx) {
		return errors.New(message[0])
	}
	defer s.conn.Release()

	d := dbUsers{
		user_id:    sql.NullString{String: u.User, Valid: true},
		password:   sql.NullString{String: u.Password, Valid: true},
		accural:    sql.NullFloat64{Float64: u.Accrual, Valid: true},
		withdrawal: sql.NullFloat64{Float64: u.Withdrawal, Valid: true},
	}

	tag, err := s.conn.Exec(ctx, createOrUpdateIfExistsUsersTable, d.user_id, d.password, d.accural, d.withdrawal)
	logFatalf(message[3], err)
	log.Println(tag)
	return err
}

func (s DBStorage) GetOrder(ctx context.Context, orderNumber int64) (o *schema.Order, err error) {
	if !s.connectDB(ctx) {
		return nil, errors.New(message[0])
	}
	defer s.conn.Release()
	d := dbOrders{order_id: sql.NullInt64{Int64: orderNumber, Valid: true}}
	row := s.conn.QueryRow(ctx, selectLineOrdersTable, &d.order_id)
	err = row.Scan(&d.order_id, &d.user_id, &d.status, &d.accrual, &d.created_at)
	if err != nil {
		log.Printf("QueryRow failed: %v\n", err)
		return nil, err
	}
	created, err := time.Parse(time.RFC3339, d.created_at.String)
	return &schema.Order{
		Order:   d.order_id.Int64,
		User:    d.user_id.String,
		Status:  d.status.Int64,
		Accrual: d.accrual.Float64,
		Created: schema.CreatedTime(created),
	}, nil
}
func (s DBStorage) SaveOrder(ctx context.Context, o schema.Order) (err error) {
	if !s.connectDB(ctx) {
		return errors.New(message[0])
	}
	d := &dbOrders{
		order_id:   sql.NullInt64{Int64: o.Order, Valid: true},
		user_id:    sql.NullString{String: o.User, Valid: true},
		accrual:    sql.NullFloat64{Float64: o.Accrual, Valid: true},
		created_at: sql.NullString{String: time.Time(o.Created).Format(time.RFC3339), Valid: true},
	}

	tag, err := s.conn.Exec(ctx, createOrUpdateIfExistsOrdersTable, d.order_id, d.user_id, d.accrual, d.created_at)
	logFatalf(message[3], err)
	log.Println(tag)
	return err
}

func (s DBStorage) GetOrdersList(ctx context.Context, userName string) (wl schema.Orders, err error) {
	if !s.connectDB(ctx) {
		return nil, errors.New(message[0])
	}
	defer s.conn.Release()

	wl = make(schema.Orders)

	d := &dbOrders{user_id: sql.NullString{String: userName, Valid: true}}

	rows, err := s.conn.Query(ctx, selectAllOrdersTableByUser, d.user_id)
	if err != nil {
		log.Printf(message[4], err)
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(d.order_id, d.user_id, d.accrual, d.created_at)
		logFatalf(message[5], err)
		created, err := time.Parse(time.RFC3339, d.created_at.String)
		logFatalf(message[6], err)
		wl[d.order_id.Int64] = schema.Order{
			Order:   d.order_id.Int64,
			User:    d.user_id.String,
			Accrual: d.accrual.Float64,
			Created: schema.CreatedTime(created),
		}
	}

	return wl, nil
}

func (s DBStorage) GetNewOrdersList(ctx context.Context) (ol schema.Orders, err error) {
	if !s.connectDB(ctx) {
		return nil, errors.New(message[0])
	}
	defer s.conn.Release()

	ol = make(schema.Orders)

	d := &dbOrders{status: sql.NullInt64{Int64: schema.OrderStatus["NEW"], Valid: true}}

	rows, err := s.conn.Query(ctx, selectAllOrdersTableByStatus, d.status)
	if err != nil {
		log.Printf(message[4], err)
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(d.order_id, d.user_id, d.accrual, d.created_at)
		logFatalf(message[5], err)
		created, err := time.Parse(time.RFC3339, d.created_at.String)
		logFatalf(message[6], err)
		ol[d.order_id.Int64] = schema.Order{
			Order:   d.order_id.Int64,
			User:    d.user_id.String,
			Accrual: d.accrual.Float64,
			Created: schema.CreatedTime(created),
		}
	}

	return ol, nil
}

func (s DBStorage) SaveWithdrawal(ctx context.Context, w schema.Withdrawal) (err error) {

	if !s.connectDB(ctx) {
		return errors.New(message[0])
	}
	defer s.conn.Release()

	d := dbWithdrawals{
		user_id:    sql.NullString{String: w.User, Valid: true},
		created_at: sql.NullString{String: time.Time(w.Processed).Format(time.RFC3339), Valid: true},
		withdrawal: sql.NullFloat64{Float64: w.Withdrawal, Valid: true},
	}
	tag, err := s.conn.Exec(ctx, createOrUpdateIfExistsWithdrawalsTable, d.user_id, d.created_at, d.withdrawal)
	logFatalf(message[3], err)
	log.Println(tag)
	return err
}
func (s DBStorage) GetWithdrawalsList(ctx context.Context, username string) (wl *schema.Withdrawals, err error) {
	if !s.connectDB(ctx) {
		return nil, errors.New(message[0])
	}
	defer s.conn.Release()

	wl = new(schema.Withdrawals)

	d := &dbWithdrawals{user_id: sql.NullString{String: username, Valid: true}}

	rows, err := s.conn.Query(ctx, selectAllWithdrawalsTableByUser, d.user_id)
	if err != nil {
		log.Printf(message[4], err)
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(d.user_id, d.created_at, d.withdrawal)
		logFatalf(message[5], err)
		created, err := time.Parse(time.RFC3339, d.created_at.String)
		logFatalf(message[6], err)

		w := schema.Withdrawal{
			User:       d.user_id.String,
			Processed:  schema.CreatedTime(created),
			Withdrawal: d.withdrawal.Float64,
		}
		*wl = append(*wl, w)
	}

	return wl, nil
}
