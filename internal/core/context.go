package core

import (
	"context"
	"time"
)

type SessionContext struct {
	UserID    string
	SessionID int
	ExpiresAt time.Time
}

type contextKey int

const (
	// Stores the current session in the context.
	sessionContextKey = contextKey(iota + 1)
)

func NewContextWithSession(ctx context.Context, session SessionContext) context.Context {
	return context.WithValue(ctx, sessionContextKey, session)
}

func ValidSessionFromContext(ctx context.Context) bool {
	session := SessionFromContext(ctx)
	if session == nil {
		return false
	}
	return time.Now().After(session.ExpiresAt)
}

func SessionFromContext(ctx context.Context) *SessionContext {
	session, _ := ctx.Value(sessionContextKey).(*SessionContext)
	return session
}

func GetUserIDFromContext(ctx context.Context) string {
	s, ok := ctx.Value(sessionContextKey).(SessionContext)
	if !ok {
		return ""
	}
	return s.UserID
}

func GetSessionIDFromContext(ctx context.Context) int {
	s, ok := ctx.Value(sessionContextKey).(SessionContext)
	if !ok {
		return -1
	}
	return s.SessionID
}
