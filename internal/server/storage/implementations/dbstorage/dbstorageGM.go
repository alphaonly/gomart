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

//-d=postgres://postgres:mypassword@localhost:5432/yandexxx

const selectLineUsersTable = `SELECT user_id, password, accural, withdrawal FROM public.users WHERE user_id=$1;`

// const selectAllUsersTable = `SELECT user_id,balance,withdrawn FROM public.users;`
const selectAllOrdersTableByUser = `SELECT order_id,user_id,accural, withdrawal FROM public.orders WHERE user_id = $1;`
const selectAllWithdrawalsTableByUser = `SELECT user_id, uploaded_at, withdrawal FROM public.withdrawals WHERE user_id = $1;`

const createOrUpdateIfExistsUsersTable = `
	INSERT INTO public.users (user_id, password, accural, withdrawal) 
	VALUES ($1, $2, $3, $4)
	ON CONFLICT (user_id) DO UPDATE 
  	SET password 	= $2,
	  	accural 	= $3,
		withdrawal 	= $4; 
  	`
const createOrUpdateIfExistsOrdersTable = `
	  INSERT INTO public.orders (order_id, user_id, status,accural) 
	  VALUES ($1, $2, $3)
	  ON CONFLICT (order_id) DO UPDATE 
		SET password = $2,
			accural = $3; 
		`
const createOrUpdateIfExistsWithdrawalsTable = `
		INSERT INTO public.withdrawals (user_id, uploaded_at, withdrawal) 
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id,uploaded_at) DO UPDATE 
		  SET accural = $2,
			withdrawn = $3; 
		  `

const createUsersTable = `create table public.users
	(	user_id varchar(40) not null primary key,
		password  TEXT not null,
		accural integer,
		withdrawal integer 
	);`

const createOrdersTable = `create table public.orders
	(	order_id integer not null primary key,
		user_id varchar(40) not null,
		status integer,
		sum double precision,
		accural integer,
		uploaded_at TEXT not null 
	);`

const createWithdrawalsTable = `create table public.withdrawals
	(	user_id 		varchar(40) primary key,
		uploaded_at 	TEXT 		unique not null,
		withdrawal 		integer 	not null	
	);`

const checkIfUsersTableExists = `SELECT 'public.users'::regclass;`
const checkIfOrdersTableExists = `SELECT 'public.users'::regclass;`
const checkIfWithdrawalsTableExists = `SELECT 'public.withdrawals'::regclass;`

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
	accural    sql.NullInt64
	withdrawal sql.NullInt64
}

type dbOrders struct {
	order_id   sql.NullInt64
	user_id    sql.NullString
	accural    sql.NullInt64
	created_at sql.NullString
}

type dbWithdrawals struct {
	user_id    sql.NullString
	created_at sql.NullString
	withdrawal sql.NullInt64
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

// type Storage interface {
// 	GetUser(ctx context.Context, name string) (u *schema.User, err error)
// 	SaveUser(ctx context.Context, u schema.User) (err error)

// 	SaveOrder(ctx context.Context, o schema.Order) (err error)
// 	GetOrdersList(ctx context.Context, u schema.User) (wl schema.Orders, err error)

// 	SaveWithdrawal(ctx context.Context, w schema.Withdrawal) (err error)
// 	GetWithdrawalsList(ctx context.Context, u schema.User) (wl schema.Withdrawals, err error)
// }

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
		Accural:    d.accural.Int64,
		Withdrawal: d.withdrawal.Int64,
	}, nil
}

func (s DBStorage) SaveUser(ctx context.Context, u schema.User) (err error) {
	if !s.connectDB(ctx) {
		return errors.New(message[0])
	}
	defer s.conn.Release()

	d := dbUsers{
		user_id:    sql.NullString{String: u.User, Valid: true},
		password:   sql.NullString{String: u.Password, Valid: true},
		accural:    sql.NullInt64{Int64: u.Accural, Valid: true},
		withdrawal: sql.NullInt64{Int64: u.Withdrawal, Valid: true},
	}

	tag, err := s.conn.Exec(ctx, createOrUpdateIfExistsUsersTable, d.user_id, d.password, d.accural, d.withdrawal)
	logFatalf(message[3], err)
	log.Println(tag)
	return err
}

func (s DBStorage) GetOrdersList(ctx context.Context, u schema.User) (ol schema.Orders, err error) {
	if !s.connectDB(ctx) {
		return nil, errors.New(message[0])
	}
	defer s.conn.Release()

	ol = make(schema.Orders)

	d := &dbOrders{user_id: sql.NullString{String: u.User, Valid: true}}

	rows, err := s.conn.Query(ctx, selectAllOrdersTableByUser, d.user_id)
	if err != nil {
		log.Printf(message[4], err)
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(d.order_id, d.user_id, d.accural, d.created_at)
		logFatalf(message[5], err)
		created, err := time.Parse(time.RFC3339, d.created_at.String)
		logFatalf(message[6], err)
		ol[d.order_id.Int64] = schema.Order{
			Order:   d.order_id.Int64,
			User:    d.user_id.String,
			Accural: d.accural.Int64,
			Created: created,
		}
	}

	return ol, nil
}

func (s DBStorage) SaveOrder(ctx context.Context, o schema.Order) (err error) {
	if !s.connectDB(ctx) {
		return errors.New(message[0])
	}
	d := &dbOrders{
		order_id:   sql.NullInt64{Int64: o.Order, Valid: true},
		user_id:    sql.NullString{String: o.User, Valid: true},
		accural:    sql.NullInt64{Int64: o.Accural, Valid: true},
		created_at: sql.NullString{String: o.Created.Format(time.RFC3339), Valid: true},
	}

	tag, err := s.conn.Exec(ctx, createOrUpdateIfExistsOrdersTable, d.order_id, d.user_id, d.accural, d.created_at)
	logFatalf(message[3], err)
	log.Println(tag)
	return err
}

func (s DBStorage) SaveWithdrawal(ctx context.Context, w schema.Withdrawal) (err error) {

	if !s.connectDB(ctx) {
		return errors.New(message[0])
	}
	defer s.conn.Release()

	d := dbWithdrawals{
		user_id:    sql.NullString{String: w.User, Valid: true},
		created_at: sql.NullString{String: w.Created.Format(time.RFC3339), Valid: true},
		withdrawal: sql.NullInt64{Int64: w.Withdrawal, Valid: true},
	}
	tag, err := s.conn.Exec(ctx, createOrUpdateIfExistsWithdrawalsTable, d.user_id, d.created_at, d.withdrawal)
	logFatalf(message[3], err)
	log.Println(tag)
	return err
}
func (s DBStorage) GetWithdrawalsList(ctx context.Context, u schema.User) (wl *schema.Withdrawals, err error) {
	if !s.connectDB(ctx) {
		return nil, errors.New(message[0])
	}
	defer s.conn.Release()

	wl = new(schema.Withdrawals)

	d := &dbWithdrawals{user_id: sql.NullString{String: u.User, Valid: true}}

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
			Created:    created,
			Withdrawal: d.withdrawal.Int64,
		}
		*wl = append(*wl, w)
	}

	return wl, nil
}
