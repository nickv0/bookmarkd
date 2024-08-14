package core

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"

	"aidanwoods.dev/go-paseto"
	"github.com/cristalhq/otp"
)

type Config struct {
	Test bool
	// sqlite
	DbDsn string
	// http
	HttpAddress          string
	HttpBasePath         string
	HttpDomain           string
	HttpPort             string
	HttpTimeoutInSeconds int
	// jwt
	PasetoSecretKey                       paseto.V4AsymmetricSecretKey
	PasetoPublicKey                       paseto.V4AsymmetricPublicKey
	PasetoAccessTokenExpirationInSeconds  int
	PasetoRefreshTokenExpirationInSeconds int
	// error logging
	RollbarToken string
	// totp settings
	TotpAlgo   otp.Algorithm
	TotpDigits uint
	TotpIssuer string
	TotpPeriod uint64
	TotpSkew   uint
}

func NewConfig(getenv func(string) string) (Config, error) {
	sk := paseto.NewV4AsymmetricSecretKey()

	c := Config{
		Test:                                  false,
		DbDsn:                                 ".database/database.sqlite",
		HttpAddress:                           "localhost",
		HttpPort:                              "8080",
		HttpDomain:                            "http://localhost:8080",
		HttpBasePath:                          "/api",
		HttpTimeoutInSeconds:                  60,
		PasetoSecretKey:                       sk,
		PasetoPublicKey:                       sk.Public(),
		PasetoAccessTokenExpirationInSeconds:  300,
		PasetoRefreshTokenExpirationInSeconds: 1200,
		RollbarToken:                          "",
		TotpAlgo:                              otp.AlgorithmSHA1,
		TotpDigits:                            8,
		TotpIssuer:                            "bookmarkd",
		TotpPeriod:                            30,
		TotpSkew:                              1,
	}

	// update config values from getenv function

	if s := getenv("BOOKMARKD_TEST"); s != "" {
		c.Test = true
	}

	if s := getenv("BOOKMARKD_DSN"); s != "" {
		if dsn, err := ExpandDSN(s); err != nil {
			return c, err
		} else {
			c.DbDsn = dsn
		}
	}

	if s := getenv("BOOKMARKD_HTTP_ADDRESS"); s != "" {
		c.HttpAddress = s
	}

	if s := getenv("BOOKMARKD_HTTP_PORT"); s != "" {
		c.HttpPort = s
	}

	if s := getenv("BOOKMARKD_HTTP_DOMAIN"); s != "" {
		c.HttpDomain = s
	}

	if s := getenv("BOOKMARKD_HTTP_BASE_PATH"); s != "" {
		c.HttpBasePath = s
	}

	if s := getenv("BOOKMARKD_PASETO_ACESS_TOKEN_EXPIRATION_IN_SECONDS"); s != "" {
		if i, err := strconv.Atoi(s); err != nil {
			return c, err
		} else {
			c.PasetoAccessTokenExpirationInSeconds = i
		}
	}

	if s := getenv("BOOKMARKD_PASETO_REFRESH_TOKEN_EXPIRATION_IN_SECONDS"); s != "" {
		if i, err := strconv.Atoi(s); err != nil {
			return c, err
		} else {
			c.PasetoRefreshTokenExpirationInSeconds = i
		}
	}

	if s := getenv("BOOKMARKD_PASETO_SECRET"); s != "" {
		if h, err := paseto.NewV4AsymmetricSecretKeyFromHex(s); err != nil {
			return c, err
		} else {
			c.PasetoSecretKey = h
			c.PasetoPublicKey = h.Public()
		}
	}

	if s := getenv("BOOKMARKD_ROLLBAR_TOKEN"); s != "" {
		c.RollbarToken = s
	}

	if s := getenv("BOOKMARKD_TOTP_DIGITS"); s != "" {
		if i, err := strconv.Atoi(s); err != nil {
			return c, err
		} else {
			c.TotpDigits = uint(i)
		}
	}

	if s := getenv("BOOKMARKD_TOTP_ISSUER"); s != "" {
		c.TotpIssuer = s
	}

	return c, nil
}

// expand returns path using tilde expansion. This means that a file path that
// begins with the "~" will be expanded to prefix the user's home directory.
func ExpandPath(path string) (string, error) {
	// Ignore if path has no leading tilde.
	if path != "~" && !strings.HasPrefix(path, "~"+string(os.PathSeparator)) {
		return path, nil
	}

	// Fetch the current user to determine the home path.
	u, err := user.Current()
	if err != nil {
		return path, err
	} else if u.HomeDir == "" {
		return path, fmt.Errorf("home directory unset")
	}

	if path == "~" {
		return u.HomeDir, nil
	}
	return filepath.Join(u.HomeDir, strings.TrimPrefix(path, "~"+string(os.PathSeparator))), nil
}

// expandDSN expands a datasource name. Ignores in-memory databases.
func ExpandDSN(dsn string) (string, error) {
	if dsn == ":memory:" {
		return dsn, nil
	}
	return ExpandPath(dsn)
}
