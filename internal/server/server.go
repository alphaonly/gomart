package server

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	conf "github.com/alphaonly/gomart/internal/configuration"
	"github.com/alphaonly/gomart/internal/server/handlers"
	stor "github.com/alphaonly/gomart/internal/server/storage/interfaces"
)

type Configuration struct {
	serverPort string
}

type Server struct {
	configuration   *conf.ServerConfiguration
	InternalStorage stor.Storage
	ExternalStorage stor.Storage
	handlers        *handlers.Handlers
	httpServer      *http.Server
}

func NewConfiguration(serverPort string) *Configuration {
	return &Configuration{serverPort: ":" + serverPort}
}

func New(configuration *conf.ServerConfiguration, ExStorage stor.Storage, handlers *handlers.Handlers) (server Server) {
	return Server{
		configuration:   configuration,
		InternalStorage: handlers.Storage,
		ExternalStorage: ExStorage,
		handlers:        handlers,
	}
}

func (s Server) ListenData(ctx context.Context) {
	// err := http.ListenAndServe(s.configuration.Port, s.handlers.NewRouter())
	err := s.httpServer.ListenAndServe()
	if err != nil {
		log.Println(err)
	}
}

func (s *Server) Run(ctx context.Context) error {

	// маршрутизация запросов обработчику
	s.httpServer = &http.Server{
		Addr: s.configuration.RunAddress,
		Handler:  s.handlers.NewRouter(),
	}

	// s.restoreData(ctx, s.ExternalStorage)

	go s.ListenData(ctx)
	// go s.ParkData(ctx, s.ExternalStorage)

	osSignal := make(chan os.Signal, 1)
	signal.Notify(osSignal, os.Interrupt)

	<-osSignal
	err := s.Shutdown(ctx)

	return err
}
func (s Server) Shutdown(ctx context.Context) error {
	time.Sleep(time.Second * 2)
	err := s.httpServer.Shutdown(ctx)
	log.Println("Server shutdown")
	return err
}

// func (s Server) restoreData(ctx context.Context, storageFrom stor.Storage) {
// 	if storageFrom == nil {
// 		log.Println("external storage  not initiated ")
// 		return
// 	}
// 	if s.configuration.Restore {
// 		mvList, err := storageFrom.GetAllMetrics(ctx)
// 		if err != nil {
// 			log.Println("cannot initially read metrics from file storage")
// 			return
// 		}
// 		if len(*mvList) == 0 {
// 			log.Println("file storage is empty, nothing to recover")
// 			return
// 		}

// 		err = s.InternalStorage.SaveAllMetrics(ctx, mvList)
// 		if err != nil {
// 			log.Fatal("cannot save metrics to internal storage")
// 		}

// 	}

// }

// func (s Server) ParkData(ctx context.Context, storageTo stor.Storage) {
// 	if storageTo == nil {
// 		return
// 	}
// 	if s.handlers.Storage == storageTo {
// 		log.Fatal("a try to save to it is own")
// 		return
// 	}

// 	ticker := time.NewTicker(time.Duration(s.configuration.StoreInterval))
// 	defer ticker.Stop()

// DoItAgain:
// 	select {

// 	case <-ticker.C:
// 		{

// 			mvList, err := s.InternalStorage.GetAllMetrics(ctx)
// 			if err != nil {
// 				log.Fatal("cannot read metrics from internal storage")
// 			}
// 			if mvList == nil {
// 				log.Println("read insufficient, internal storage empty")
// 			} else if len(*mvList) == 0 {
// 				log.Println("internal storage is empty, nothing to save to file")
// 			} else {
// 				err = storageTo.SaveAllMetrics(ctx, mvList)
// 				if err != nil {
// 					log.Fatal("cannot write metrics to file storage:" + err.Error())
// 				}
// 				log.Println("saved to file")
// 			}

// 		}
// 	case <-ctx.Done():
// 		return

// 	}
// 	goto DoItAgain
// }
