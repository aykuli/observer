// Agent is the application for getting OS metrics periodically
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

	"github.com/aykuli/observer/cmd/agent/client"
	"github.com/aykuli/observer/internal/agent/config"
	"github.com/aykuli/observer/internal/agent/storage"
	"github.com/aykuli/observer/internal/ldflags"
	"github.com/aykuli/observer/internal/logger"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

// @title           Observer agent API
// @version         1.0
// @description     Agent garbages information from OS and send it to Observer server.
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
	aLogger, err := logger.New()
	if err != nil {
		log.Fatal(err)
	}

	// ------- INITIATION CONFIG OPTIONS  -------
	var aOptions config.Config
	aOptions.Init(aLogger)

	// ------- CLIENT -------
	memStorage := storage.NewMemStorage()
	metricsClient := client.NewMetricsClient(aOptions, &memStorage, aLogger)

	// ------- SERVER -------
	pprofSrv := http.Server{Addr: "localhost:6061"}

	// ------- CONTEXT & WAIT GROUP FOR SYNC AGENT GOROUTINE GRACEFULLY SHUTDOWN WITH APPLICATION'S & DB -------
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	exit := make(chan os.Signal, 1)
	waiter := make(chan os.Signal, 1)

	// ------- PPROF PROFILER SERVER -------
	wg.Add(1)
	go func(w chan os.Signal, pprofwg *sync.WaitGroup) {
		if err := pprofSrv.ListenAndServe(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				fmt.Println(err)
				w <- syscall.SIGTERM
			}
			pprofwg.Done()
		}
	}(waiter, &wg)

	// ------- GRACEFULLY SHUTDOWN -------
	go func() {
		// Waiter wait until it gets one of:
		//   |-- app interrupting signals on err | syscall.SIGTERM -- terminated
		//   |-- keyboard Ctrl+C                 | syscall.SIGINT  -- interrupt
		//   |-- Ctrl + Backslash                | syscall.SIGQUIT -- quit
		signal.Notify(waiter, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT) // terminated
		sig := <-waiter
		aLogger.Info("1 Signal notify got signal " + sig.String())

		aLogger.Info("2 Started to shut down pprof server")
		if err := pprofSrv.Shutdown(ctx); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				log.Fatal("Pprof HTTP Server shutdown error: " + err.Error())
			}
		}
		aLogger.Info("3 Finished to shut down pprof server")

		// Cancel context
		cancel()
		aLogger.Info("4 Context cancelled")

		// Wait until all goroutines finishes their work
		wg.Wait()

		// Signal main goroutine that gracefully shutdown finished
		exit <- syscall.SIGSTOP
	}()
	wg.Add(1)

	// ------- START GRAB AND SEND METRICS TO SERVER -------
	metricsClient.Start(ctx, &wg)

	// Wait until gracefully shutdown finished
	<-exit
	aLogger.Info("7 Main goroutine exited.")
}
