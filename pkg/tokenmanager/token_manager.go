package tokenmanager

import (
	"encoding/hex"
	"errors"
	"math/rand"
	"stocks_calculator/internal/model"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var ErrInvalidAccessToken = errors.New("Invalid access token")

type tokenManager struct {
	key []byte
	ttl time.Duration
}

func New(key []byte, ttl time.Duration) *tokenManager {
	return &tokenManager{key: key, ttl: ttl}
}

func (tm *tokenManager) GenerateJWT(userId int) (model.Token, error) {
	expAt := time.Now().Add(tm.ttl)
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(expAt),
		Subject:   strconv.Itoa(userId),
	})
	str, err := t.SignedString(tm.key)
	if err != nil {
		return model.Token{}, err
	}
	return model.Token{AccessToken: str, ExpiresAt: expAt}, nil
}

func (tm *tokenManager) GenerateRefreshToken() (string, error) {
	b := make([]byte, 32)
	s := rand.NewSource(time.Now().Unix())
	r := rand.New(s)
	_, err := r.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (tm *tokenManager) ExtractClaims(tokenStr string) (model.Claims, error) {
	var claims model.Claims
	token, err := jwt.ParseWithClaims(tokenStr, &jwt.RegisteredClaims{}, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, ErrInvalidAccessToken
		}
		return tm.key, nil
	})
	if err != nil || !token.Valid {
		return claims, ErrInvalidAccessToken
	}
	sub, err := token.Claims.GetSubject()
	if err != nil {
		return claims, ErrInvalidAccessToken
	}
	id, err := strconv.Atoi(sub)
	if err != nil {
		return claims, err
	}
	claims.UserId = id
	return claims, nil
}
