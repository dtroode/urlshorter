package service

import (
	"errors"
)

// ErrNotFound is returned when a requested resource is not found.
// This error typically indicates a 404 Not Found HTTP status.
var ErrNotFound = errors.New("not found")

// ErrInternal is returned when an internal server error occurs.
// This error typically indicates a 500 Internal Server Error HTTP status.
var ErrInternal = errors.New("internal error")

// ErrConflict is returned when there is a conflict with the current state of the resource.
// This error typically indicates a 409 Conflict HTTP status.
var ErrConflict = errors.New("conflict")

// ErrNoContent is returned when a request is successful but there is no content to return.
// This error typically indicates a 204 No Content HTTP status.
var ErrNoContent = errors.New("no content")

// ErrGone is returned when a requested resource has been permanently deleted.
// This error typically indicates a 410 Gone HTTP status.
var ErrGone = errors.New("gone")
