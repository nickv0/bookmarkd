package core

import (
	"context"
	"fmt"
	"time"
	"unicode/utf8"

	"bookmarkd"
)

// Bookmark constants.
const (
	MaxBookmarkNameLen        = 255
	MaxBookmarkDescriptionLen = 255
	MaxBookmarkUrlLen         = 255
)

type Bookmark struct {
	ID int `json:"id"`

	// Owner of the bookmark. Only the owner may delete the bookmark.
	UserID string `json:"userID"`
	User   *User  `json:"user"`

	// Human-readable name of the bookmark.
	Name string `json:"name"`

	// Human-readable description of the bookmark.
	Description string `json:"description"`

	// http url of  the bookmark.
	Url string `json:"url"`

	// Timestamps for bookmark creation & last update.
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Validate returns an error if bookmark has invalid fields. Only performs basic validation.
func (d *Bookmark) Validate() error {
	if d.Name == "" {
		return fmt.Errorf("%w: bookmark name required", bookmarkd.ErrInvalidInput)
	} else if utf8.RuneCountInString(d.Name) > MaxBookmarkNameLen {
		return fmt.Errorf("%w: bookmark name too long", bookmarkd.ErrInvalidInput)
	} else if utf8.RuneCountInString(d.Description) > MaxBookmarkDescriptionLen {
		return fmt.Errorf("%w: bookmark description too long", bookmarkd.ErrInvalidInput)
	} else if d.Url == "" {
		return fmt.Errorf("%w: bookmark url required", bookmarkd.ErrInvalidInput)
	} else if utf8.RuneCountInString(d.Url) > MaxBookmarkUrlLen {
		return fmt.Errorf("%w: bookmark url too long", bookmarkd.ErrInvalidInput)
	} else if d.UserID == "" {
		return fmt.Errorf("%w: bookmark creator required", bookmarkd.ErrInvalidInput)
	}
	return nil
}

// CanEditBookmark returns true if the current user can edit the bookmark.
// Only the bookmark owner can edit the bookmark.
func CanEditBookmark(ctx context.Context, bookmark *Bookmark) bool {
	return bookmark.UserID == GetUserIDFromContext(ctx)
}

// BookmarkStore represents a service for managing bookmarks.
type BookmarkStore interface {
	FindBookmarkByID(ctx context.Context, id int) (*Bookmark, error)
	FindBookmarks(ctx context.Context, filter BookmarkFilter) ([]*Bookmark, int, error)
	CreateBookmark(ctx context.Context, bookmark *Bookmark) error
	UpdateBookmark(ctx context.Context, id int, update BookmarkUpdate) (*Bookmark, error)
	DeleteBookmark(ctx context.Context, id int) (*Bookmark, error)
}

// BookmarkFilter represents a filter used by FindBookmarks().
type BookmarkFilter struct {
	// Filtering fields.
	ID *int `json:"id"`

	// Restrict to subset of range.
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

// BookmarkUpdate represents a set of fields to update on a bookmark.
type BookmarkUpdate struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Url         *string `json:"url"`
}
