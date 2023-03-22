package main

import (
	"context"
	"log"

	metricvalueI "github.com/alphaonly/gomart/internal/server/metricvaluei"
	db "github.com/alphaonly/gomart/internal/server/storage/implementations/dbstorage"
)

var urlExample = "postgres://postgres:mypassword@localhost:5432/yandex"

const createMetricsTable = `create table public.metrics
(	id varchar(40) not null primary key,
	type integer not null,
	delta integer,
	value double precision
);`

const checkIfMetricsTableExists = `SELECT 'public.metrics'::regclass;`

const insertLineIntoMetricsTable = `
	INSERT INTO public.metrics (id, type, delta, value)VALUES ($1, $2, $3, $4);`

func main() {

	s := db.NewDBStorage(context.Background(), urlExample)
	var mv metricvalueI.MetricValue = metricvalueI.NewInt(123)

	s.SaveMetric(context.Background(), "PollCounter", &mv)
	log.Println(s)
}
