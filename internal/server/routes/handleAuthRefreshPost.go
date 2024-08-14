package routes

import (
	"net/http"

	"bookmarkd/internal/core"
	"bookmarkd/internal/server/encoder"
	"bookmarkd/internal/server/jwt"
)

func handleAuthRefreshPost(
	config core.Config,
	sessionStore core.SessionStore,
) http.HandlerFunc {

	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			token := jwt.GetJwtTokenFromRequest(r)

			// create a session object for the auth
			s, err := sessionStore.RefreshSession(r.Context(), token)
			if err != nil {
				encoder.EncodeError(w, r, err)
				return
			}

			t := jwt.CreateJWT(config, s.ID, s.RefreshToken)

			if err := encoder.EncodeJson(w, http.StatusOK, t); err != nil {
				encoder.EncodeError(w, r, err)
				return
			}
		})
}
