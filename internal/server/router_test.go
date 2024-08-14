package server_test

import (
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"bookmarkd/internal/core"
	"bookmarkd/internal/mock"
	"bookmarkd/internal/server/routes"
)

func TestServerRuns(t *testing.T) {
	// var config, _ = core.NewConfig(os.Getenv)
	// r := server.NewRouter(nil, config, nil, nil, nil, nil)

	t.Run("server should start without error", func(t *testing.T) {
		// pass
	})

	t.Run("server should close without error", func(t *testing.T) {
		// pass
	})
}

func Test_Paths(t *testing.T) {
	config, _ := core.NewConfig(os.Getenv)

	mockRegistrationStore := mock.RegistrationStore{}
	mockEventService := mock.EventService{}
	mockBookmarkStore := mock.BookmarkStore{}
	mockSessionStore := mock.SessionStore{}
	mockUserStore := mock.UserStore{}

	r := chi.NewRouter()
	routes.AddRoutes(r, config, &mockRegistrationStore, &mockBookmarkStore, &mockEventService, &mockSessionStore, &mockUserStore)

	walkFunc := func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		route = strings.Replace(route, "/*/", "/", -1)
		t.Logf("%s %s\n", method, route)
		return nil
	}

	if err := chi.Walk(r, walkFunc); err != nil {
		t.Fatalf("Logging err: %s\n", err.Error())
	}

	//t.Fatal()
}
