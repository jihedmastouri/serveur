package main

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog/v2"
	"github.com/go-chi/render"
)

type RestSever struct {
	mux *chi.Mux
}

func NewRestServer(options ...func(*RestSever)) *RestSever {
	r := chi.NewRouter()
	r.Use(middleware.Heartbeat("/ping"))
	r.Use(middleware.Recoverer)

	server := &RestSever{
		mux: r,
	}

	for _, opt := range options {
		opt(server)
	}

	return server
}

func AddLogger() func(*RestSever) {
	return func(s *RestSever) {
		logger := httplog.NewLogger("serveur", httplog.Options{
			// JSON:             true,
			LogLevel:         slog.LevelDebug,
			Concise:          true,
			RequestHeaders:   true,
			MessageFieldName: "message",
			// TimeFieldFormat: time.RFC850,
			Tags: map[string]string{
				"version": "v1.0-81aa4244d9fc8076a",
				"env":     "dev",
			},
			QuietDownRoutes: []string{
				"/",
				"/ping",
			},
			QuietDownPeriod: 10 * time.Second,
			// SourceFieldName: "source",
		})

		s.mux.Use(httplog.RequestLogger(logger))
	}
}

func AddHomePage(s *RestSever) {}

func (s *RestSever) InitRouter(entities []Entity) {
	for _, entity := range entities {
		basePath := "/" + entity.name + "/"
		s.mux.Route(basePath, func(r chi.Router) {
			r.Post("/", func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("Hello, World!"))
			})
			r.Get("/", func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("Hello, World!"))
			})
			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", func(w http.ResponseWriter, r *http.Request) {
					w.Write([]byte("Hello, World!"))
				})
				r.Delete("/", func(w http.ResponseWriter, r *http.Request) {
					w.Write([]byte("Hello, World!"))
				})
				r.Put("/", func(w http.ResponseWriter, r *http.Request) {
					w.Write([]byte("Hello, World!"))
				})
				r.Patch("/", func(w http.ResponseWriter, r *http.Request) {
					w.Write([]byte("Hello, World!"))
				})
			})
		})
	}
}

func paginate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

func Response(fn func() ([]byte, error)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := fn()
		if err != nil {
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, map[string]string{"error": err.Error()})
		}
		render.JSON(w, r, data)
	}
}
