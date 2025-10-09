package store

import (
	"errors"
)

// Common store errors
var (
	ErrNotFound = errors.New("record not found")
	ErrConflict = errors.New("record already exists")
)
