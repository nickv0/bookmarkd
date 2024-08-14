package sqlite_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"bookmarkd"
	"bookmarkd/internal/core"
	"bookmarkd/internal/sqlite"
	"bookmarkd/utils/require"
)

func Test_BookmarkService_CreateBookmark(t *testing.T) {
	db := MustOpenDB(t)
	defer MustCloseDB(t, db)
	u := sqlite.NewUserStore(db)
	s := sqlite.NewSessionStore(db)
	b := sqlite.NewBookmarkStore(db)
	ctx := context.Background()

	user := MustCreateUser(t, ctx, u, &core.User{Username: "NAME0"})
	_, userCtx := MustCreateSession(t, ctx, s, user.ID)

	// Ensure a bookmark can be created by a user & a membership for the user is automatically created.
	t.Run("OK", func(t *testing.T) {
		bookmark := &core.Bookmark{Name: "mybookmark", Description: "Description", Url: "https://mybookmark"}
		err := b.CreateBookmark(userCtx, bookmark)

		require.Equal(t, err, nil)
		require.Equal(t, bookmark.ID, 1)
		require.Equal(t, bookmark.UserID, user.ID)
		require.Equal(t, bookmark.Name, "mybookmark")
		require.Equal(t, bookmark.Description, "Description")
		require.Equal(t, bookmark.Url, "https://mybookmark")
		require.Equal(t, bookmark.CreatedAt.IsZero(), false)
		require.Equal(t, bookmark.UpdatedAt.IsZero(), false)
		require.Equal(t, bookmark.User, nil)
	})

	// Ensure that creating a nameless bookmark returns an error.
	t.Run("ErrNameRequired", func(t *testing.T) {
		err := b.CreateBookmark(userCtx, &core.Bookmark{})

		require.AssertError(t, err)
		require.Equal(t, errors.Is(err, bookmarkd.ErrInvalidInput), true)
	})

	// Ensure that creating a bookmark with a long name returns an error.
	t.Run("ErrNameTooLong", func(t *testing.T) {
		err := b.CreateBookmark(userCtx, &core.Bookmark{Name: strings.Repeat("X", core.MaxBookmarkNameLen+1)})

		require.AssertError(t, err)
		require.Equal(t, errors.Is(err, bookmarkd.ErrInvalidInput), true)
	})

	t.Run("ErrDescriptionTooLong", func(t *testing.T) {
		err := b.CreateBookmark(userCtx, &core.Bookmark{Name: "bookmark", Description: strings.Repeat("X", core.MaxBookmarkDescriptionLen+1)})

		require.AssertError(t, err)
		require.Equal(t, errors.Is(err, bookmarkd.ErrInvalidInput), true)
	})

	t.Run("ErrUrlRequired", func(t *testing.T) {
		err := b.CreateBookmark(userCtx, &core.Bookmark{Name: "bookmark"})

		require.AssertError(t, err)
		require.Equal(t, errors.Is(err, bookmarkd.ErrInvalidInput), true)
	})

	t.Run("ErrUrlTooLong", func(t *testing.T) {
		err := b.CreateBookmark(userCtx, &core.Bookmark{Name: "bookmark", Url: strings.Repeat("X", core.MaxBookmarkUrlLen+1)})

		require.AssertError(t, err)
		require.Equal(t, errors.Is(err, bookmarkd.ErrInvalidInput), true)
	})

	// Ensure user is logged in when creating a bookmark.
	t.Run("ErrUserRequired", func(t *testing.T) {
		err := b.CreateBookmark(context.Background(), &core.Bookmark{})

		require.AssertError(t, err)
		require.Equal(t, errors.Is(err, bookmarkd.ErrUnauthorized), true)
	})
}

