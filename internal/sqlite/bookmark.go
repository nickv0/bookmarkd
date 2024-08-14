package sqlite

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"bookmarkd"
	"bookmarkd/internal/core"
)

// Ensure service implements interface.
var _ core.BookmarkStore = (*BookmarkStore)(nil)

// BookmarkStore represents a service for managing bookmarks.
type BookmarkStore struct {
	db *DB
}

// NewBookmarkStore returns a new instance of BookmarkStore.
func NewBookmarkStore(db *DB) *BookmarkStore {
	return &BookmarkStore{db: db}
}

// FindBookmarkByID retrieves a single bookmark by ID along with associated memberships.
// Only the bookmark owner & members can see a bookmark. Returns ENOTFOUND if bookmark does
// not exist or user does not have permission to view it.
func (s *BookmarkStore) FindBookmarkByID(ctx context.Context, id int) (*core.Bookmark, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Fetch bookmark object and attach owner user.
	bookmark, err := findBookmarkByID(ctx, tx, id)
	if err != nil {
		return nil, err
	}
	return bookmark, nil
}

// FindBookmarks retrieves a list of bookmarks based on a filter. Only returns bookmarks
// that the user owns or is a member of.
//
// Also returns a count of total matching bookmarks which may different from the
// number of returned bookmarks if the  "Limit" field is set.
func (s *BookmarkStore) FindBookmarks(ctx context.Context, filter core.BookmarkFilter) ([]*core.Bookmark, int, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, 0, err
	}
	defer tx.Rollback()

	// Fetch list of matching bookmark objects.
	bookmarks, n, err := findBookmarks(ctx, tx, filter)
	if err != nil {
		return bookmarks, n, err
	}

	return bookmarks, n, nil
}

// CreateBookmark creates a new bookmark and assigns the current user as the owner.
// The owner will automatically be added as a member of the new bookmark.
func (s *BookmarkStore) CreateBookmark(ctx context.Context, bookmark *core.Bookmark) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Create bookmark and attach associated owner user.
	if err := createBookmark(ctx, tx, bookmark); err != nil {
		return err
	}
	return tx.Commit()
}

// UpdateBookmark updates an existing bookmark by ID. Only the bookmark owner can update a bookmark.
// Returns the new bookmark state even if there was an error during update.
//
// Returns ENOTFOUND if bookmark does not exist. Returns EUNAUTHORIZED if user
// is not the bookmark owner.
func (s *BookmarkStore) UpdateBookmark(ctx context.Context, id int, upd core.BookmarkUpdate) (*core.Bookmark, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Update the bookmark object and attach associated user to returned bookmark.
	bookmark, err := updateBookmark(ctx, tx, id, upd)
	if err != nil {
		return bookmark, err
	}
	return bookmark, tx.Commit()
}

// DeleteBookmark permanently removes a bookmark by ID. Only the bookmark owner may delete
// a bookmark. Returns ENOTFOUND if bookmark does not exist. Returns EUNAUTHORIZED if
// user is not the bookmark owner.
func (s *BookmarkStore) DeleteBookmark(ctx context.Context, id int) (*core.Bookmark, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	bookmark, err := deleteBookmark(ctx, tx, id)
	if err != nil {
		return bookmark, err
	}
	return bookmark, tx.Commit()
}

// findBookmarkByID is a helper function to retrieve a bookmark by ID.
// Returns ENOTFOUND if bookmark doesn't exist.
func findBookmarkByID(ctx context.Context, tx *Tx, id int) (*core.Bookmark, error) {
	bookmarks, _, err := findBookmarks(ctx, tx, core.BookmarkFilter{ID: &id})
	if err != nil {
		return nil, fmt.Errorf("find bookmarks: %w", err)
	} else if len(bookmarks) == 0 {
		return nil, bookmarkd.ErrNotFound
	}
	return bookmarks[0], nil
}

