package app

import "errors"

var (
	ErrUserExists          = errors.New("User already exists")
	ErrWrongCredentials    = errors.New("Wrong credentials provided")
	ErrInvalidRefreshToken = errors.New("Invalid refresh token")
	ErrProfileNotFound     = errors.New("Profile not found")
	ErrProfileExists       = errors.New("Profile with that name already exists")
	ErrNoDefaultProfile    = errors.New("Default profile not found")
	ErrGroupExists         = errors.New("Group with that name already exists")
	ErrGroupNotFound       = errors.New("Group not found")
	ErrInvalidGroupName    = errors.New("Group name is too long")
	ErrInvalidWeights      = errors.New("Weights are bigger than 100 or lower than 0 or total weight exeeds 100")
)
