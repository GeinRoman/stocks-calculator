package servererrors

import "errors"

var (
	ErrInvalidCrypto       = errors.New("Invalid crypto asset or transaction")
	ErrCryptoNotFound      = errors.New("Crypto asset not found")
	ErrCryptoExists        = errors.New("Crypto asset name already exists in this profile")
	ErrCryptoGroupNotEmpty = errors.New("Crypto group is not empty")
)
