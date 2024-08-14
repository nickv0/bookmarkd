package routes

import (
	"bookmarkd"
	"fmt"
	"net/http"
	"time"

	"bookmarkd/internal/core"
	"bookmarkd/internal/server/encoder"
	"bookmarkd/internal/server/jwt"

	"github.com/cristalhq/otp"
)

type AuthLoginPostPayload struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	Expires      int    `json:"expires"`
}

type AuthLoginPostInput struct {
	Username string `json:"username"`
	Totp     string `json:"totp"`
}

func (o *AuthLoginPostInput) validate() error {
	if o.Username == "" {
		return fmt.Errorf("missing username")
	} else if o.Totp == "" {
		return fmt.Errorf("missing totp")
	}
	return nil
}

func handleAuthLoginPost(
	config core.Config,
	userStore core.UserStore,
	sessionStore core.SessionStore,
) http.HandlerFunc {

	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			// slow all login attempts
			time.Sleep(1 * time.Second)

			input, err := encoder.DecodeJson[AuthLoginPostInput](r)
			if err != nil {
				encoder.EncodeError(w, r, err)
				return
			}

			// validate input
			if err := input.validate(); err != nil {
				encoder.EncodeError(w, r, err)
				return
			}

			user, err := userStore.FindUserByUsername(r.Context(), input.Username)
			if err != nil {
				encoder.EncodeError(w, r, bookmarkd.ErrUnauthorized)
				return
			}

			totp, err := otp.NewTOTP(otp.TOTPConfig{
				Algo:   config.TotpAlgo,
				Digits: config.TotpDigits,
				Issuer: config.TotpIssuer,
				Period: config.TotpPeriod,
				Skew:   config.TotpSkew,
			})
			if err != nil {
				encoder.EncodeError(w, r, err)
				return
			}

			if err := totp.Validate(input.Totp, time.Now(), user.Seed); err != nil {
				encoder.EncodeError(w, r, bookmarkd.ErrUnauthorized)
				return
			}

			// login validated, create auth session and return tokens
			s := core.Session{UserID: user.ID}
			if err := sessionStore.CreateSession(r.Context(), &s); err != nil {
				encoder.EncodeError(w, r, fmt.Errorf("create session: %w", err))
				return
			}

			encoder.EncodeJson(w, http.StatusOK, jwt.CreateJWT(config, s.ID, s.RefreshToken))
		})
}
