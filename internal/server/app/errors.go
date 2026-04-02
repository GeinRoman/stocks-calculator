package app

import "errors"

var (
	ErrUserExists       = errors.New("user already exists")
	ErrWrongCredentials = errors.New("wrong credentials provided")
)
