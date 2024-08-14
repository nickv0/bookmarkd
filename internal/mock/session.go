package mock

import (
	"context"

	"bookmarkd/internal/core"
)

var _ core.SessionStore = (*SessionStore)(nil)

type SessionStore struct {
	FindSessionByIDFn     func(ctx context.Context, id int) (*core.Session, error)
	FindSessionByUserIDFn func(ctx context.Context, id string) (*core.Session, error)
	FindSessionsFn        func(ctx context.Context, filter core.SessionFilter) ([]*core.Session, int, error)
	CreateSessionFn       func(ctx context.Context, session *core.Session) error
	DeleteSessionFn       func(ctx context.Context, id int) error
	RefreshSessionFn      func(ctx context.Context, refreshToken string) (*core.Session, error)
}

func (s *SessionStore) FindSessionByID(ctx context.Context, id int) (*core.Session, error) {
	return s.FindSessionByIDFn(ctx, id)
}

func (s *SessionStore) FindSessionByUserID(ctx context.Context, id string) (*core.Session, error) {
	return s.FindSessionByUserIDFn(ctx, id)
}

func (s *SessionStore) FindSessions(ctx context.Context, filter core.SessionFilter) ([]*core.Session, int, error) {
	return s.FindSessionsFn(ctx, filter)
}

func (s *SessionStore) CreateSession(ctx context.Context, session *core.Session) error {
	return s.CreateSessionFn(ctx, session)
}

func (s *SessionStore) DeleteSession(ctx context.Context, id int) error {
	return s.DeleteSessionFn(ctx, id)
}

func (s *SessionStore) RefreshSession(ctx context.Context, refreshToken string) (*core.Session, error) {
	return s.RefreshSessionFn(ctx, refreshToken)
}