func Test_BookmarkService_UpdateBookmark(t *testing.T) {
	db := MustOpenDB(t)
	defer MustCloseDB(t, db)
	u := sqlite.NewUserStore(db)
	s := sqlite.NewSessionStore(db)
	b := sqlite.NewBookmarkStore(db)
	ctx := context.Background()

	user := MustCreateUser(t, ctx, u, &core.User{Username: "NAME0"})
	_, userCtx := MustCreateSession(t, ctx, s, user.ID)
	bookmark := MustCreateBookmark(t, userCtx, b, &core.Bookmark{Name: "NAME", Description: "", Url: "http://bookmark1`"})

	// Ensure a bookmark name can be updated.
	t.Run("OK", func(t *testing.T) {
		// Update bookmark.
		newName := "mybookmark2"
		uu, err := b.UpdateBookmark(userCtx, bookmark.ID, core.BookmarkUpdate{Name: &newName})

		require.Equal(t, err, nil)
		require.Equal(t, uu.Name, "mybookmark2")
	})
}

func Test_BookmarkService_FindBookmarks(t *testing.T) {
	db := MustOpenDB(t)
	defer MustCloseDB(t, db)
	u := sqlite.NewUserStore(db)
	s := sqlite.NewSessionStore(db)
	b := sqlite.NewBookmarkStore(db)
	ctx := context.Background()

	user := MustCreateUser(t, ctx, u, &core.User{Username: "NAME0"})
	_, userCtx := MustCreateSession(t, ctx, s, user.ID)
	MustCreateBookmark(t, userCtx, b, &core.Bookmark{Name: "NAME1", Description: "", Url: "http://bookmark1`"})
	MustCreateBookmark(t, userCtx, b, &core.Bookmark{Name: "NAME2", Description: "", Url: "http://bookmark2`"})
	MustCreateBookmark(t, userCtx, b, &core.Bookmark{Name: "NAME3", Description: "", Url: "http://bookmark3`"})

	// Ensure all bookmarks that are owned by user can be fetched.
	t.Run("Ok", func(t *testing.T) {
		a, n, err := b.FindBookmarks(userCtx, core.BookmarkFilter{})

		require.Equal(t, err, nil)
		require.Equal(t, n, 3)
		require.Equal(t, len(a), 3)
		require.Equal(t, a[0].Name, "NAME1")
		require.Equal(t, a[1].Name, "NAME2")
		require.Equal(t, a[2].Name, "NAME3")
	})
}

func TestBookmarkService_DeleteBookmark(t *testing.T) {
	db := MustOpenDB(t)
	defer MustCloseDB(t, db)
	u := sqlite.NewUserStore(db)
	s := sqlite.NewSessionStore(db)
	b := sqlite.NewBookmarkStore(db)
	ctx := context.Background()

	user := MustCreateUser(t, ctx, u, &core.User{Username: "NAME0"})
	_, userCtx := MustCreateSession(t, ctx, s, user.ID)
	deleteMe := MustCreateBookmark(t, userCtx, b, &core.Bookmark{Name: "NAME1", Description: "", Url: "http://bookmark1`"})
	MustCreateBookmark(t, userCtx, b, &core.Bookmark{Name: "NAME2", Description: "", Url: "http://bookmark2`"})

	// Ensure a bookmark can be deleted by the owner.
	t.Run("OK", func(t *testing.T) {
		bookmark, err := b.DeleteBookmark(userCtx, deleteMe.ID)
		require.Equal(t, err, nil)

		_, err = b.FindBookmarkByID(userCtx, bookmark.ID)
		require.Equal(t, errors.Is(err, bookmarkd.ErrNotFound), true)
	})
}

// MustCreateBookmark creates a bookmark in the database. Fatal on error.
func MustCreateBookmark(tb testing.TB, ctx context.Context, bookmarkStore core.BookmarkStore, bookmark *core.Bookmark) *core.Bookmark {
	tb.Helper()
	if err := bookmarkStore.CreateBookmark(ctx, bookmark); err != nil {
		tb.Fatal(err)
	}
	return bookmark
}
