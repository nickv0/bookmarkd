package middleware

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/httplog/v2"

	"bookmarkd"
	"bookmarkd/internal/core"
	"bookmarkd/internal/server/encoder"
	"bookmarkd/internal/server/jwt"
)

func AuthMiddleware(config core.Config, sessionStore core.SessionStore) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			if core.ValidSessionFromContext(r.Context()) {
				next.ServeHTTP(w, r)
				return
			}

			tokenString := jwt.GetJwtTokenFromRequest(r)

			if tokenString == "" {
				encoder.EncodeError(w, r, bookmarkd.ErrUnauthorized)
				return
			}

			token, err := jwt.ValidateJWT(config, tokenString)
			if err != nil {
				encoder.EncodeError(w, r, bookmarkd.ErrUnauthorized)
				return
			}

			// decode the bearerToken and take sessionID from payuload
			idString, err := token.GetString("sessionID")
			if err != nil {
				encoder.EncodeError(w, r, bookmarkd.ErrUnauthorized)
				return
			}

			id, err := strconv.Atoi(idString)
			if err != nil {
				encoder.EncodeError(w, r, bookmarkd.ErrUnauthorized)
				return
			}

			s, err := sessionStore.FindSessionByID(r.Context(), id)
			if err != nil {
				encoder.EncodeError(w, r, bookmarkd.ErrUnauthorized)
				return
			}

			// Add the session to the context and update the request with context
			ctx := r.Context()
			httplog.LogEntrySetField(ctx, "user", slog.StringValue(s.UserID))
			ctx = core.NewContextWithSession(ctx, core.SessionContext{SessionID: s.ID, UserID: s.UserID})
			r = r.WithContext(ctx)

			// Call the function if the token is valid
			next.ServeHTTP(w, r)
		})
	}
}
