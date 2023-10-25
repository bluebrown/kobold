package transport

import "errors"

var (
	ErrAlreadyExists = errors.New("repo already exists")
	ErrNoChange      = errors.New("repo does not contain changes")
)
