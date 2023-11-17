package main

import (
	"expvar"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/rs/zerolog"
)

const version = "1.0.0"

type application struct {
	errorLog *zerolog.Logger
	infoLog  *zerolog.Logger
}

func (app *application) logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.infoLog.Printf("%s - %s %s %s", r.RemoteAddr, r.Proto, r.Method, r.URL.RequestURI())
		next.ServeHTTP(w, r)
	})
}

func (app *application) metricsMiddleware(next http.Handler) http.Handler {
	totalRequestsReceived := expvar.NewInt("total_requests_received")
	totalResponsesSent := expvar.NewInt("total_responses_sent")
	totalProcessingTimeMicroseconds := expvar.NewInt("total_processing_time_microsecs")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		totalRequestsReceived.Add(1)

		next.ServeHTTP(w, r)

		totalResponsesSent.Add(1)

		duration := time.Since(start).Microseconds()
		totalProcessingTimeMicroseconds.Add(duration)
	})
}

func (app *application) routes() http.Handler {
	router := http.NewServeMux()
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "hello world!")
	})
	router.Handle("/metrics", expvar.Handler())
	return app.metricsMiddleware(app.logMiddleware(router))
}

func main() {
	expvar.NewString("version").Set(version)
	expvar.Publish("goroutines", expvar.Func(func() interface{} {
		return runtime.NumGoroutine()
	}))
	expvar.Publish("timestamp", expvar.Func(func() interface{} {
		return time.Now().Unix()
	}))

	errorLog := zerolog.New(os.Stderr).With().Timestamp().Caller().Logger()
	infoLog := zerolog.New(os.Stdout).With().Timestamp().Logger()
	app := application{
		errorLog: &errorLog,
		infoLog:  &infoLog,
	}
	addr := "localhost:17171"
	svr := http.Server{
		Addr: addr,
		// ErrorLog:     errorLog,
		Handler:      app.routes(),
		ReadTimeout:  time.Second * 5,
		WriteTimeout: time.Second * 5,
		IdleTimeout:  time.Minute,
	}
	infoLog.Printf("Starting server on %s", addr)
	err := svr.ListenAndServe()
	errorLog.Fatal().Err(err).Send()
}
