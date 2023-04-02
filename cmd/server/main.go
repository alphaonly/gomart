package main

import (
	"context"
	"github.com/alphaonly/gomart/internal/server/accrual"
	"log"

	db "github.com/alphaonly/gomart/internal/server/storage/implementations/dbstorage"
	stor "github.com/alphaonly/gomart/internal/server/storage/interfaces"

	conf "github.com/alphaonly/gomart/internal/configuration"
	"github.com/alphaonly/gomart/internal/server"
	"github.com/alphaonly/gomart/internal/server/handlers"
)

func main() {

	configuration := conf.NewServerConf(conf.UpdateSCFromEnvironment, conf.UpdateSCFromFlags)

	var (
		externalStorage stor.Storage
		internalStorage stor.Storage
	)

	externalStorage = nil
	internalStorage = db.NewDBStorage(context.Background(), configuration.DatabaseURI)

	handlers := &handlers.Handlers{
		Storage:       internalStorage,
		Conf:          conf.ServerConfiguration{DatabaseURI: configuration.DatabaseURI},
		EntityHandler: &handlers.EntityHandler{Storage: internalStorage},
	}
	accrualChecker := accrual.NewChecker(configuration.AccrualSystemAddress, configuration.AccrualTime, internalStorage)

	gmServer := server.New(configuration, externalStorage, handlers, accrualChecker)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := gmServer.Run(ctx)
	if err != nil {
		log.Fatal(err)
	}

}
