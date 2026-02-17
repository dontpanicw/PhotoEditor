package http

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/dontpanicw/ImageProcessor/internal/port"
	"github.com/gorilla/mux"
)

type Server struct {
	handler *Handler
	server  *http.Server
}

func NewServer(port string, usecases port.ImageUsecases) *Server {
	handler := NewHandler(usecases)

	router := mux.NewRouter()

	// CORS middleware
	router.Use(corsMiddleware)

	// Статические файлы
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./web/static"))))
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./web/index.html")
	})

	// API маршруты
	router.HandleFunc("/upload", handler.UploadImage).Methods("POST", "OPTIONS")
	router.HandleFunc("/image/{id}", handler.GetImage).Methods("GET", "OPTIONS")
	router.HandleFunc("/image/{id}/status", handler.GetImageStatus).Methods("GET", "OPTIONS")
	router.HandleFunc("/image/{id}", handler.DeleteImage).Methods("DELETE", "OPTIONS")

	server := &http.Server{
		Addr:         port,
		Handler:      router,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  300 * time.Second,
	}

	return &Server{
		handler: handler,
		server:  server,
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

func (s *Server) Start() error {
	log.Printf("Starting HTTP server on %s\n", s.server.Addr)
	return s.server.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	fmt.Println("Shutting down HTTP server...")
	return s.server.Shutdown(ctx)
}
