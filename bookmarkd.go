package bookmarkd

import (
	"context"
	"errors"
)

// Build version and commit SHA.
var (
	Version string
	Commit  string
)

// ReportError notifies an external service of errors. No-op by default.
var ReportError = func(ctx context.Context, err error, args ...interface{}) {}

// ReportPanic notifies an external service of panics. No-op by default.
var ReportPanic = func(err interface{}) {}

// Application errors.
var (
	// general errors.
	ErrInternal     = errors.New("we encountered an error while processing your request")
	ErrNotFound     = errors.New("the requested resource was not found")
	ErrUnauthorized = errors.New("you are not authenticated to perform the requested action")
	ErrForbidden    = errors.New("you are not authorized to perform the requested action")
	ErrBadRequest   = errors.New("your request is in a bad format")
	ErrInvalidInput = errors.New("there is a problem with the data you submitted")

	// sqlite specific errors.
	ErrUsersUsernameConflict = errors.New("username is already in use")
)
