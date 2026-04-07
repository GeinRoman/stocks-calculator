package tokenmanager

import (
	"encoding/hex"
	"math/rand"
	"stocks_calculator/internal/model"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type tokenManager struct {
	key []byte
	ttl time.Duration
}

func New(key []byte, ttl time.Duration) *tokenManager {
	return &tokenManager{key: key, ttl: ttl}
}

func (g *tokenManager) GenerateJWT(userId int) (model.Token, error) {
	expAt := time.Now().Add(g.ttl)
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(expAt),
		Subject:   strconv.Itoa(userId),
	})
	str, err := t.SignedString(g.key)
	if err != nil {
		return model.Token{}, err
	}
	return model.Token{AccessToken: str, ExpiresAt: expAt}, nil
}

func (g *tokenManager) GenerateRefreshToken() (string, error) {
	b := make([]byte, 32)
	s := rand.NewSource(time.Now().Unix())
	r := rand.New(s)
	_, err := r.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
