package main

import (
	"embed"
	"encoding/json"
	"errors"
	"html/template"
	"io"
	"io/fs"
	"log"
	"log/slog"
	"net/http"
	"reflect"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog/v2"
	"github.com/go-chi/render"
	"github.com/go-faker/faker/v4"
)

//go:embed templates/*
var tmpls embed.FS

//go:embed assets/*
var assets embed.FS

type RestSever struct {
	db       Store
	mux      *chi.Mux
	entities []Entity
}

func NewRestServer(db Store, entities []Entity, options ...func(*RestSever)) *RestSever {
	mux := chi.NewRouter()
	mux.Use(middleware.RedirectSlashes)
	mux.Use(middleware.Heartbeat("/ping"))
	mux.Use(middleware.Recoverer)

	server := &RestSever{
		db,
		mux,
		entities,
	}

	for _, opt := range options {
		opt(server)
	}

	return server
}

// Generates CRUD routes for each entity
func (s *RestSever) InitRouter() {
	for _, entity := range s.entities {
		s.mux.Route("/"+entity.Name, func(r chi.Router) {
			r.Post("/", Response(s.PostHandler(entity.Name)))
			r.Get("/", Response(s.GetAllHandler(entity.Name)))
			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", Response(s.GetHandler(entity.Name)))
				r.Delete("/", Response(s.DeleteHandler(entity.Name)))
				r.Put("/", Response(s.PutHandler(entity.Name)))
				r.Patch("/", Response(s.PatchHandler(entity.Name)))
			})
		})
	}
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

/*************
* Middleware
*************/

// Middleware: Adds pagination to the server
// TODO: implement pagination
func Paginate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

// Middleware: Adds a static file server
func AddStaticFiles(path string) func(*RestSever) {
	if path == "" {
		return func(s *RestSever) {}
	}
	return func(s *RestSever) {
		s.mux.Handle("/static/", http.FileServer(http.Dir(path)))
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

			tmpl := template.Must(template.ParseFS(tmpls, "templates/home/*.html"))

			err := tmpl.ExecuteTemplate(w, "index", page)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		})

		subFS, err := fs.Sub(assets, "assets")
		if err != nil {
			ErrExit("Home Page: Couldn't get the assets:", err)
		}

		fileServer := http.FileServer(http.FS(subFS))
		s.mux.Handle("/assets/*", http.StripPrefix("/assets/", fileServer))
	}
}

/*************
* Handlers
*************/

func (s *RestSever) PostHandler(entityName string) handlerResponse {
	return func(r *http.Request) (any, *ResError) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			return nil, &ResError{
				Error:  err.Error(),
				Status: http.StatusBadRequest,
			}
		}
		defer r.Body.Close()

		var params map[string]interface{}
		err = json.Unmarshal(body, &params)
		if err != nil {
			return nil, &ResError{
				Error:  errors.New("invalid json").Error(),
				Status: http.StatusBadRequest,
			}
		}

		id := faker.UUIDDigit()
		if params["id"] != nil {
			id = params["id"].(string)
		}

		err = s.db.Set(entityName, []byte(id), []byte(body))
		if err != nil {
			return nil, &ResError{
				Error:  err.Error(),
				Status: http.StatusInternalServerError,
			}
		}
		return []byte(SuccessMessage), nil
	}
}

func (s *RestSever) GetAllHandler(entityName string) handlerResponse {
	return func(r *http.Request) (any, *ResError) {
		res, err := s.db.GetAll(entityName, nil)
		if err != nil {
			return nil, &ResError{
				Error:  err.Error(),
				Status: http.StatusInternalServerError,
			}
		}
		return res, nil
	}
}

func (s *RestSever) GetHandler(entityName string) handlerResponse {
	return func(r *http.Request) (any, *ResError) {
		id := r.URL.Query().Get("id")
		if id == "" {
			return nil, &ResError{
				Error:  errors.New("id is required").Error(),
				Status: http.StatusBadRequest,
			}
		}
		res, err := s.db.Get(entityName, []byte(id))
		if err != nil {
			return nil, &ResError{
				Error:  err.Error(),
				Status: http.StatusInternalServerError,
			}
		}
		return res, nil
	}
}

func (s *RestSever) DeleteHandler(entityName string) handlerResponse {
	return func(r *http.Request) (any, *ResError) {
		id := r.URL.Query().Get("id")
		if id == "" {
			return nil, &ResError{
				Error:  errors.New("id is required").Error(),
				Status: http.StatusBadRequest,
			}
		}
		err := s.db.Delete(entityName, []byte(id))
		if err != nil {
			return nil, &ResError{
				Error:  err.Error(),
				Status: http.StatusInternalServerError,
			}
		}

		return []byte(SuccessMessage), nil
	}
}

func (s *RestSever) PutHandler(entityName string) handlerResponse {
	return func(r *http.Request) (any, *ResError) {
		id := r.URL.Query().Get("id")
		if id == "" {
			return nil, &ResError{
				Error:  errors.New("id is required").Error(),
				Status: http.StatusBadRequest,
			}
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			return nil, &ResError{
				Error:  err.Error(),
				Status: http.StatusBadRequest,
			}
		}
		defer r.Body.Close()

		err = s.db.Set(entityName, []byte(id), body)
		if err != nil {
			return nil, &ResError{
				Error:  err.Error(),
				Status: http.StatusInternalServerError,
			}
		}

		return []byte(SuccessMessage), nil
	}
}

func (s *RestSever) PatchHandler(entityName string) handlerResponse {
	return func(r *http.Request) (any, *ResError) {
		id := r.URL.Query().Get("id")
		if id == "" {
			return nil, &ResError{
				Error:  errors.New("id is required").Error(),
				Status: http.StatusBadRequest,
			}
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			return nil, &ResError{
				Error:  err.Error(),
				Status: http.StatusBadRequest,
			}
		}
		defer r.Body.Close()

		err = s.db.Patch(entityName, []byte(id), body)
		if err != nil {
			return nil, &ResError{
				Error:  err.Error(),
				Status: http.StatusInternalServerError,
			}
		}
		return []byte(SuccessMessage), nil
	}
}

/*************
* Utils
*************/

type ResError struct {
	Error  string `json:"error"`
	Status int    `json:"status"`
}

const SuccessMessage = "{\"message\": \"ok\"}"

type handlerResponse func(*http.Request) (any, *ResError)

// Helper function to return a json response
func Response(fn func(*http.Request) (any, *ResError)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := fn(r)
		if err != nil {
			render.Status(r, err.Status)
			render.JSON(w, r, map[string]string{"error": err.Error})
		}

		if reflect.TypeOf(data).Kind() == reflect.Slice {
			var res []map[string]any
			for _, v := range data.([][]byte) {
				var params map[string]any
				err := json.Unmarshal(v, &params)
				if err != nil {
					render.Status(r, http.StatusInternalServerError)
					render.JSON(w, r, map[string]string{"error": "Failed to Parse Data"})
					return
				}
				res = append(res, params)
			}
			render.JSON(w, r, res)
			return
		}

		render.JSON(w, r, data)
	}
}
