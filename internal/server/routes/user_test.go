package routes_test

import (
	"bytes"
	"context"
	"encoding/base32"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/cristalhq/otp"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"bookmarkd/internal/core"
	"bookmarkd/internal/mock"
	"bookmarkd/internal/server/routes"
	"bookmarkd/utils/require"
)

const (
	UserID          = "01911e85-970b-73e7-b916-24e91b3d2fc6"
	UserName        = "john_smith@google.com"
	UserDisplayName = "john_smith"
)

var regID = uuid.MustParse("e0c7249e-7409-4731-837a-0bb31d71bd86")
var seed = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString([]byte(regID.String()))

var User = core.User{
	ID:        "01911e85-970b-73e7-b916-24e91b3d2fc6",
	Username:  "john_smith@google.com",
	Seed:      seed,
	CreatedAt: time.Now(),
	UpdatedAt: time.Now(),
	Sessions:  []*core.Session{},
}

func Test_login(t *testing.T) {
	config, _ := core.NewConfig(func(string) string { return "" })

	username := "john_smith@google.com"

	registrationStore := mock.RegistrationStore{}
	mockEventService := mock.EventService{}
	mockBookmarkStore := mock.BookmarkStore{}
	mockSessionStore := mock.SessionStore{}
	mockUserStore := mock.UserStore{}

	r := chi.NewRouter()
	routes.AddRoutes(r, config, &registrationStore, &mockBookmarkStore, &mockEventService, &mockSessionStore, &mockUserStore)

	// setup server mocks
	registrationStore.StartRegistrationSessionFn = func(username string) (*core.Registration, error) {
		return &core.Registration{
			ID:        regID,
			Username:  username,
			Seed:      seed,
			ExpiresAt: time.Now().Add(5 * time.Minute),
		}, nil
	}
	registrationStore.FindRegistrationSessionByIDFn = func(id string) (*core.Registration, error) {
		return &core.Registration{
			ID:        regID,
			Username:  username,
			Seed:      seed,
			ExpiresAt: time.Now().Add(5 * time.Minute),
		}, nil
	}
	registrationStore.DeleteRegistrationSessionFn = func(id string) error {
		return nil
	}
	mockUserStore.CreateUserFn = func(ctx context.Context, user *core.User) error {
		user.ID = User.ID
		user.Seed = User.Seed
		user.Username = User.Username
		user.CreatedAt = User.CreatedAt
		user.UpdatedAt = User.UpdatedAt
		user.Sessions = User.Sessions
		return nil
	}

	mockUserStore.FindUserByUsernameFn = func(ctx context.Context, username string) (*core.User, error) {
		return &User, nil
	}

	mockSessionStore.CreateSessionFn = func(ctx context.Context, session *core.Session) error {
		session.RefreshToken = "opaque"
		return nil
	}

	// Register
	startResponse, err := startRegister(t, r, username)

	require.Equal(t, err, nil)
	require.Equal(t, len(startResponse.RegistrationID) > 0, true)
	require.Equal(t, len(startResponse.TotpUrl) > 0, true)

	u, err := url.Parse(startResponse.TotpUrl)
	require.Equal(t, err, nil)

	totp, err := otp.NewTOTP(otp.TOTPConfig{
		Algo:   config.TotpAlgo,
		Digits: config.TotpDigits,
		Issuer: config.TotpIssuer,
		Period: config.TotpPeriod,
		Skew:   config.TotpSkew,
	})

	require.Equal(t, err, nil)

	str := u.Query().Get("secret")
	secret, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(str)
	if err != nil {
		t.Fatal(err)
	}

	now := time.Now().UTC()
	passcode, err := totp.GenerateCode(string(secret), now)
	require.Equal(t, err, nil)

	if err := totp.Validate(passcode, time.Now(), string(secret)); err != nil {
		t.Fatal(err)
	}

	finishResponse, err := finishRegister(t, r, startResponse.RegistrationID, passcode)
	require.Equal(t, err, nil)

	require.Equal(t, len(finishResponse.AccessToken) > 0, true)
	require.Equal(t, len(finishResponse.RefreshToken) > 0, true)
	require.Equal(t, finishResponse.TokenType, "bearer")
	require.Equal(t, finishResponse.Expires, 300)

	loginResponse, err := login(t, r, username, passcode)
	require.Equal(t, err, nil)

	require.Equal(t, len(loginResponse.AccessToken) > 0, true)
	require.Equal(t, len(loginResponse.RefreshToken) > 0, true)
	require.Equal(t, loginResponse.TokenType, "bearer")
	require.Equal(t, loginResponse.Expires, 300)
}

// Register
func startRegister(t *testing.T, m *chi.Mux, username string) (*routes.AuthRegisterPostPayload, error) {
	b := routes.AuthRegisterPostInput{
		Username: username,
	}

	j, err := json.Marshal(b)
	if err != nil {
		return nil, err
	}

	bodyReader := bytes.NewReader(j)
	req, _ := http.NewRequest("POST", "/auth/register", bodyReader)

	rr := httptest.NewRecorder()
	m.ServeHTTP(rr, req)

	require.Equal(t, rr.Code, http.StatusOK)

	var resp routes.AuthRegisterPostPayload
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func finishRegister(t *testing.T, m *chi.Mux, registrationID string, totp string) (*routes.AuthRegisterConfirmPostPayload, error) {
	b := routes.AuthRegisterConfirmPostInput{
		RegistrationID: registrationID,
		Totp:           totp,
	}

	j, err := json.Marshal(b)
	if err != nil {
		return nil, err
	}

	bodyReader := bytes.NewReader(j)
	req, _ := http.NewRequest("POST", "/auth/register/confirm", bodyReader)

	rr := httptest.NewRecorder()
	m.ServeHTTP(rr, req)

	t.Log(rr.Body)
	require.Equal(t, rr.Code, http.StatusOK)

	var resp routes.AuthRegisterConfirmPostPayload
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// Login
func login(t *testing.T, m *chi.Mux, username string, totp string) (*routes.AuthLoginPostPayload, error) {
	b := routes.AuthLoginPostInput{
		Username: username,
		Totp:     totp,
	}

	j, err := json.Marshal(b)
	if err != nil {
		return nil, err
	}

	bodyReader := bytes.NewReader(j)
	req, _ := http.NewRequest("POST", "/auth/login", bodyReader)

	rr := httptest.NewRecorder()
	m.ServeHTTP(rr, req)

	require.Equal(t, rr.Code, http.StatusOK)

	var resp routes.AuthLoginPostPayload
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}
