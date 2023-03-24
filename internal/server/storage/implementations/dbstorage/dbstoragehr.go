package storage

// import (
// 	"context"
// 	"database/sql"
// 	"errors"
// 	"log"
// 	"time"

// 	metricsjson "github.com/alphaonly/gomart/internal/server/metricsJSON"
// 	mVal "github.com/alphaonly/gomart/internal/server/metricvaluei"
// 	storage "github.com/alphaonly/gomart/internal/server/storage/interfaces"
// 	"github.com/jackc/pgx/v5"
// 	"github.com/jackc/pgx/v5/pgxpool"
// )

// //	type Storage interface {
// //		GetMetric(ctx context.Context, name string) (mv *M.MetricValue, err error)
// //		SaveMetric(ctx context.Context, name string, mv *M.MetricValue) (err error)
// //		GetAllMetrics(ctx context.Context) (mvList *metricsjson.MetricsMapType, err error)
// //		SaveAllMetrics(ctx context.Context, mvList *metricsjson.MetricsMapType) (err error)
// //	}

// //-d=postgres://postgres:mypassword@localhost:5432/yandexxx

// const selectLineMetricsTable = `SELECT id,type,delta,value FROM public.metrics2 WHERE id=$1;`
// const selectAllMetricsTable = `SELECT id,type,delta,value FROM public.metrics2;`

// const createOrUpdateIfExistsMetricsTable = `
// 	INSERT INTO public.metrics2 (id, type, delta,value) 
// 	VALUES ($1, $2, $3, $4)
// 	ON CONFLICT (id) DO UPDATE 
//   	SET delta = $3, 
//       	value = $4;`

// const createMetricsTable = `create table public.metrics2
// 	(	id varchar(40) not null primary key,
// 		type integer not null,
// 		delta bigint,
// 		value double precision
// 	);`

// const checkIfMetricsTableExists = `SELECT 'public.metrics2'::regclass;`

// var message = []string{
// 	0:  "DBStorage:unable to connect to database",
// 	1:  "DBStorage:table metrics does not exists, a try to create:",
// 	2:  "DBStorage:db metrics table creation response text:",
// 	3:  "DBStorage:db metrics table existence check response text:",
// 	4:  "DBStorage:getMetrics: Unknown metric type value",
// 	5:  "DBStorage:NullValue type not valid",
// 	6:  "DBStorage:nil pointer in mvList",
// 	7:  "DBStorage:undefined type in type switch dbStorage",
// 	8:  "DBStorage:sendBatch tag:",
// 	9:  "DBStorage:unable to rollback, error fatal",
// 	10: "DBStorage:unable to commit, trying again",
// 	11: "DBStorage:unable to prepare statement, fatal error",
// 	12: "DBStorage:unable to lease a connection , try again",
// 	13: "DBStorage:createOrUpdateIfExistsMetricsTable exec error",
// 	14: "DBStorage:unable to lease a connection after a few tries",
// }

// type dbMetrics struct {
// 	id    sql.NullString
// 	_type sql.NullInt64
// 	delta sql.NullInt64
// 	value sql.NullFloat64
// }

// type DBStorage struct {
// 	dataBaseURL string
// 	pool        *pgxpool.Pool
// 	conn        *pgxpool.Conn
// }

// func NewDBStorage(ctx context.Context, dataBaseURL string) storage.Storage {
// 	//get params
// 	s := DBStorage{dataBaseURL: dataBaseURL}
// 	//connect db
// 	var err error
// 	//s.conn, err = pgx.Connect(ctx, s.dataBaseURL)
// 	s.pool, err = pgxpool.New(ctx, s.dataBaseURL)
// 	if err != nil {
// 		logFatalf(message[0], err)
// 		return nil
// 	}
// 	// check metrics table exists
// 	resp, err := s.pool.Exec(context.Background(), checkIfMetricsTableExists)
// 	if err != nil {
// 		log.Println(message[1] + err.Error())
// 		//create metrics Table
// 		resp, err = s.pool.Exec(context.Background(), createMetricsTable)
// 		if err != nil {
// 			log.Fatal(err)
// 		}
// 		log.Println(message[2] + resp.String())
// 	} else {
// 		log.Println(message[3] + resp.String())
// 	}
// 	return s
// }

// func logFatalf(mess string, err error) {
// 	if err != nil {
// 		log.Fatalf(mess+": %v\n", err)
// 	}
// }
// func (s *DBStorage) connectDB(ctx context.Context) (ok bool) {
// 	ok = false
// 	var err error

// 	if s.pool == nil {
// 		s.pool, err = pgxpool.New(ctx, s.dataBaseURL)
// 		logFatalf(message[0], err)
// 	}
// 	for i := 0; i < 10; i++ {
// 		s.conn, err = s.pool.Acquire(ctx)
// 		if err != nil {
// 			log.Println(message[12] + " " + err.Error())
// 			time.Sleep(time.Millisecond * 200)
// 			continue
// 		}
// 		break
// 	}

// 	err = s.conn.Ping(ctx)
// 	if err != nil {
// 		logFatalf(message[0], err)
// 	}

// 	ok = true
// 	return ok
// }

// func (s DBStorage) GetMetric(ctx context.Context, name string, MType string) (mv mVal.MetricValue, err error) {
// 	if !s.connectDB(ctx) {
// 		return nil, errors.New(message[14])
// 	}
// 	defer s.conn.Release()

// 	d := dbMetrics{id: sql.NullString{String: name, Valid: true}}

