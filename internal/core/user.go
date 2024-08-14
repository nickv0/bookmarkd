package core

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"bookmarkd"
)

type User struct {
	ID string `json:"id"`

	Username string `json:"username"`
	Seed     string

	// Timestamps for user creation and last update.
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	// List of active sessions
	Sessions []*Session `json:"sessions"`
}

func (u User) Validate() error {
	if err := uuid.Validate(u.ID); err != nil {
		return fmt.Errorf("invalid id generated: %w", bookmarkd.ErrInternal)
	} else if u.Username == "" {
		return fmt.Errorf("%w: username required", bookmarkd.ErrInvalidInput)
	} else if u.Seed == "" {
		return fmt.Errorf("%w: seed required", bookmarkd.ErrInvalidInput)
	}
	return nil
}

// UserFilter represents a filter passed to FindUsers().
type UserFilter struct {
	// Filtering fields.
	ID       *string `json:"id"`
	Username *string `json:"username"`

	// Restrict to subset of results.
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

// UserUpdate represents a set of fields to be updated via UpdateUser().
type UserUpdate struct {
	Username *string `json:"username"`
}

type UserStore interface {
	CreateUser(ctx context.Context, user *User) error
	DeleteUser(ctx context.Context, id string) (*User, error)
	FindUserByID(ctx context.Context, id string) (*User, error)
	FindUserByUsername(ctx context.Context, username string) (*User, error)
	FindUsers(ctx context.Context, filter UserFilter) ([]*User, int, error)
	UpdateUser(ctx context.Context, id string, update UserUpdate) (*User, error)
}
