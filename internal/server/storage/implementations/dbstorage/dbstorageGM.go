package storage

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"time"

	metricsjson "github.com/alphaonly/gomart/internal/server/metricsJSON"
	mVal "github.com/alphaonly/gomart/internal/server/metricvaluei"
	storage "github.com/alphaonly/gomart/internal/server/storage/interfaces"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

//	type Storage interface {
//		GetMetric(ctx context.Context, name string) (mv *M.MetricValue, err error)
//		SaveMetric(ctx context.Context, name string, mv *M.MetricValue) (err error)
//		GetAllMetrics(ctx context.Context) (mvList *metricsjson.MetricsMapType, err error)
//		SaveAllMetrics(ctx context.Context, mvList *metricsjson.MetricsMapType) (err error)
//	}

//-d=postgres://postgres:mypassword@localhost:5432/yandexxx

const selectLineUsersTable = `SELECT user_id, password, accural, withdrawal FROM public.users WHERE id=$1;`
const selectAllUsersTable = `SELECT user_id,balance,withdrawn FROM public.users;`
const selectAllWithdrawalsTable = `SELECT user_id, uploaded_at, withdrawal FROM public.withdrawals;`

const createOrUpdateIfExistsUsersTable = `
	INSERT INTO public.users (user_id, password, accural, withdrawal) 
	VALUES ($1, $2, $3, $4)
	ON CONFLICT (id) DO UPDATE 
  	SET password = $2,
	  	accural = $3,
		withdrawal = $4; 
  	`
const createOrUpdateIfExistsOrdersTable = `
	  INSERT INTO public.orders (order_id, user_id, status,accural) 
	  VALUES ($1, $2, $3)
	  ON CONFLICT (id) DO UPDATE 
		SET password = $2,
			accural = $3; 
		`
const createOrUpdateIfExistsWithdrawalsTable = `
		INSERT INTO public.withdrawals (user_id, uploaded_at, withdrawal) 
		VALUES ($1, $2, $3)
		ON CONFLICT (id) DO UPDATE 
		  SET accural = $2,
			withdrawn = $3; 
		  `

const createUsersTable = `create table public.users
	(	user_id varchar(40) not null primary key,
		password  TEXT not null,
		accural integer
		withdrawal integer 
	);`

const createOrdersTable = `create table public.orders
	(	order_id integer not null primary key,
		user_id varchar(40) not null,
		status integer
		sum double precision
		accural integer
		uploaded_at TEXT not null 
	);`

const createWithdrawalsTable = `create table public.withdrawals
	(	user_id varchar(40) not null primary key,
		uploaded_at TEXT not null primary key,
		withdrawal integer not null	
	);`

const checkIfUsersTableExists = `SELECT 'public.users'::regclass;`
const checkIfOrdersTableExists = `SELECT 'public.users'::regclass;`
const checkIfWithdrawalsTableExists = `SELECT 'public.withdrawals'::regclass;`

