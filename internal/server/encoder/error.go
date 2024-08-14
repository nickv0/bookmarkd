package encoder

import (
	"errors"
	"net/http"

	"github.com/go-chi/httplog/v2"

	"bookmarkd"
)

// ErrorResponse is the response that represents an error.
type ErrorResponse struct {
	Error string `json:"error"`
}

// Error prints & optionally logs an error message.
func EncodeError(w http.ResponseWriter, r *http.Request, err error) {
	oplog := httplog.LogEntry(r.Context())

	switch {
	case errors.Is(err, bookmarkd.ErrInternal):
		oplog.Error("internal error", "err", err)
		bookmarkd.ReportError(r.Context(), err, r)
		EncodeJson(w, http.StatusInternalServerError, &ErrorResponse{Error: "Internal error"})
	case errors.Is(err, bookmarkd.ErrNotFound):
		EncodeJson(w, http.StatusNotFound, &ErrorResponse{Error: err.Error()})
	case errors.Is(err, bookmarkd.ErrUnauthorized):
		EncodeJson(w, http.StatusUnauthorized, &ErrorResponse{Error: err.Error()})
	case errors.Is(err, bookmarkd.ErrForbidden):
		EncodeJson(w, http.StatusForbidden, &ErrorResponse{Error: err.Error()})
	case errors.Is(err, bookmarkd.ErrBadRequest):
		EncodeJson(w, http.StatusBadRequest, &ErrorResponse{Error: err.Error()})
	case errors.Is(err, bookmarkd.ErrInvalidInput):
		EncodeJson(w, http.StatusNotAcceptable, &ErrorResponse{Error: err.Error()})
	case errors.Is(err, bookmarkd.ErrUsersUsernameConflict):
		EncodeJson(w, http.StatusNotAcceptable, &ErrorResponse{Error: err.Error()})

	// not one of "our" errors, most likely an unhanded package related error
	default:
		oplog.Error("unhandled error", "err", err)
		bookmarkd.ReportError(r.Context(), err, r)
		EncodeJson(w, http.StatusInternalServerError, &ErrorResponse{Error: "Internal error"})
	}
}