// findBookmarks retrieves a list of matching bookmarks. Also returns a total matching
// count which may different from the number of results if filter.Limit is set.
func findBookmarks(ctx context.Context, tx *Tx, filter core.BookmarkFilter) (_ []*core.Bookmark, n int, err error) {
	// Build WHERE clause. Each part of the WHERE clause is AND-ed together.
	// Values are appended to an arg list to avoid SQL injection.
	where, args := []string{"1 = 1"}, []interface{}{}
	if v := filter.ID; v != nil {
		where, args = append(where, "id = ?"), append(args, *v)
	}

	// Limit to bookmarks user owns.
	userID := core.GetUserIDFromContext(ctx)
	where, args = append(where, "user_id = ?"), append(args, userID)

	// Execue query with limiting WHERE clause and LIMIT/OFFSET injected.
	rows, err := tx.QueryContext(ctx, `
		SELECT 
		  id,
		  user_id,
		  name,
		  description,
		  url,
		  created_at,
		  updated_at,
		  COUNT(*) OVER()
		FROM bookmarks
		WHERE `+strings.Join(where, " AND ")+`
		ORDER BY id ASC
		`+FormatLimitOffset(filter.Limit, filter.Offset),
		args...,
	)
	if err != nil {
		return nil, n, fmt.Errorf("db select bookmarks: %w", FormatError(err))
	}
	defer rows.Close()

	// Iterate over rows and deserialize into Bookmark objects.
	bookmarks := make([]*core.Bookmark, 0)
	for rows.Next() {
		var bookmark core.Bookmark
		if err := rows.Scan(
			&bookmark.ID,
			&bookmark.UserID,
			&bookmark.Name,
			&bookmark.Description,
			&bookmark.Url,
			(*NullTime)(&bookmark.CreatedAt),
			(*NullTime)(&bookmark.UpdatedAt),
			&n,
		); err != nil {
			return nil, 0, fmt.Errorf("db scan bookmark row: %w", FormatError(err))
		}
		bookmarks = append(bookmarks, &bookmark)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("db bookmark rows: %w", FormatError(err))
	}

	return bookmarks, n, nil
}

// createBookmark creates a new bookmark.
func createBookmark(ctx context.Context, tx *Tx, bookmark *core.Bookmark) error {
	// Assign bookmark to the current user.
	// Return an error if the user is not currently logged in.
	userID := core.GetUserIDFromContext(ctx)
	if userID == "" {
		return bookmarkd.ErrUnauthorized
	}
	bookmark.UserID = userID

	// Set timestamps to current time.
	bookmark.CreatedAt = tx.Now()
	bookmark.UpdatedAt = bookmark.CreatedAt

	// Perform basic field validation
	if err := bookmark.Validate(); err != nil {
		return err
	}

	// race condition, if user session was valid, then deleted
	// we should check to ensure the user exists before adding a bookmark
	// Actaully this should throw a foreign key constraint failed if the user
	// was deleted from the database.
	/*
		if _, err := findUserByID(ctx, tx, bookmark.UserID); err != nil {
			return fmt.Errorf("create bookmark error: %w", err)
		}
	*/

	// Insert row into database.
	result, err := tx.ExecContext(ctx, `
		INSERT INTO bookmarks (
			user_id,
		  name,
		  description,
		  url,
		  created_at,
		  updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?)
	`,
		bookmark.UserID,
		bookmark.Name,
		bookmark.Description,
		bookmark.Url,
		(*NullTime)(&bookmark.CreatedAt),
		(*NullTime)(&bookmark.UpdatedAt),
	)
	if err != nil {
		return fmt.Errorf("db insert bookmark: %w", FormatError(err))
	}

	// Read back new bookmark ID into caller argument.
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("db get bookmark id: %w", FormatError(err))
	}
	bookmark.ID = int(id)

	return nil
}

