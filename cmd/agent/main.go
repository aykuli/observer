// Agent is the application for getting OS metrics periodically
package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/aykuli/observer/cmd/agent/client"
	"github.com/aykuli/observer/internal/agent/config"
	"github.com/aykuli/observer/internal/agent/storage"
	"github.com/aykuli/observer/internal/ldflags"
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
	ldflags.Print(ldflags.BuildInfo{
		BuildVersion: buildVersion,
		BuildDate:    buildDate,
		BuildCommit:  buildCommit,
	})

	memStorage := storage.NewMemStorage()
	newClient := client.NewMetricsClient(config.Options, &memStorage)

	collectTicker := time.NewTicker(time.Duration(config.Options.PollInterval) * time.Second)
	sendTicker := time.NewTicker(time.Duration(config.Options.ReportInterval) * time.Second)
	defer collectTicker.Stop()
	defer sendTicker.Stop()

	go func() {
		if err := http.ListenAndServe("localhost:6061", nil); err != nil {
			log.Fatal(err)
		}
	}()

	for {
		select {
		case <-collectTicker.C:
			memStorage.GarbageStats()
			memStorage.GetSystemUtilInfo()
		case <-sendTicker.C:
			if config.Options.RateLimit > 0 {
				newClient.SendMetrics()
			} else {
				newClient.SendBatchMetrics()
			}
		}
	}
}
