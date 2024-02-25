package main

import (
	"html/template"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog/v2"
	"github.com/go-chi/render"
)

type RestSever struct {
	mux      *chi.Mux
	entities []Entity
}

func NewRestServer(entities []Entity, options ...func(*RestSever)) *RestSever {
	mux := chi.NewRouter()
	mux.Use(middleware.RedirectSlashes)
	mux.Use(middleware.Heartbeat("/ping"))
	mux.Use(middleware.Recoverer)

	server := &RestSever{
		mux,
		entities,
	}

	for _, opt := range options {
		opt(server)
	}

	return server
}

// Middleware: Adds a logger to the server
// TODO: This is a placeholder, it should be replaced with a real logger
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

// Generates CRUD routes for each entity
func (s *RestSever) InitRouter() {
	for _, entity := range s.entities {
		basePath := "/" + entity.Name
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

// Middleware: Adds pagination to the server
// TODO: implement pagination
func Paginate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

// Helper function to return a json response
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

// Middleware: Adds a static file server
func AddStaticFiles(path string) func(*RestSever) {
	return func(s *RestSever) {
		s.mux.Handle("/*", http.FileServer(http.Dir(path)))
	}
}

// Middleware: Adds a home page to the server similar to Swagger
func AddHomePage(schemaPath string) func(*RestSever) {
	return func(s *RestSever) {
		s.mux.Get("/", func(w http.ResponseWriter, r *http.Request) {
			// Define the data to be passed to the template
			page := struct {
				Schema   string
				Entities []Entity
				URL      string
			}{
				Schema:   schemaPath,
				Entities: s.entities,
				URL:      r.URL.String(),
			}

			tmpl := template.Must(template.ParseFiles("./assets/home-template.gohtmltmpl"))
			err := tmpl.Execute(w, page)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		})
	}
}
