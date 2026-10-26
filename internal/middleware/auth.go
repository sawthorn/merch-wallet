package middleware

import (
	"context"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type usernameContextKey struct{}

func WithJWTAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenStr := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		if tokenStr == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		claims := &jwt.RegisteredClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		exp, err := token.Claims.GetExpirationTime()
		if err != nil || time.Now().After(exp.Time) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
		}

		ctx := context.WithValue(r.Context(), usernameContextKey{}, claims.Subject)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
