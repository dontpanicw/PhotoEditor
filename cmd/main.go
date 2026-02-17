package main

import (
	"github.com/dontpanicw/ImageProcessor/config"
	"github.com/dontpanicw/ImageProcessor/internal/app"
	"log"
	"net/http"
	"net/http/pprof"
)

func main() {

	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatal("error creating config")
	}

	// Создаем отдельный маршрутизатор для pprof
	go func() {
		mux := http.NewServeMux()

		// Явно регистрируем все обработчики pprof
		mux.HandleFunc("/debug/pprof/", pprof.Index)
		mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

		// Регистрируем профили
		mux.Handle("/debug/pprof/goroutine", pprof.Handler("goroutine"))
		mux.Handle("/debug/pprof/heap", pprof.Handler("heap"))
		mux.Handle("/debug/pprof/threadcreate", pprof.Handler("threadcreate"))
		mux.Handle("/debug/pprof/block", pprof.Handler("block"))
		mux.Handle("/debug/pprof/mutex", pprof.Handler("mutex"))

		log.Println("Starting pprof server on :6060")
		if err := http.ListenAndServe("localhost:6060", mux); err != nil {
			log.Printf("pprof server error: %v", err)
		}
	}()

	if err := app.Start(cfg); err != nil {
		log.Fatal("failed to start application")
	}

}
