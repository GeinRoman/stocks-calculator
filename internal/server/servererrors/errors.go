package servererrors

import "errors"

var (
	ErrUserExists           = errors.New("User already exists")
	ErrWrongCredentials     = errors.New("Wrong credentials provided")
	ErrInvalidRefreshToken  = errors.New("Invalid refresh token")
	ErrProfileNotFound      = errors.New("Profile not found")
	ErrProfileExists        = errors.New("Profile with that name already exists")
	ErrNoDefaultProfile     = errors.New("Default profile not found")
	ErrGroupExists          = errors.New("Group with that name already exists")
	ErrGroupNotFound        = errors.New("Group not found")
	ErrInvalidGroupName     = errors.New("Group name is too long")
	ErrInvalidWeights       = errors.New("Weights are bigger than 100 or lower than 0 or total weight exeeds 100")
	ErrWrongStockGroup      = errors.New("Stock is already in another group")
	ErrWrongStockAmount     = errors.New("Failed to add stocks. Cannot add 0 or less stocks")
	ErrStockNotFound        = errors.New("Stock not found")
	ErrMoexUnhandled        = errors.New("Error in requesting moex api")
	ErrMoexStockNotFound    = errors.New("Failed to find stock on Moscow exchange")
	ErrStocksNotFound       = errors.New("Portfolio doesn't contain stocks")
	ErrGroupsNotFound       = errors.New("Portfolio doesn't contain groups")
	ErrGroupsAreNotWeighted = errors.New("Groups are not weighted, or weights' sum is not 100")
)
