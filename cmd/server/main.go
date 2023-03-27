package main

import (
	"context"
	"log"

	db "github.com/alphaonly/gomart/internal/server/storage/implementations/dbstorage"
	stor "github.com/alphaonly/gomart/internal/server/storage/interfaces"

	conf "github.com/alphaonly/gomart/internal/configuration"
	"github.com/alphaonly/gomart/internal/server"
	"github.com/alphaonly/gomart/internal/server/handlers"
)

func main() {

	configuration := conf.NewServerConf(conf.UpdateSCFromEnvironment, conf.UpdateSCFromFlags)
	// configuration.UpdateFromEnvironment()
	// configuration.UpdateFromFlags()

	var (
		externalStorage stor.Storage
		internalStorage stor.Storage
	)
	// externalStorage = nil
	// internalStorage = mapstorage.New()

	externalStorage = nil
	internalStorage = db.NewDBStorage(context.Background(), configuration.DatabaseURI)

	handlers := &handlers.Handlers{
		Storage: internalStorage,
		// Signer:  signchecker.NewSHA256(configuration.Key),
		Conf: conf.ServerConfiguration{DatabaseURI: configuration.DatabaseURI},
	}

	metricsServer := server.New(configuration, externalStorage, handlers)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := metricsServer.Run(ctx)
	if err != nil {
		log.Fatal(err)
	}

}
