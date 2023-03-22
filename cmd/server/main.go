package main

import (
	"context"
	"log"

	db "github.com/alphaonly/gomart/internal/server/storage/implementations/dbstorage"
	stor "github.com/alphaonly/gomart/internal/server/storage/interfaces"

	conf "github.com/alphaonly/gomart/internal/configuration"
	"github.com/alphaonly/gomart/internal/server"
	"github.com/alphaonly/gomart/internal/server/handlers"
	"github.com/alphaonly/gomart/internal/server/storage/implementations/mapstorage"
	"github.com/alphaonly/gomart/internal/signchecker"
)

func main() {

	configuration := conf.NewServerConf(conf.UpdateSCFromEnvironment, conf.UpdateSCFromFlags)
	// configuration.UpdateFromEnvironment()
	// configuration.UpdateFromFlags()

	var (
		externalStorage stor.Storage
		internalStorage stor.Storage
	)
	externalStorage = nil
	internalStorage = mapstorage.New()

	if configuration.DatabaseDsn != "" {
		externalStorage = nil
		internalStorage = db.NewDBStorage(context.Background(), configuration.DatabaseDsn)
	}

	handlers := &handlers.Handlers{
		Storage: internalStorage,
		Signer:  signchecker.NewSHA256(configuration.Key),
		Conf:    conf.ServerConfiguration{DatabaseDsn: configuration.DatabaseDsn},
	}

	metricsServer := server.New(configuration, externalStorage, handlers)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := metricsServer.Run(ctx)
	if err != nil {
		log.Fatal(err)
	}

}