var message = []string{
	0: "DBStorage:unable to connect to database",
	1: "DBStorage:%v table has created",
	2: "DBStorage:unable to create %v table",
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

func createTable(ctx context.Context, s DBStorage, sql string, tableName string) error {

	resp, err := s.pool.Exec(ctx, checkIfUsersTableExists)
	if err != nil {
		log.Println(message[1] + err.Error())
		//create metrics Table
		resp, err = s.pool.Exec(ctx, createUsersTable)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf(message[1]+resp.String(), tableName)
	} else {
		log.Printf(message[2]+resp.String(), tableName)
	}

	return err
}

func NewDBStorage(ctx context.Context, dataBaseURL string) *storage.Storage {
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
	err = createTable(ctx, s, createOrUpdateIfExistsUsersTable, "Users")
	logFatalf("error:", err)
	// check orders table exists
	err = createTable(ctx, s, createOrUpdateIfExistsOrdersTable, "Orders")
	logFatalf("error:", err)
	// check withdrawals table exists
	err = createTable(ctx, s, createOrUpdateIfExistsUsersTable, "Withdrawals")
	logFatalf("error:", err)

	return s
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



func (s DBStorage) GetUser(ctx context.Context, name string) (u *User, err error) {
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

	return &User{
		user:       d.user_id.String,
		password:   d.password.String,
		accural:    d.accural.Int64,
		withdrawal: d.withdrawal.Int64,
	}, nil
}
func (s DBStorage) SaveMetric(ctx context.Context, name string, mv *mVal.MetricValue) (err error) {
	var m mVal.MetricValue
	if mv == nil {
		return errors.New(message[6])
	}
	m = *mv
	if !s.connectDB(ctx) {
		return errors.New(message[14])
	}
	defer s.conn.Release()

	var (
		_type int
		delta int64
		value float64
	)

	switch v := m.GetInternalValue().(type) {
	case int64:
		{
			_type = 1
			delta = v
		}
	case float64:
		{
			_type = 2
			value = v
		}
	default:
		return errors.New(message[7])
	}
	tag, err := s.conn.Exec(ctx, createOrUpdateIfExistsMetricsTable, name, _type, delta, value)
	logFatalf("", err)
	log.Println(tag)
	return err
}

// GetAllMetrics Restore data from database to mem storage
func (s DBStorage) GetAllMetrics(ctx context.Context) (mvList *metricsjson.MetricsMapType, err error) {
	if !s.connectDB(ctx) {
		return nil, errors.New(message[14])
	}
	defer s.conn.Release()
	rows, err := s.conn.Query(ctx, selectAllMetricsTable)
	if err != nil {
		log.Printf("QueryRow failed: %v\n", err)
		return nil, err
	}
	defer rows.Close()
	m := make(metricsjson.MetricsMapType)
	emptyList := make(metricsjson.MetricsMapType)
	for rows.Next() {
		d := dbMetrics{}
		err = rows.Scan(&d.id, &d._type, &d.delta, &d.value)
		if err != nil {
			return nil, err
		}
		var mv mVal.MetricValue
		switch d._type.Int64 {
		case 1:
			{
				if !d.delta.Valid {
					return &emptyList, errors.New(message[5])
				}
				mv = mVal.NewInt(d.delta.Int64)
			}
		case 2:
			{
				if !d.value.Valid {
					return &emptyList, errors.New(message[5])
				}
				mv = mVal.NewFloat(d.value.Float64)
			}
		default:
			log.Fatalf(message[4])
		}
		if !d.id.Valid {
			return &emptyList, errors.New(message[5])
		}
		m[d.id.String] = mv
	}
	return &m, nil
}

// SaveAllMetrics Park data to database
func (s DBStorage) SaveAllMetrics(ctx context.Context, mvList *metricsjson.MetricsMapType) (err error) {
	log.Println("DBStorage SaveAllMetrics invoked")
	if mvList == nil {
		return errors.New(message[6])
	}
	if !s.connectDB(ctx) {
		return errors.New(message[14])
	}
	defer s.conn.Release()

	mv := *mvList

	batch := &pgx.Batch{}
	for k, v := range mv {
		var d dbMetrics
		switch value := v.(type) {
		case *mVal.GaugeValue:
			d = dbMetrics{
				id:    sql.NullString{String: k, Valid: true},
				_type: sql.NullInt64{Int64: 2, Valid: true},
				value: sql.NullFloat64{Float64: value.GetInternalValue().(float64), Valid: true},
				delta: sql.NullInt64{},
			}
		case *mVal.CounterValue:
			d = dbMetrics{
				id:    sql.NullString{String: k, Valid: true},
				_type: sql.NullInt64{Int64: 1, Valid: true},
				value: sql.NullFloat64{},
				delta: sql.NullInt64{Int64: value.GetInternalValue().(int64), Valid: true},
			}
		default:
			return errors.New(message[7])
		}

		batch.Queue(createOrUpdateIfExistsMetricsTable, d.id, d._type, d.delta, d.value)
	}

	br := s.conn.SendBatch(ctx, batch)
	for range mv {
		tag, err := br.Exec()
		if err != nil {
			logFatalf(message[9], err)
			return err
		}
		log.Println(message[8] + tag.String())
	}
	defer br.Close()

	return nil
}
