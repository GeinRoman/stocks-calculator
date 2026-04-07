package app

import "errors"

var (
	ErrUserExists          = errors.New("User already exists")
	ErrWrongCredentials    = errors.New("Wrong credentials provided")
	ErrInvalidRefreshToken = errors.New("Invalid refresh token")
)
