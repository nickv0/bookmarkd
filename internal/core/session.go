package core

import (
	"context"
	"time"
)

type Session struct {
	ID int `json:"id"`

	// User can have one or more methods of session.
	// However, only one per source is allowed per user.
	UserID string `json:"userID"`
	User   *User  `json:"user"`

	// Fields to handle extending a session
	RefreshToken string    `json:"-"`
	ExpiresAt    time.Time `json:"-"`

	// Timestamps of creation and last update.
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAd"`
}

// UserSessionFilter represents a filter accepted by FindSessions().
type SessionFilter struct {
	// Filtering fields.
	ID           *int    `json:"id"`
	UserID       *string `json:"userID"`
	RefreshToken *string `json:"refreshToken"`

	// Restricts results to a subset of the total range.
	// Can be used for pagination.
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

type SessionUpdate struct {
	RefreshToken string    `json:"-"`
	ExpiresAt    time.Time `json:"-"`
}

type SessionStore interface {
	CreateSession(ctx context.Context, session *Session) error
	FindSessionByID(ctx context.Context, id int) (*Session, error)
	FindSessionByUserID(ctx context.Context, id string) (*Session, error)
	FindSessions(ctx context.Context, filter SessionFilter) ([]*Session, int, error)
	RefreshSession(ctx context.Context, token string) (*Session, error)
	DeleteSession(ctx context.Context, id int) error
}
