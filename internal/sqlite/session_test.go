package sqlite_test

import (
	"context"
	"errors"
	"testing"

	"bookmarkd"
	"bookmarkd/internal/core"
	"bookmarkd/internal/sqlite"
	"bookmarkd/utils/require"
)

func TestUserStore_CreateSession(t *testing.T) {
	db := MustOpenDB(t)
	defer MustCloseDB(t, db)
	u := sqlite.NewUserStore(db)
	s := sqlite.NewSessionStore(db)
	ctx := context.Background()

	user := MustCreateUser(t, ctx, u, &core.User{Username: "NAME0"})

	// Ensure user can be created.
	t.Run("returns Session", func(t *testing.T) {
		session := core.Session{UserID: user.ID}
		err := s.CreateSession(ctx, &session)

		require.Equal(t, err, nil)
		require.Equal(t, session.UserID, user.ID)
		require.Equal(t, session.CreatedAt.IsZero(), false)
		require.Equal(t, session.UpdatedAt.IsZero(), false)
	})

	// Ensure an error is returned if user name is not set.
	t.Run("returns error EINTERNAL if session is missing userID", func(t *testing.T) {
		session := core.Session{}
		err := s.CreateSession(ctx, &session)

		require.Equal(t, err == nil, false)
		require.Equal(t, errors.Is(err, bookmarkd.ErrInternal), true)
	})
}

func TestUserStore_RefreshSession(t *testing.T) {
	db := MustOpenDB(t)
	defer MustCloseDB(t, db)
	u := sqlite.NewUserStore(db)
	s := sqlite.NewSessionStore(db)
	ctx := context.Background()

	user := MustCreateUser(t, ctx, u, &core.User{Username: "NAME0"})
	session, userCtx := MustCreateSession(t, ctx, s, user.ID)

	t.Run("OK", func(t *testing.T) {
		upd, err := s.RefreshSession(userCtx, "refresh_token")

		require.Equal(t, err, nil)
		require.Equal(t, upd.ID, session.ID)
		require.NotEqual(t, upd.RefreshToken, session.RefreshToken)
		require.Equal(t, upd.ExpiresAt.IsZero(), false)
	})

	t.Run("ErrInvalid", func(t *testing.T) {
		upd, err := s.RefreshSession(userCtx, "refresh_token")

		require.Equal(t, err, nil)
		require.Equal(t, upd.ID, session.ID)
		require.NotEqual(t, upd.RefreshToken, session.RefreshToken)
		require.Equal(t, upd.ExpiresAt.IsZero(), false)
	})
}

func TestUserStore_DeleteSession(t *testing.T) {
	db := MustOpenDB(t)
	defer MustCloseDB(t, db)
	u := sqlite.NewUserStore(db)
	s := sqlite.NewSessionStore(db)
	ctx := context.Background()

	user := MustCreateUser(t, ctx, u, &core.User{Username: "NAME0"})
	session, userCtx := MustCreateSession(t, ctx, s, user.ID)

	t.Run("OK", func(t *testing.T) {
		err := s.DeleteSession(userCtx, session.ID)
		require.Equal(t, err, nil)

	})
}

func TestUserStore_FindSessionById(t *testing.T) {
	db := MustOpenDB(t)
	defer MustCloseDB(t, db)
	u := sqlite.NewUserStore(db)
	s := sqlite.NewSessionStore(db)
	ctx := context.Background()

	user := MustCreateUser(t, ctx, u, &core.User{Username: "NAME0"})
	session, userCtx := MustCreateSession(t, ctx, s, user.ID)

	t.Run("returns Session", func(t *testing.T) {
		result, err := s.FindSessionByID(userCtx, session.ID)

		require.Equal(t, err, nil)
		require.Equal(t, result.ID, session.ID)
		require.Equal(t, result.UserID, user.ID)
	})
}

func TestUserStore_FindSessions(t *testing.T) {
	db := MustOpenDB(t)
	defer MustCloseDB(t, db)
	u := sqlite.NewUserStore(db)
	s := sqlite.NewSessionStore(db)
	ctx := context.Background()

	user := MustCreateUser(t, ctx, u, &core.User{Username: "NAME0"})
	session, userCtx := MustCreateSession(t, ctx, s, user.ID)

	t.Run("returns 1 Session", func(t *testing.T) {
		sessions, n, err := s.FindSessions(userCtx, core.SessionFilter{UserID: &user.ID})

		require.Equal(t, err, nil)
		require.Equal(t, len(sessions), 1)
		require.Equal(t, n, 1)
		require.Equal(t, sessions[0].ID, session.ID)
		require.Equal(t, sessions[0].UserID, user.ID)
	})
}

func MustCreateSession(tb testing.TB, ctx context.Context, s core.SessionStore, userID string) (*core.Session, context.Context) {
	tb.Helper()

	session := &core.Session{UserID: userID}
	err := s.CreateSession(ctx, session)
	if err != nil {
		tb.Fatal(err)
	}

	return session, core.NewContextWithSession(ctx, core.SessionContext{
		UserID:    session.UserID,
		SessionID: session.ID,
		ExpiresAt: session.ExpiresAt,
	})
}
