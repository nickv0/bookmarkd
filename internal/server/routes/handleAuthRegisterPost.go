package routes

import (
	"fmt"
	"net/http"

	"bookmarkd"
	"bookmarkd/internal/core"
	"bookmarkd/internal/server/encoder"

	"github.com/cristalhq/otp"
)

type AuthRegisterPostPayload struct {
	RegistrationID string `json:"registrationId"`
	TotpUrl        string `json:"totpUrl"`
}

type AuthRegisterPostInput struct {
	Username string `json:"username"`
}

func (o *AuthRegisterPostInput) validate() error {
	if o.Username == "" {
		return fmt.Errorf("missing username")
	}
	return nil
}

func handleAuthRegisterPost(
	config core.Config,
	userStore core.UserStore,
	registrationStore core.RegistrationStore,
) http.HandlerFunc {

	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			input, err := encoder.DecodeJson[AuthRegisterPostInput](r)
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

			// ensure username is not already in use
			if _, err := userStore.FindUserByUsername(r.Context(), input.Username); err != nil {
				encoder.EncodeError(w, r, fmt.Errorf("%w: username in use", bookmarkd.ErrBadRequest))
				return
			}

			// TODO: filter bad usernames

			// start registration session
			reg, err := registrationStore.StartRegistration(input.Username)
			if err != nil {
				encoder.EncodeError(w, r, err)
			}

			encoder.EncodeJson(w, http.StatusOK, &AuthRegisterPostPayload{
				RegistrationID: reg.ID.String(),
				TotpUrl:        totp.GenerateURL(input.Username, []byte(reg.Seed)),
			})
		})
}