// updateBookmark updates a bookmark by ID. Returns the new state of the bookmark after update.
func updateBookmark(ctx context.Context, tx *Tx, id int, upd core.BookmarkUpdate) (*core.Bookmark, error) {
	// Fetch current object state. Return an error if current user is not owner.
	bookmark, err := findBookmarkByID(ctx, tx, id)
	if err != nil {
		return bookmark, err
	} else if !core.CanEditBookmark(ctx, bookmark) {
		return bookmark, bookmarkd.ErrUnauthorized
	}

	// Update fields, if set.
	if v := upd.Name; v != nil {
		bookmark.Name = *v
	}
	if v := upd.Description; v != nil {
		bookmark.Description = *v
	}
	if v := upd.Url; v != nil {
		bookmark.Url = *v
	}
	bookmark.UpdatedAt = tx.Now()

	// Perform basic field validation.
	if err := bookmark.Validate(); err != nil {
		return bookmark, err
	}

	// Execute update query.
	if _, err := tx.ExecContext(ctx, `
		UPDATE bookmarks
		SET name = ?,
				description = ?,
		  	url = ?,
		    updated_at = ?
		WHERE id = ?
	`,
		bookmark.Name,
		bookmark.Description,
		bookmark.Url,
		(*NullTime)(&bookmark.UpdatedAt),
		id,
	); err != nil {
		// should we return inconsitent data or nil?
		return bookmark, fmt.Errorf("db update bookmark: %w", FormatError(err))
	}

	if upd.Name != nil {
		if err := publishBookmarkEvent(ctx, tx, id, core.Event{
			Type: core.EventTypeBookmarkNameChanged,
			Payload: &core.EventTypeBookmarkNameChangedPayload{
				ID:        bookmark.ID,
				Name:      bookmark.Name,
				UpdatedAt: bookmark.UpdatedAt,
			},
		}); err != nil {
			return bookmark, fmt.Errorf("publish bookmark name event: %w", err)
		}
	}

	if upd.Description != nil {
		if err := publishBookmarkEvent(ctx, tx, id, core.Event{
			Type: core.EventTypeBookmarkDescriptionChanged,
			Payload: &core.EventTypeBookmarkDescriptionChangedPayload{
				ID:          bookmark.ID,
				Description: bookmark.Description,
				UpdatedAt:   bookmark.UpdatedAt,
			},
		}); err != nil {
			return bookmark, fmt.Errorf("publish bookmark description event: %w", err)
		}
	}

	if upd.Url != nil {
		if err := publishBookmarkEvent(ctx, tx, id, core.Event{
			Type: core.EventTypeBookmarkUrlChanged,
			Payload: &core.EventTypeBookmarkUrlChangedPayload{
				ID:        bookmark.ID,
				Url:       bookmark.Url,
				UpdatedAt: bookmark.UpdatedAt,
			},
		}); err != nil {
			return bookmark, fmt.Errorf("publish bookmark url event: %w", err)
		}
	}

	return bookmark, nil
}

// deleteBookmark permanently deletes a bookmark by ID. Returns EUNAUTHORIZED if user
// does not own the bookmark.
func deleteBookmark(ctx context.Context, tx *Tx, id int) (*core.Bookmark, error) {
	// Verify object exists & the current user is the owner.
	bookmark, err := findBookmarkByID(ctx, tx, id)
	if err != nil {
		return bookmark, err
	} else if !core.CanEditBookmark(ctx, bookmark) {
		return bookmark, bookmarkd.ErrUnauthorized
	}

	// Remove row from database.
	if _, err := tx.ExecContext(ctx, `DELETE FROM bookmarks WHERE id = ?`, id); err != nil {
		return bookmark, fmt.Errorf("db delete bookmark: %w", FormatError(err))
	}
	return bookmark, nil
}

// publishBookmarkEvent publishes event to the bookmark members.
func publishBookmarkEvent(ctx context.Context, tx *Tx, id int, event core.Event) error {
	// Find owner of the bookmark.
	rows, err := tx.QueryContext(ctx, `SELECT user_id FROM bookmarks WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("db select user_id: %w", FormatError(err))
	}
	defer rows.Close()

	// Iterate over users and publish event.
	for rows.Next() {
		var userID uuid.UUID
		if err := rows.Scan(&userID); err != nil {
			return fmt.Errorf("generate uuid for event: %w", (err))
		}
		tx.PublishEvent(userID.String(), event)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("db select user_id rows: %w", FormatError(err))
	}
	return nil
}
