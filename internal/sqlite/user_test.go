package sqlite_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"bookmarkd"
	"bookmarkd/internal/core"
	"bookmarkd/internal/sqlite"
	"bookmarkd/utils/require"
)

func Test_UserStore_CreateUser(t *testing.T) {
	db := MustOpenDB(t)
	defer MustCloseDB(t, db)
	s := sqlite.NewUserStore(db)
	ctx := context.Background()

	var firstUser core.User

	t.Run("OK", func(t *testing.T) {
		u := &core.User{Username: "susy", Seed: "susy_seed"}
		err := s.CreateUser(ctx, u)

		require.Equal(t, err, nil)
		require.Equal(t, uuid.Validate(u.ID), nil)
		require.Equal(t, u.CreatedAt.IsZero(), false)
		require.Equal(t, u.UpdatedAt.IsZero(), false)

		// Fetch user from database & compare.
		other, err := s.FindUserByID(ctx, u.ID)

		require.Equal(t, err, nil)
		require.DeepEqual(t, other, u)

		firstUser = *u
	})

	t.Run("Users have different ID", func(t *testing.T) {
		// Create second user with email.
		u := &core.User{Username: "jane", Seed: "jane_seed"}
		err := s.CreateUser(ctx, u)

		require.Equal(t, err, nil)
		require.Equal(t, uuid.Validate(u.ID), nil)

		require.Equal(t, u.ID == firstUser.ID, false)
	})

	// Ensure an error is returned if user name is not set.
	t.Run("Throws Error if missing username", func(t *testing.T) {
		err := s.CreateUser(ctx, &core.User{})

		require.Equal(t, err == nil, false)
		require.Equal(t, errors.Is(err, bookmarkd.ErrInvalidInput), true)
	})
}

func Test_UserStore_UpdateUser(t *testing.T) {
	db := MustOpenDB(t)
	defer MustCloseDB(t, db)

	s := sqlite.NewUserStore(db)
	ss := sqlite.NewSessionStore(db)

	user0 := MustCreateUser(t, context.Background(), s, &core.User{
		Username: "susy",
	})
	user1 := MustCreateUser(t, context.Background(), s, &core.User{
		Username: "karen",
	})

	_, ctx0 := MustCreateSession(t, context.Background(), ss, user0.ID)

	t.Run("OK", func(t *testing.T) {
		// Update user.
		newName := "jill"
		uu, err := s.UpdateUser(ctx0, user0.ID, core.UserUpdate{
			Username: &newName,
		})

		require.Equal(t, err, nil)
		require.Equal(t, uu.Username, "jill")

		// Fetch user from database & compare.
		other, err := s.FindUserByID(context.Background(), user0.ID)
		require.Equal(t, err, nil)
		require.DeepEqual(t, other, uu)
	})

	// Ensure updating a user is restricted only to the current user.
	t.Run("ErrUnauthorized", func(t *testing.T) {

		// Update user as another user.
		newName := "NEWNAME"
		_, err := s.UpdateUser(ctx0, user1.ID, core.UserUpdate{Username: &newName})

		require.AssertError(t, err)
		require.Equal(t, errors.Is(err, bookmarkd.ErrUnauthorized), true)
	})
}

func Test_UserStore_DeleteUser(t *testing.T) {
	db := MustOpenDB(t)
	defer MustCloseDB(t, db)

	s := sqlite.NewUserStore(db)
	ss := sqlite.NewSessionStore(db)

	user0 := MustCreateUser(t, context.Background(), s, &core.User{Username: "john"})
	user1 := MustCreateUser(t, context.Background(), s, &core.User{Username: "beth"})

	_, ctx0 := MustCreateSession(t, context.Background(), ss, user0.ID)
	_, ctx1 := MustCreateSession(t, context.Background(), ss, user1.ID)

	t.Run("OK", func(t *testing.T) {
		// Delete user
		_, err := s.DeleteUser(ctx0, user0.ID)
		require.Equal(t, err, nil)

		// Ensure it is actually deleted
		_, err = s.FindUserByID(ctx0, user0.ID)
		require.AssertError(t, err)
		require.Equal(t, errors.Is(err, bookmarkd.ErrNotFound), true)
	})

	// Ensure an error is returned if deleting a non-existent user.
	t.Run("ErrNotFound", func(t *testing.T) {
		_, err := s.DeleteUser(context.Background(), "")

		require.AssertError(t, err)
		require.Equal(t, errors.Is(err, bookmarkd.ErrNotFound), true)
	})

	// Ensure deleting a user is restricted only to the current user.
	t.Run("ErrUnauthorized", func(t *testing.T) {
		_, err := s.DeleteUser(ctx1, user0.ID)

		require.AssertError(t, err)
		require.Equal(t, errors.Is(err, bookmarkd.ErrNotFound), true)
	})
}

func Test_UserStore_FindUser(t *testing.T) {
	db := MustOpenDB(t)
	defer MustCloseDB(t, db)

	s := sqlite.NewUserStore(db)
	ss := sqlite.NewSessionStore(db)

	user0 := MustCreateUser(t, context.Background(), s, &core.User{Username: "john"})
	user1 := MustCreateUser(t, context.Background(), s, &core.User{Username: "beth"})

	_, ctx0 := MustCreateSession(t, context.Background(), ss, user0.ID)
	_, ctx1 := MustCreateSession(t, context.Background(), ss, user1.ID)

	t.Run("Returns User", func(t *testing.T) {

		u, err := s.FindUserByID(ctx0, user0.ID)

		require.Equal(t, err, nil)
		require.Equal(t, u.ID, user0.ID)
		require.Equal(t, u.Username, user0.Username)
		require.Equal(t, len(u.Sessions), 1)
		require.Equal(t, u.CreatedAt, user0.CreatedAt)
		require.Equal(t, u.UpdatedAt, user0.UpdatedAt)
	})

	// Ensure an error is returned if fetching a non-existent user.
	t.Run("Invalid User Returns Error", func(t *testing.T) {
		db := MustOpenDB(t)
		defer MustCloseDB(t, db)

		s := sqlite.NewUserStore(db)

		_, err := s.FindUserByID(ctx1, user0.ID)

		require.AssertError(t, err)
		require.Equal(t, errors.Is(err, bookmarkd.ErrNotFound), true)

	})
}

func Test_UserStore_FindUsers(t *testing.T) {
	db := MustOpenDB(t)
	defer MustCloseDB(t, db)

	s := sqlite.NewUserStore(db)
	ss := sqlite.NewSessionStore(db)

	user0 := MustCreateUser(t, context.Background(), s, &core.User{Username: "john"})
	MustCreateUser(t, context.Background(), s, &core.User{Username: "beth"})

	_, ctx0 := MustCreateSession(t, context.Background(), ss, user0.ID)

	// Ensure users can be fetched by email address.
	t.Run("Username", func(t *testing.T) {
		username := "john"
		a, n, err := s.FindUsers(ctx0, core.UserFilter{Username: &username})
		require.Equal(t, err, nil)
		require.Equal(t, len(a), 1)
		require.Equal(t, a[0].Username, username)
		require.Equal(t, n, 1)
	})
}

func MustCreateUser(tb testing.TB, ctx context.Context, s core.UserStore, user *core.User) *core.User {
	tb.Helper()
	user.Seed = "random_seed"
	if err := s.CreateUser(ctx, user); err != nil {
		tb.Fatal(err)
	}

	return user
}
