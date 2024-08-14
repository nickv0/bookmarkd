package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"bookmarkd"
	"bookmarkd/internal/core"
)

// Ensure service implements interface.
var _ core.SessionStore = (*SessionStore)(nil)

// SessionStore represents a service for managing user sessions.
type SessionStore struct {
	db *DB
}

// NewSessionStore returns a new instance of SessionStore.
func NewSessionStore(db *DB) *SessionStore {
	return &SessionStore{db: db}
}

func (s *SessionStore) FindSessionByID(ctx context.Context, id int) (*core.Session, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	session, err := findSessionByID(ctx, tx, id)
	if err != nil {
		return nil, err
	} else if err := attachSessionAssociations(ctx, tx, session); err != nil {
		return session, err
	}
	return session, nil
}

func (s *SessionStore) FindSessionByUserID(ctx context.Context, id string) (*core.Session, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	session, err := findSessionByUserID(ctx, tx, id)
	if err != nil {
		return nil, err
	} else if err := attachSessionAssociations(ctx, tx, session); err != nil {
		return session, err
	}
	return session, nil
}

func (s *SessionStore) FindSessions(ctx context.Context, filter core.SessionFilter) ([]*core.Session, int, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, 0, err
	}
	defer tx.Rollback()
	return findSessions(ctx, tx, filter)
}

func (s *SessionStore) CreateSession(ctx context.Context, session *core.Session) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Create a new user object and attach associated Session objects.
	if err := createSession(ctx, tx, session); err != nil {
		return err
	} else if err := attachSessionAssociations(ctx, tx, session); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *SessionStore) RefreshSession(ctx context.Context, refreshToken string) (*core.Session, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Update user & attach associated OAuth objects.
	session, err := refreshSession(ctx, tx, refreshToken)
	if err != nil {
		return session, err
	} else if err := attachSessionAssociations(ctx, tx, session); err != nil {
		return session, err
	} else if err := tx.Commit(); err != nil {
		return session, err
	}

	return session, nil
}

func (s *SessionStore) DeleteSession(ctx context.Context, id int) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := deleteSession(ctx, tx, id); err != nil {
		return err
	} else if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

//
// helper functions
//

func findSessionByID(ctx context.Context, tx *Tx, id int) (*core.Session, error) {
	sessions, _, err := findSessions(ctx, tx, core.SessionFilter{ID: &id})
	if err != nil {
		return nil, fmt.Errorf("find sessions: %w", err)
	} else if len(sessions) == 0 {
		return nil, bookmarkd.ErrNotFound
	}
	return sessions[0], nil
}

func findSessionByUserID(ctx context.Context, tx *Tx, id string) (*core.Session, error) {
	sessions, _, err := findSessions(ctx, tx, core.SessionFilter{UserID: &id})
	if err != nil {
		return nil, fmt.Errorf("find sessions: %w", err)
	} else if len(sessions) == 0 {
		return nil, bookmarkd.ErrNotFound
	}
	return sessions[0], nil
}

func findSessionByRefreshToken(ctx context.Context, tx *Tx, token string) (*core.Session, error) {
	sessions, _, err := findSessions(ctx, tx, core.SessionFilter{RefreshToken: &token})
	if err != nil {
		return nil, fmt.Errorf("find sessions: %w", err)
	} else if len(sessions) == 0 {
		return nil, bookmarkd.ErrNotFound
	}
	return sessions[0], nil
}

