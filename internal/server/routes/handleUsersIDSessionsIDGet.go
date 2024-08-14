package routes

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"bookmarkd"
	"bookmarkd/internal/core"
	"bookmarkd/internal/server/encoder"
)

func handleUsersIDSessionsIDGet(
	sessionStore core.SessionStore,
) http.HandlerFunc {

	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			uid := chi.URLParam(r, "uid")
			sid := chi.URLParam(r, "sid")
			id, err := strconv.Atoi(sid)
			if err != nil {
				encoder.EncodeError(w, r, bookmarkd.ErrNotFound)
				return
			}

			s, err := sessionStore.FindSessionByID(r.Context(), id)
			if err != nil {
				encoder.EncodeError(w, r, err)
				return
			}

			if s.UserID != uid {
				encoder.EncodeError(w, r, bookmarkd.ErrNotFound)
				return
			}

			encoder.EncodeJson(w, http.StatusOK, s)
		})
}
