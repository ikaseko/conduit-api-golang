package repository

import "errors"

var (
	ErrUserNotFound          = errors.New("user not found")
	ErrUsernameAlreadyExists = errors.New("username already exists")
	ErrEmailAlreadyExists    = errors.New("email already exists")
	ErrTokenIsNotFound       = errors.New("token is not found")
	ErrNoResult              = errors.New("no result")
)
