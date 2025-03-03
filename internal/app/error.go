package app

import (
	"errors"
)

var ErrNotFound = errors.New("not found")
var ErrInternal = errors.New("internal error")
