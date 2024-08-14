package routes

import (
	"net/http"

	"github.com/go-chi/httplog/v2"

	"bookmarkd/internal/core"
	"bookmarkd/internal/server/encoder"
)

type AuthLogoutDeleteResponse struct {
	Logout string `json:"logout"`
}

func handleAuthLogoutDelete(
	sessionStore core.SessionStore,
) http.HandlerFunc {

	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {

			sessionID := core.GetSessionIDFromContext(r.Context())

			if err := sessionStore.DeleteSession(r.Context(), sessionID); err != nil {
				oplog := httplog.LogEntry(r.Context())
				oplog.Warn("delete user session")
				// we only log that there was a problem, session could already be deleted
				// for a number of reasons, import part is that client side
				// continues the logout process.
			}

			encoder.EncodeJson(w, http.StatusOK, &AuthLogoutDeleteResponse{Logout: "OK"})
		})
}
