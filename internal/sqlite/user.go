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
var _ core.UserStore = (*UserStore)(nil)

// UserStore represents a service for managing users.
type UserStore struct {
	db *DB
}

// NewUserStore returns a new instance of UserStore.
func NewUserStore(db *DB) *UserStore {
	return &UserStore{db: db}
}

// FindUserByID retrieves a user by ID along with their associated auth objects.
// Returns ENOTFOUND if user does not exist.
func (s *UserStore) FindUserByID(ctx context.Context, id string) (*core.User, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	user, err := findUserByID(ctx, tx, id)
	if err != nil {
		return nil, err
	} else if err := attachUserAssociations(ctx, tx, user); err != nil {
		return user, err
	}
	return user, nil
}

func (s *UserStore) FindUserByUsername(ctx context.Context, username string) (*core.User, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	user, err := findUserByUsername(ctx, tx, username)
	if err != nil {
		return nil, err
	} else if err := attachUserAssociations(ctx, tx, user); err != nil {
		return user, err
	}
	return user, nil
}

// FindUsers retrieves a list of users by filter. Also returns total count of
// matching users which may differ from returned results if filter.Limit is specified.
func (s *UserStore) FindUsers(ctx context.Context, filter core.UserFilter) ([]*core.User, int, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, 0, err
	}
	defer tx.Rollback()
	return findUsers(ctx, tx, filter)
}

// CreateUser creates a new user. This is only used for testing since users are
// typically created during the OAuth creation process in AuthService.CreateAuth().
func (s *UserStore) CreateUser(ctx context.Context, user *core.User) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Create a new user object and attach associated Session objects.
	if err := createUser(ctx, tx, user); err != nil {
		return err
	} else if err := attachUserAssociations(ctx, tx, user); err != nil {
		return err
	}
	return tx.Commit()
}

// UpdateUser updates a user object. Returns EUNAUTHORIZED if current user is
// not the user that is being updated. Returns ENOTFOUND if user does not exist.
func (s *UserStore) UpdateUser(ctx context.Context, id string, upd core.UserUpdate) (*core.User, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Update user & attach associated OAuth objects.
	user, err := updateUser(ctx, tx, id, upd)
	if err != nil {
		return user, err
	} else if err := attachUserAssociations(ctx, tx, user); err != nil {
		return user, err
	} else if err := tx.Commit(); err != nil {
		return user, err
	}

	return user, nil
}

// DeleteUser permanently deletes a user and all owned dials.
// Returns EUNAUTHORIZED if current user is not the user being deleted.
// Returns ENOTFOUND if user does not exist.
func (s *UserStore) DeleteUser(ctx context.Context, id string) (*core.User, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	user, err := deleteUser(ctx, tx, id)
	if err != nil {
		return nil, err
	} else if err := tx.Commit(); err != nil {
		return user, err
	}

	return user, nil
}

//
// Database actions
//

// findUserByID is a helper function to fetch a user by ID.
// Returns ENOTFOUND if user does not exist.
func findUserByID(ctx context.Context, tx *Tx, id string) (*core.User, error) {
	a, _, err := findUsers(ctx, tx, core.UserFilter{ID: &id})
	if err != nil {
		return nil, fmt.Errorf("find users: %w", err)
	} else if len(a) == 0 {
		return nil, bookmarkd.ErrNotFound
	}
	return a[0], nil
}

func findUserByUsername(ctx context.Context, tx *Tx, username string) (*core.User, error) {
	a, _, err := findUsers(ctx, tx, core.UserFilter{Username: &username})
	if err != nil {
		return nil, fmt.Errorf("find users: %w", err)
	} else if len(a) == 0 {
		return nil, bookmarkd.ErrNotFound
	}
	return a[0], nil
}

