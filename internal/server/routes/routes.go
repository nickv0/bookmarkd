package routes

import (
	"github.com/go-chi/chi/v5"

	"bookmarkd/internal/core"
	"bookmarkd/internal/server/middleware"
)

func AddRoutes(
	mux chi.Router,
	config core.Config,
	registrationStore core.RegistrationStore,
	bookmarkStore core.BookmarkStore,
	eventService core.EventService,
	sessionStore core.SessionStore,
	userStore core.UserStore,
) {

	// Protected Routes
	mux.Group(func(r chi.Router) {
		// Ensure all paths that follow have a valid user session
		r.Use(middleware.AuthMiddleware(config, sessionStore))

		// Delete the current user session
		r.Delete("/auth/logout", handleAuthLogoutDelete(sessionStore))

		// Get a single user
		r.Get("/users/{uid}", handleUsersIDGet(userStore))

		// List all sessions belonging to user
		r.Get("/users/{uid}/sessions", handleUsersIDSessionsGet(userStore))

		// Get a single user session
		r.Get("/users/{uid}/sessions/{sid}", handleUsersIDSessionsIDGet(sessionStore))

		// List all bookmarks.
		r.Get("/bookmarks", handleBookmarksGet(bookmarkStore))

		// Create a bookmark.
		r.Post("/bookmarks", handleBookmarksCreate(bookmarkStore))

		// View a single bookmark.
		r.Get("/bookmarks/{id}", handleBookmarksIDGet(bookmarkStore))

		// Update a bookmark.
		r.Patch("/bookmarks/{id}", handleBookmarksIDPatch(bookmarkStore))

		// Remove a bookmark.
		r.Delete("/bookmarks/{id}", handleBookmarksIDDelete(bookmarkStore))
	})

	// Public Routes
	mux.Group(func(r chi.Router) {

		// Initiate a websocket events subscription
		r.Get("/events", handleEventsGet(eventService))

		// Start a register flow
		r.Post("/auth/register", handleAuthRegisterPost(config, userStore, registrationStore))

		// Complete register
		r.Post("/auth/register/confirm", handleAuthRegisterConfirmPost(config, userStore, registrationStore, sessionStore))

		// Start a login flow
		r.Post("/auth/login", handleAuthLoginPost(config, userStore, sessionStore))

		// Exchange a refresh token for new refresh token and access token
		r.Post("/auth/refresh", handleAuthRefreshPost(config, sessionStore))
	})
}
