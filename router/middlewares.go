package router

import (
	"context"
	"net/http"
	"os"

	"github.com/dgrijalva/jwt-go"
)

// Auth is a middleware for routes
func Auth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, cookieErr := r.Cookie(os.Getenv("COOKIE_NAME"))

		if cookieErr != nil {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		tk := &Token{}

		_, tokenErr := jwt.ParseWithClaims(cookie.Value, tk, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		if tokenErr != nil {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		ctx := context.WithValue(r.Context(), "user", tk)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
