package jwt

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"aidanwoods.dev/go-paseto"

	"bookmarkd/internal/core"
)

type JwtResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	Expires      int    `json:"expires"`
}

func GetJwtTokenFromRequest(r *http.Request) string {
	tokenAuth := r.Header.Get("Authorization")
	tokenQuery := r.URL.Query().Get("token")

	if strings.HasPrefix(tokenAuth, "Bearer ") {
		return strings.TrimPrefix(tokenAuth, "Bearer ")
	}

	if tokenQuery != "" {
		return tokenQuery
	}

	return ""
}

func CreateJWT(config core.Config, sessionID int, refreshToken string) JwtResponse {

	expiration := time.Second * time.Duration(config.PasetoAccessTokenExpirationInSeconds)
	refreshExpiration := time.Second * time.Duration(config.PasetoRefreshTokenExpirationInSeconds)

	now := time.Now()

	// create an access token
	at := paseto.NewToken()
	at.SetIssuedAt(now)
	at.SetNotBefore(now)
	at.SetIssuer(config.HttpDomain)
	at.SetAudience(config.HttpDomain + config.HttpBasePath)
	at.SetExpiration(now.Add(time.Second * expiration))
	at.SetSubject(strconv.Itoa(sessionID))

	// create a refresh token
	rt := paseto.NewToken()
	rt.SetIssuedAt(now)
	rt.SetNotBefore(now)
	rt.SetAudience(config.HttpDomain + config.HttpBasePath)
	rt.SetIssuer(config.HttpDomain)
	rt.SetExpiration(now.Add(time.Second * refreshExpiration))
	rt.SetSubject(refreshToken)

	return JwtResponse{
		AccessToken:  at.V4Sign(config.PasetoSecretKey, nil),
		RefreshToken: rt.V4Sign(config.PasetoSecretKey, nil),
		TokenType:    "bearer",
		Expires:      config.PasetoAccessTokenExpirationInSeconds,
	}
}

func ValidateJWT(config core.Config, tokenString string) (*paseto.Token, error) {
	parser := paseto.NewParser()

	// this will fail if parsing failes, cryptographic checks fail, or validation rules fail
	return parser.ParseV4Public(config.PasetoPublicKey, tokenString, nil)
}
