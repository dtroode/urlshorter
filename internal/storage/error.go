package storage

import "errors"

// ErrConflict is returned when there is a conflict during URL creation.
var ErrConflict = errors.New("conflict")

// ErrNotFound is returned when a URL is not found.
var ErrNotFound = errors.New("not found")
