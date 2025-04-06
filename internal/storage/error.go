package storage

import "errors"

var ErrConflict = errors.New("data conflict")
var ErrNotFound = errors.New("not found")
