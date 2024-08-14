package server

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/go-chi/httplog/v2"

	"bookmarkd"
	"bookmarkd/internal/core"
	"bookmarkd/internal/server/encoder"
	"bookmarkd/internal/server/routes"
)

// NewServer returns a new instance of Server.
func NewRouter(
	logger *httplog.Logger,
	config core.Config,

	// stores and services
	registrationStore core.RegistrationStore,
	bookmarkStore core.BookmarkStore,
	eventService core.EventService,
	sessionStore core.SessionStore,
	userStore core.UserStore,
) *chi.Mux {

	r := chi.NewRouter()

	// enable middleware as long as we are not running a test
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(httplog.RequestLogger(logger))

	// Dont report errors or attempt to recover in test
	if !config.Test {
		r.Use(handleReportPanic)
		r.Use(middleware.Recoverer)
	}

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	t := time.Second * time.Duration(config.HttpTimeoutInSeconds)
	r.Use(middleware.Timeout(t))

	r.Use(middleware.Heartbeat(config.HttpBasePath + "/ping"))

	// Setup error handling routes.
	r.NotFound(handleNotFound)

	// handle cors
	// Basic CORS
	// for more ideas, see: https://developer.github.com/v3/#cross-origin-resource-sharing
	r.Use(cors.Handler(cors.Options{
		// AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins: []string{"https://*", "http://*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	r.Route(config.HttpBasePath, func(r chi.Router) {
		routes.AddRoutes(
			r,
			config,
			registrationStore,
			bookmarkStore,
			eventService,
			sessionStore,
			userStore,
		)
	})

	return r
}

func handleReportPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				bookmarkd.ReportPanic(err)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// handleNotFound handles requests to routes that don't exist.
func handleNotFound(w http.ResponseWriter, r *http.Request) {

	type response struct {
		StatusCode int    `json:"statusCode"`
		Message    string `json:"message"`
	}

	if err := encoder.EncodeJson[response](w, http.StatusNotFound, response{
		StatusCode: http.StatusNotFound,
		Message:    "Sorry, it looks like we can't find what you're looking for.",
	}); err != nil {
		http.Error(w, "Error creating JSON response", http.StatusInternalServerError)
	}
}
