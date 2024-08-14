package routes

import (
	"fmt"
	"net/http"
	"time"

	"bookmarkd"
	"bookmarkd/internal/core"
	"bookmarkd/internal/server/encoder"
	"bookmarkd/internal/server/jwt"

	"github.com/cristalhq/otp"
)

type AuthRegisterConfirmPostPayload struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	Expires      int    `json:"expires"`
}

type AuthRegisterConfirmPostInput struct {
	RegistrationID string `json:"registrationId"`
	Totp           string `json:"totp"`
}

func (o *AuthRegisterConfirmPostInput) validate() error {
	if o.RegistrationID == "" {
		return fmt.Errorf("missing username")
	} else if o.Totp == "" {
		return fmt.Errorf("missing totp")
	}
	return nil
}

func handleAuthRegisterConfirmPost(
	config core.Config,
	userStore core.UserStore,
	registrationStore core.RegistrationStore,
	sessionStore core.SessionStore,
) http.HandlerFunc {

	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			input, err := encoder.DecodeJson[AuthRegisterConfirmPostInput](r)
			if err != nil {
				encoder.EncodeError(w, r, bookmarkd.ErrInvalidInput)
				return
			}

			// validate input
			if err := input.validate(); err != nil {
				encoder.EncodeError(w, r, bookmarkd.ErrInvalidInput)
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

			reg, err := registrationStore.FindRegistrationByID(input.RegistrationID)
			if err != nil {
				encoder.EncodeError(w, r, err)
				return
			}

			if err := totp.Validate(input.Totp, time.Now(), reg.Seed); err != nil {
				encoder.EncodeError(w, r, fmt.Errorf("%w: invalid passcode", bookmarkd.ErrBadRequest))
				return
			}

			registrationStore.DeleteRegistration(reg.ID.String())

			user := core.User{
				Username: reg.Username,
				Seed:     reg.Seed,
			}

			if err := userStore.CreateUser(r.Context(), &user); err != nil {
				encoder.EncodeError(w, r, err)
				return
			}

			session := core.Session{
				UserID: user.ID,
			}
			if err := sessionStore.CreateSession(r.Context(), &session); err != nil {
				encoder.EncodeError(w, r, err)
				return
			}

			encoder.EncodeJson(w, http.StatusOK, jwt.CreateJWT(config, session.ID, session.RefreshToken))
		})
}