// findUsers returns a list of users matching a filter. Also returns a count of
// total matching users which may differ if filter.Limit is set.
func findUsers(ctx context.Context, tx *Tx, filter core.UserFilter) (_ []*core.User, n int, err error) {
	// Build WHERE clause.
	where, args := []string{"1 = 1"}, []interface{}{}
	if v := filter.ID; v != nil {
		where, args = append(where, "id = ?"), append(args, *v)
	}
	if v := filter.Username; v != nil {
		where, args = append(where, "username = ?"), append(args, *v)
	}

	// Execute query to fetch user rows.
	rows, err := tx.QueryContext(ctx, `
		SELECT 
		    id,
		    username,
		    seed,
		    created_at,
		    updated_at,
		    COUNT(*) OVER()
		FROM users
		WHERE `+strings.Join(where, " AND ")+`
		ORDER BY id ASC
		`+FormatLimitOffset(filter.Limit, filter.Offset),
		args...,
	)
	if err != nil {
		return nil, n, fmt.Errorf("db select users: %w", FormatError(err))
	}
	defer rows.Close()

	// Deserialize rows into User objects.
	users := make([]*core.User, 0)
	for rows.Next() {
		var user core.User
		if err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Seed,
			(*NullTime)(&user.CreatedAt),
			(*NullTime)(&user.UpdatedAt),
			&n,
		); err != nil {
			return nil, 0, fmt.Errorf("db scan users: %w", FormatError(err))
		}

		users = append(users, &user)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("db users rows: %w", FormatError(err))
	}

	return users, n, nil
}

// createUser creates a new user. Sets the new database ID to user.ID and sets
// the timestamps to the current time.
func createUser(ctx context.Context, tx *Tx, user *core.User) error {
	// Set timestamps to the current time.
	user.CreatedAt = tx.Now()
	user.UpdatedAt = user.CreatedAt

	// generate a new userID in uuidv7 format
	userId, err := uuid.NewV7()
	if err != nil {
		return fmt.Errorf("generate a new userID: %w", FormatError(err))
	}

	// we save the uuid as a string here,
	// in the database it is left as a number
	user.ID = userId.String()

	// Validate the user object
	if err := user.Validate(); err != nil {
		return fmt.Errorf("validate user: %w", FormatError(err))
	}

	// Execute insertion query.
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO users (
			id,
			username,
			seed,
			created_at,
			updated_at
		)
		VALUES (?, ?, ?, ?, ?)
	`,
		userId,
		user.Username,
		user.Seed,
		(*NullTime)(&user.CreatedAt),
		(*NullTime)(&user.UpdatedAt),
	); err != nil {
		return fmt.Errorf("db insert user: %w", FormatError(err))
	}

	return nil
}

// updateUser updates fields on a user object. Returns EUNAUTHORIZED if current
// user is not the user being updated.
func updateUser(ctx context.Context, tx *Tx, id string, upd core.UserUpdate) (*core.User, error) {
	// Fetch current object state.
	user, err := findUserByID(ctx, tx, id)
	if err != nil {
		return user, fmt.Errorf("find user by id: %w", err)
	} else if user.ID != core.GetUserIDFromContext(ctx) {
		return nil, bookmarkd.ErrUnauthorized
	}

	// Update fields.
	if v := upd.Username; v != nil {
		user.Username = *v
	}

	// Set last updated date to current time.
	user.UpdatedAt = tx.Now()

	// Execute update query.
	if _, err := tx.ExecContext(ctx, `
		UPDATE users
		SET username = ?,
		    updated_at = ?
		WHERE id = ?
	`,
		user.Username,
		(*NullTime)(&user.UpdatedAt),
		id,
	); err != nil {
		return user, fmt.Errorf("db update user: %w", FormatError(err))
	}

	return user, nil
}

// deleteUser permanently removes a user by ID. Returns EUNAUTHORIZED if current
// user is not the one being deleted.
func deleteUser(ctx context.Context, tx *Tx, id string) (*core.User, error) {
	// Verify object exists.
	user, err := findUserByID(ctx, tx, id)
	if err != nil {
		return nil, fmt.Errorf("find user by id: %w", err)
	} else if user.ID != core.GetUserIDFromContext(ctx) {
		return nil, bookmarkd.ErrUnauthorized
	}

	// Remove row from database.
	if _, err := tx.ExecContext(ctx, `DELETE FROM users WHERE id = ?`, id); err != nil {
		return nil, fmt.Errorf("db delete user: %w", FormatError(err))
	}
	return user, nil
}

//
// Helpers
//

func attachUserAssociations(ctx context.Context, tx *Tx, user *core.User) (err error) {
	if user.Sessions, _, err = findSessions(ctx, tx, core.SessionFilter{UserID: &user.ID}); err != nil {
		return fmt.Errorf("attach user sessions to user: %w", err)
	}
	return nil
}
