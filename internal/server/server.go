package server

import (
	"net"
	"net/http"

	"github.com/go-chi/httplog/v2"

	"bookmarkd/internal/core"
)

func NewServer(logger *httplog.Logger,
	config core.Config,

	// stores and services
	registrationStore core.RegistrationStore,
	bookmarkStore core.BookmarkStore,
	eventService core.EventService,
	sessionStore core.SessionStore,
	userStore core.UserStore,
) *http.Server {

	r := NewRouter(
		logger,
		config,
		registrationStore,
		bookmarkStore,
		eventService,
		sessionStore,
		userStore,
	)

	return &http.Server{
		Addr:    net.JoinHostPort(config.HttpAddress, config.HttpPort),
		Handler: r,
	}
}
