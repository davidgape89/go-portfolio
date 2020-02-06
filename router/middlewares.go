package router

import (
	"context"
	"go-portfolio/db"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
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

// Helpers

func comparePasswordHashes(password string, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return false
	}
	return true
}

func generateTokenFromUser(user *db.User, expirationTime time.Time) (string, error) {
	token := &Token{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	responseToken := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), token)
	tokenString, error := responseToken.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if error != nil {
		return "", error
	}

	return tokenString, nil
}
