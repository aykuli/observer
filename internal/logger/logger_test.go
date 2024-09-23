package logger

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func Example() {
	logger := zap.NewExample()
	defer logger.Sync()

	r := chi.NewRouter()
	r.Use(WithLogging(logger))
	r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if err := r.Body.Close(); err != nil {
			fmt.Println("request body close error")
			return
		}
	})

	testServer := httptest.NewServer(r)
	res, err := http.Get(testServer.URL + "/test")
	if err != nil {
		fmt.Println("req sending error")
		return
	}
	time.Sleep(time.Microsecond)
	res.Body.Close()
	// Output:
	// {"level":"info","msg":"server","Method":"GET","Status code":200,"URI":"/test","Size":0}
}
