package main

import (
	"expvar"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

type application struct {
	errorLog *log.Logger
	infoLog  *log.Logger
}

func (app *application) routes() http.Handler {
	router := http.NewServeMux()
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "hello world!")
	})
	router.Handle("/metrics", expvar.Handler())
	return router
}

func main() {
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	app := application{
		errorLog: errorLog,
		infoLog:  infoLog,
	}
	addr := "localhost:17171"
	svr := http.Server{
		Addr:         addr,
		ErrorLog:     errorLog,
		Handler:      app.routes(),
		ReadTimeout:  time.Second * 5,
		WriteTimeout: time.Second * 5,
		IdleTimeout:  time.Minute,
	}
	infoLog.Printf("Starting server on %s", addr)
	err := svr.ListenAndServe()
	errorLog.Fatal(err)
}
