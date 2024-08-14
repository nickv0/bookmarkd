package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"bookmarkd/internal/core"
	"bookmarkd/internal/server/encoder"
)

func handleUsersIDSessionsGet(
	userStore core.UserStore,
) http.HandlerFunc {

	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			uid := chi.URLParam(r, "uid")

			u, err := userStore.FindUserByID(r.Context(), uid)
			if err != nil {
				encoder.EncodeError(w, r, err)
				return
			}

			encoder.EncodeJson(w, http.StatusOK, u.Sessions)

		})
}
