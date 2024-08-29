// Server is the application for storing metrics sent by agent application.
package main

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"

	"go.uber.org/zap"

	"github.com/aykuli/observer/cmd/server/routers"
	"github.com/aykuli/observer/internal/ldflags"
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
	fmt.Println(ldflags.BuildInfo(ldflags.Info{
		BuildVersion: buildVersion,
		BuildDate:    buildDate,
		BuildCommit:  buildCommit,
	}))

	serverLogger, err := zap.NewProduction()
	defer func() {
		if err = serverLogger.Sync(); err != nil {
			panic(err)
		}
	}()
	sugar := *serverLogger.Sugar()

	memStorage, err := initStorage(sugar)
	if err != nil {
		log.Print(err)
	}

	go func() {
		if er := http.ListenAndServe("localhost:6060", nil); err != nil {
			log.Fatal(er)
		}
	}()

	if err = http.ListenAndServe(config.Options.Address, routers.MetricsRouter(memStorage, sugar)); err != nil {
		sugar.Fatalw(err.Error(), "event", "start server")
	}
}

// initStorage configures storage type by parameters provided when app was started.
func initStorage(logger zap.SugaredLogger) (storage.Storage, error) {
	if config.Options.DatabaseDsn != "" {
		return postgres.NewStorage(config.Options.DatabaseDsn)
	}

	return local.NewStorage(config.Options, logger)
}
