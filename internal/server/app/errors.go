package app

import "errors"

var (
	ErrUserExists          = errors.New("User already exists")
	ErrWrongCredentials    = errors.New("Wrong credentials provided")
	ErrInvalidRefreshToken = errors.New("Invalid refresh token")
	ErrProfileNotFound     = errors.New("Profile not found")
	ErrProfileExists       = errors.New("Profile with that name already exists")
)