func findSessions(ctx context.Context, tx *Tx, filter core.SessionFilter) (_ []*core.Session, n int, err error) {
	// Build WHERE clause. Each part of the clause is AND-ed together to further
	// restrict the results. Placeholders are added to "args" and are used
	// to avoid SQL injection.
	//
	// Each filter field is optional.
	where, args := []string{"1 = 1"}, []interface{}{}
	if v := filter.ID; v != nil {
		where, args = append(where, "id = ?"), append(args, *v)
	}
	if v := filter.UserID; v != nil {
		where, args = append(where, "user_id = ?"), append(args, *v)
	}

	// Execute the query with WHERE clause and LIMIT/OFFSET injected.
	rows, err := tx.QueryContext(ctx, `
		SELECT 
		    id,
		    user_id,
		    refresh_token,
		    expires_at,
		    created_at,
		    updated_at,
		    COUNT(*) OVER()
		FROM sessions
		WHERE `+strings.Join(where, " AND ")+`
		ORDER BY id ASC
		`+FormatLimitOffset(filter.Limit, filter.Offset)+`
	`,
		args...,
	)
	if err != nil {
		return nil, n, fmt.Errorf("db select user session: %w", FormatError(err))
	}
	defer rows.Close()

	// Iterate over result set and deserialize rows into Session objects.
	sessions := make([]*core.Session, 0)
	for rows.Next() {
		var session core.Session
		var expiry sql.NullString
		if err := rows.Scan(
			&session.ID,
			&session.UserID,
			&session.RefreshToken,
			&expiry,
			(*NullTime)(&session.CreatedAt),
			(*NullTime)(&session.UpdatedAt),
			&n,
		); err != nil {
			return nil, 0, fmt.Errorf("db scan user session row: %w", FormatError(err))
		}

		if expiry.Valid {
			if v, _ := time.Parse(time.RFC3339, expiry.String); !v.IsZero() {
				session.ExpiresAt = v
			}
		}

		sessions = append(sessions, &session)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("db scan user session rows: %w", FormatError(err))
	}

	return sessions, n, nil
}

func createSession(ctx context.Context, tx *Tx, session *core.Session) error {
	if session.UserID == "" {
		return fmt.Errorf("unable to determine userID: %w", bookmarkd.ErrInternal)
	}
	// Set timestamp fields to current time.
	session.CreatedAt = tx.Now()
	session.UpdatedAt = session.CreatedAt

	// generate a random refresh Token
	refreshToken, err := uuid.NewRandom()
	if err != nil {
		return fmt.Errorf("create refresh token: %w", err)
	}
	session.RefreshToken = refreshToken.String()

	// set the refresh expiration
	session.ExpiresAt = session.CreatedAt.Add(time.Duration(24*14) * time.Hour)

	// Execute insertion query.
	result, err := tx.ExecContext(ctx, `
		INSERT INTO sessions (
			user_id,
			refresh_token,
			expires_at,
			created_at,
			updated_at
		)
		VALUES (?, ?, ?, ?, ?)
	`,
		session.UserID,
		session.RefreshToken,
		(*NullTime)(&session.ExpiresAt),
		(*NullTime)(&session.CreatedAt),
		(*NullTime)(&session.UpdatedAt),
	)
	if err != nil {
		return fmt.Errorf("db create user session: %w", FormatError(err))
	}

	// Update caller object to set ID.
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("db create user session id: %w", FormatError(err))
	}
	session.ID = int(id)

	return nil
}

func refreshSession(ctx context.Context, tx *Tx, refreshToken string) (*core.Session, error) {

	s, err := findSessionByRefreshToken(ctx, tx, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("find session by refresh token: %w", err)
	}

	rt, err := uuid.NewRandom()
	if err != nil {
		return nil, fmt.Errorf("new refresh token: %w", err)
	}

	s.ExpiresAt = tx.Now().Add(time.Duration(24*14) * time.Hour)
	s.RefreshToken = rt.String()
	s.UpdatedAt = tx.Now()

	// Execute SQL update query.
	if _, err := tx.ExecContext(ctx, `
		UPDATE sessions
		SET refresh_token = ?,
		    expires_at = ?,
		    updated_at = ?
		WHERE refresh_token = ?
	`,
		s.RefreshToken,
		(*NullTime)(&s.ExpiresAt),
		(*NullTime)(&s.UpdatedAt),
		refreshToken,
	); err != nil {
		return s, fmt.Errorf("db update session: %w", FormatError(err))
	}

	return s, nil
}

func deleteSession(ctx context.Context, tx *Tx, id int) error {
	// Verify object exists & that the user is the owner of the session.
	if session, err := findSessionByID(ctx, tx, id); err != nil {
		return fmt.Errorf("find user session by id: %w", err)
	} else if session.UserID != core.GetUserIDFromContext(ctx) {
		return bookmarkd.ErrUnauthorized
	}

	// Remove row from database.
	if _, err := tx.ExecContext(ctx, `DELETE FROM sessions WHERE id = ?`, id); err != nil {
		return fmt.Errorf("db delete user session: %w", FormatError(err))
	}

	return nil
}

//
// Helpers
//

func attachSessionAssociations(ctx context.Context, tx *Tx, session *core.Session) (err error) {
	if session.User, err = findUserByID(ctx, tx, session.UserID); err != nil {
		return fmt.Errorf("attach user to session: %w", err)
	}

	return nil
}
