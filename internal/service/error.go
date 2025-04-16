package service

import (
	"errors"
)

var ErrNotFound = errors.New("not found")
var ErrInternal = errors.New("internal error")
var ErrConflict = errors.New("conflict")
