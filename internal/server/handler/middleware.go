package handler

import (
	"context"
	"net/http"
	"strings"
)

func (h *handler) accessTokenValidation(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenStr == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		claims, err := h.tm.ExtractClaims(tokenStr)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userIdContextKey, claims.UserId)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