// 	switch MType {
// 	case "counter":
// 		d._type = sql.NullInt64{Int64: 1, Valid: true}
// 	case "gauge":
// 		d._type = sql.NullInt64{Int64: 2, Valid: true}
// 	default:
// 		log.Fatalf(message[4])
// 	}

// 	row := s.conn.QueryRow(ctx, selectLineMetricsTable, &d.id)

// 	err = row.Scan(&d.id, &d._type, &d.delta, &d.value)
// 	if err != nil {
// 		log.Printf("QueryRow failed: %v\n", err)
// 		return nil, err
// 	}

// 	switch d._type.Int64 {
// 	case 1:
// 		{
// 			if d.delta.Valid {
// 				mv = mVal.NewInt(d.delta.Int64)
// 			}
// 		}
// 	case 2:
// 		{
// 			if d.value.Valid {
// 				mv = mVal.NewFloat(d.value.Float64)
// 			}
// 		}
// 	default:
// 		log.Fatalf(message[4])
// 	}
// 	return mv, nil
// }
// func (s DBStorage) SaveMetric(ctx context.Context, name string, mv *mVal.MetricValue) (err error) {
// 	var m mVal.MetricValue
// 	if mv == nil {
// 		return errors.New(message[6])
// 	}
// 	m = *mv
// 	if !s.connectDB(ctx) {
// 		return errors.New(message[14])
// 	}
// 	defer s.conn.Release()

// 	var (
// 		_type int
// 		delta int64
// 		value float64
// 	)

// 	switch v := m.GetInternalValue().(type) {
// 	case int64:
// 		{
// 			_type = 1
// 			delta = v
// 		}
// 	case float64:
// 		{
// 			_type = 2
// 			value = v
// 		}
// 	default:
// 		return errors.New(message[7])
// 	}
// 	tag, err := s.conn.Exec(ctx, createOrUpdateIfExistsMetricsTable, name, _type, delta, value)
// 	logFatalf("", err)
// 	log.Println(tag)
// 	return err
// }

// // GetAllMetrics Restore data from database to mem storage
// func (s DBStorage) GetAllMetrics(ctx context.Context) (mvList *metricsjson.MetricsMapType, err error) {
// 	if !s.connectDB(ctx) {
// 		return nil, errors.New(message[14])
// 	}
// 	defer s.conn.Release()
// 	rows, err := s.conn.Query(ctx, selectAllMetricsTable)
// 	if err != nil {
// 		log.Printf("QueryRow failed: %v\n", err)
// 		return nil, err
// 	}
// 	defer rows.Close()
// 	m := make(metricsjson.MetricsMapType)
// 	emptyList := make(metricsjson.MetricsMapType)
// 	for rows.Next() {
// 		d := dbMetrics{}
// 		err = rows.Scan(&d.id, &d._type, &d.delta, &d.value)
// 		if err != nil {
// 			return nil, err
// 		}
// 		var mv mVal.MetricValue
// 		switch d._type.Int64 {
// 		case 1:
// 			{
// 				if !d.delta.Valid {
// 					return &emptyList, errors.New(message[5])
// 				}
// 				mv = mVal.NewInt(d.delta.Int64)
// 			}
// 		case 2:
// 			{
// 				if !d.value.Valid {
// 					return &emptyList, errors.New(message[5])
// 				}
// 				mv = mVal.NewFloat(d.value.Float64)
// 			}
// 		default:
// 			log.Fatalf(message[4])
// 		}
// 		if !d.id.Valid {
// 			return &emptyList, errors.New(message[5])
// 		}
// 		m[d.id.String] = mv
// 	}
// 	return &m, nil
// }

// // SaveAllMetrics Park data to database
// func (s DBStorage) SaveAllMetrics(ctx context.Context, mvList *metricsjson.MetricsMapType) (err error) {
// 	log.Println("DBStorage SaveAllMetrics invoked")
// 	if mvList == nil {
// 		return errors.New(message[6])
// 	}
// 	if !s.connectDB(ctx) {
// 		return errors.New(message[14])
// 	}
// 	defer s.conn.Release()

// 	mv := *mvList

// 	batch := &pgx.Batch{}
// 	for k, v := range mv {
// 		var d dbMetrics
// 		switch value := v.(type) {
// 		case *mVal.GaugeValue:
// 			d = dbMetrics{
// 				id:    sql.NullString{String: k, Valid: true},
// 				_type: sql.NullInt64{Int64: 2, Valid: true},
// 				value: sql.NullFloat64{Float64: value.GetInternalValue().(float64), Valid: true},
// 				delta: sql.NullInt64{},
// 			}
// 		case *mVal.CounterValue:
// 			d = dbMetrics{
// 				id:    sql.NullString{String: k, Valid: true},
// 				_type: sql.NullInt64{Int64: 1, Valid: true},
// 				value: sql.NullFloat64{},
// 				delta: sql.NullInt64{Int64: value.GetInternalValue().(int64), Valid: true},
// 			}
// 		default:
// 			return errors.New(message[7])
// 		}

// 		batch.Queue(createOrUpdateIfExistsMetricsTable, d.id, d._type, d.delta, d.value)
// 	}

// 	br := s.conn.SendBatch(ctx, batch)
// 	for range mv {
// 		tag, err := br.Exec()
// 		if err != nil {
// 			logFatalf(message[9], err)
// 			return err
// 		}
// 		log.Println(message[8] + tag.String())
// 	}
// 	defer br.Close()

// 	return nil
// }
