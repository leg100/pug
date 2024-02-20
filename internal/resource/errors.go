package resource

import "errors"

var (
	ErrExists   = errors.New("resource already exists")
	ErrNotFound = errors.New("resource not found")
)
