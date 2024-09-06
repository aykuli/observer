// Server is the application for storing metrics sent by agent application.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"go.uber.org/zap"

	"github.com/aykuli/observer/cmd/server/routers"
	"github.com/aykuli/observer/internal/ldflags"
	"github.com/aykuli/observer/internal/logger"
	"github.com/aykuli/observer/internal/server/config"
	"github.com/aykuli/observer/internal/server/storage"
	"github.com/aykuli/observer/internal/server/storage/local"
	"github.com/aykuli/observer/internal/server/storage/postgres"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

// @title           Observer server API
// @version         1.0
// @description     Server provides functionality to handle metric values.
// @termsOfService  http://swagger.io/terms/

// @contact.name   Aynur Shauerman
// @contact.email  aykuli@ya.ru

// @license.name  MIT
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /
func main() {
	// ------- INFO IN CONSOLE ABOUT APP -------
	fmt.Println(ldflags.BuildInfo(ldflags.Info{
		BuildVersion: buildVersion,
		BuildDate:    buildDate,
		BuildCommit:  buildCommit,
	}))

	// ------- LOGGER -------
	sLogger, err := logger.New()
	if err != nil {
		log.Fatal(err)
	}

	// ------- INITIATION CONFIG OPTIONS  -------
	var sOptions config.Config
	sOptions.Init(sLogger)

	// ------- STORAGE -------
	memStorage, err := initStorage(sLogger, sOptions)
	if err != nil {
		log.Fatal(err)
	}

	// ------- SERVERS -------
	srv := http.Server{Addr: sOptions.Address, Handler: routers.MetricsRouter(memStorage, sLogger, sOptions)}
	pprofSrv := http.Server{Addr: "localhost:6060"}

	// ------- CONTEXT & WAIT GROUP FOR SYNC AGENT GOROUTINE GRACEFULLY SHUTDOWN WITH APPLICATION'S & DB -------
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	exit := make(chan os.Signal, 1)
	waiter := make(chan os.Signal, 1)

	// ------- PPROF PROFILER SERVER -------
	wg.Add(1)
	go func(w chan os.Signal, pprofwg *sync.WaitGroup) {
		if er := pprofSrv.ListenAndServe(); er != nil {
			if !errors.Is(er, http.ErrServerClosed) {
				sLogger.Error(er.Error())
				w <- syscall.SIGTERM
			}
			pprofwg.Done()
		}
	}(waiter, &wg)

	// ------- GRACEFULLY SHUTDOWN -------
	go func() {
		// Waiter wait until it gets one of:
		//   |-- application interrupting signals on err | syscall.SIGTERM -- terminated
		//   |-- keyboard Ctrl+C                         | syscall.SIGINT  -- interrupt
		//   |-- Ctrl + Backslash                        | syscall.SIGQUIT -- quit
		signal.Notify(waiter, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT) // terminated
		sig := <-waiter
		fmt.Println()
		sLogger.Info("1 Signal notify got signal " + sig.String())

		sLogger.Info("2 Started to shut down pprof server")
		if err := pprofSrv.Shutdown(ctx); err != nil {
			sLogger.Error("Pprof HTTP Server shutdown error: " + err.Error())
		}
		sLogger.Info("2 Finished to shut down pprof server")

		sLogger.Info("3 Started to shut down server")
		if err := srv.Shutdown(ctx); err != nil {
			sLogger.Fatal("HTTP Server shutdown error: " + err.Error())
		}
		sLogger.Info("4 Finished to shut down server")

		// Cancel context
		cancel()
		sLogger.Info("5 Context cancelled")

		if err = memStorage.Close(); err != nil {
			sLogger.Fatal("Storage closing error " + err.Error())
		}

		// Wait until all servers shuts down
		wg.Wait()

		sLogger.Sync()

		// Signal main goroutine that gracefully shutdown finished
		exit <- syscall.SIGSTOP
	}()

	// ------- START SERVER -------
	wg.Add(1)
	if err := srv.ListenAndServe(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			sLogger.Error(err.Error())
			waiter <- syscall.SIGTERM
		}
		wg.Done()
	}
	<-exit
	sLogger.Info("6 Main goroutine exited.")
}

// initStorage configures storage type by parameters provided when app was started.
func initStorage(logger *zap.Logger, options config.Config) (storage.Storage, error) {
	if options.DatabaseDsn != "" {
		return postgres.NewStorage(options.DatabaseDsn)
	}

	return local.NewStorage(options, logger)
}
