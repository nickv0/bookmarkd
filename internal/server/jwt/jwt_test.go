package jwt_test

import (
	"strings"
	"testing"

	"aidanwoods.dev/go-paseto"

	"bookmarkd/internal/core"
	"bookmarkd/internal/server/jwt"
	"bookmarkd/utils/require"
)

func getenv(s string) string { return "" }

var c, _ = core.NewConfig(getenv)

func TestCreateJWT(t *testing.T) {
	parser := paseto.NewParser()
	r := jwt.CreateJWT(c, 1, "opaque")

	require.Equal(t, strings.HasPrefix(r.AccessToken, "v4.public."), true)
	require.Equal(t, strings.HasPrefix(r.RefreshToken, "v4.public."), true)
	require.Equal(t, r.Expires, 300)
	require.Equal(t, r.TokenType, "bearer")

	_, err := parser.ParseV4Public(c.PasetoPublicKey, r.AccessToken, nil)
	require.Equal(t, err, nil)

	_, err = parser.ParseV4Public(c.PasetoPublicKey, r.RefreshToken, nil)
	require.Equal(t, err, nil)
}

func TestValidateJWT(t *testing.T) {
	response := jwt.CreateJWT(c, 1, "opaque")

	t.Run("AccessToken", func(t *testing.T) {
		token, err := jwt.ValidateJWT(c, response.AccessToken)

		require.Equal(t, err, nil)

		iss, err := token.GetIssuer()

		require.Equal(t, err, nil)
		require.Equal(t, iss, c.HttpDomain)

		aud, err := token.GetAudience()

		require.Equal(t, err, nil)
		require.Equal(t, aud, c.HttpDomain+c.HttpBasePath)

		sub, err := token.GetSubject()

		require.Equal(t, err, nil)
		require.Equal(t, sub, "1")
	})

	t.Run("RefreshToken", func(t *testing.T) {
		token, err := jwt.ValidateJWT(c, response.RefreshToken)

		require.Equal(t, err, nil)

		iss, err := token.GetIssuer()

		require.Equal(t, err, nil)
		require.Equal(t, iss, c.HttpDomain)

		aud, err := token.GetAudience()

		require.Equal(t, err, nil)
		require.Equal(t, aud, c.HttpDomain+c.HttpBasePath)

		sub, err := token.GetSubject()

		require.Equal(t, err, nil)
		require.Equal(t, sub, "opaque")
	})

}
