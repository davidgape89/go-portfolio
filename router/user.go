package router

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"go-portfolio/db"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/go-playground/validator.v9"
)

// SignUpRequest is the struct for the signup endpoint
type SignUpRequest struct {
	Email           string `json:"email" validate:"required,email"`
	Password        string `json:"password" validate:"required"`
	ConfirmPassword string `json:"confirmPassword" validate:"required,eqfield=Password"`
	Username        string `json:"username" validate:"required"`
}

// SignUpHandler handles the signup endpoint
func (router *Router) SignUpHandler(w http.ResponseWriter, r *http.Request) {
	var payload SignUpRequest
	err := json.NewDecoder(r.Body).Decode(&payload)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	validator := validator.New()
	err = validator.Struct(payload)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	user := db.User{
		Username: payload.Username,
		Password: payload.Password,
		Email:    payload.Email,
	}

	_, dberr := router.db.CreateUserDB(r.Context(), &user)

	if dberr != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(dberr.Error()))
	}
}

// Token holds the data for the token generation
type Token struct {
	ID        int             `json:"id"`
	Username  string          `json:"username"`
	Email     string          `json:"email"`
	CreatedOn time.Time       `json:"createdOn"`
	LastLogin db.JSONNullTime `json:"lastLogin"`
	jwt.StandardClaims
}

// Valid method returns wheather the token is valid or not
func (t *Token) Valid() error {
	if t.VerifyExpiresAt(time.Now().Unix(), true) == false {
		return errors.New("Token expired")
	}

	return nil
}

// LoginRequest holds the payload for a login request
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse is the payload returned by Login handler
type LoginResponse struct {
	ID        int             `json:"id"`
	Username  string          `json:"user"`
	Email     string          `json:"email"`
	CreatedOn time.Time       `json:"createdOn"`
	LastLogin db.JSONNullTime `json:"lastLogin"`
}

// LoginHandler handles login requests
func (router *Router) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var payload LoginRequest
	err := json.NewDecoder(r.Body).Decode(&payload)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	user, dbErr := router.db.GetUserByEmailDB(r.Context(), payload.Email)

	if dbErr != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Incorrect user or password"))
		return
	}

	if !comparePasswordHashes(payload.Password, user.Password) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Incorrect user or password"))
		return
	}

	user.LastLogin = db.JSONNullTime{
		NullTime: sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		},
	}
	updateErr := router.db.UpdateUserLastLoginByIDDB(user.ID, time.Now())

	if updateErr != nil {
		// TODO - Log this properly
		fmt.Println("There was a problem updating the last log in time")
		fmt.Println(updateErr)
	}

	expirationTime := time.Now().Add(time.Minute * 30)

	tokenString, tokenError := generateTokenFromUser(user, expirationTime)

	if tokenError != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("There was a problem generating the token"))
		return
	}

	resp := LoginResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		LastLogin: user.LastLogin,
		CreatedOn: user.CreatedOn,
	}
	cookie := http.Cookie{
		Name:    os.Getenv("COOKIE_NAME"),
		Value:   tokenString,
		Expires: expirationTime,
	}

	http.SetCookie(w, &cookie)
	w.Header().Add("Content-Type", "application/json")
	encErr := json.NewEncoder(w).Encode(resp)

	if encErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(encErr.Error()))
	}
}

// LoginHandler handles login requests
func (router *Router) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie := http.Cookie{
		Name:    os.Getenv("COOKIE_NAME"),
		Value:   "invalidated",
		Expires: time.Now().AddDate(0, 0, -7),
	}

	http.SetCookie(w, &cookie)
	w.WriteHeader(http.StatusForbidden)
	w.Write([]byte("Logged out successfully"))
}

func (router *Router) GetUserHandler(w http.ResponseWriter, r *http.Request) {
	tk := r.Context().Value("user").(*Token)
	resp := LoginResponse{
		ID:        tk.ID,
		Username:  tk.Username,
		Email:     tk.Email,
		CreatedOn: tk.CreatedOn,
		LastLogin: tk.LastLogin,
	}

	w.Header().Add("Content-Type", "application/json")
	encErr := json.NewEncoder(w).Encode(resp)

	if encErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(encErr.Error()))
	}
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
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		CreatedOn: user.CreatedOn,
		LastLogin: user.LastLogin,
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
