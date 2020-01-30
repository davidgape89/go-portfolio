package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/go-playground/validator.v9"
)

// SetUpRouter instantiates the router and all the routes
func (a *App) SetUpRouter() {
	r := mux.NewRouter()

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello"))
	})

	r.HandleFunc("/signup", a.SignUpHandler).Methods("POST")
	r.HandleFunc("/login", a.LoginHandler).Methods("POST")

	a.router = r
}

// SignUpPayload is the struct for the signup endpoint
type SignUpPayload struct {
	Email           string `json:"email" validate:"required,email"`
	Password        string `json:"password" validate:"required"`
	ConfirmPassword string `json:"confirmPassword" validate:"required,eqfield=Password"`
	Username        string `json:"username" validate:"required"`
}

// SignUpHandler handles the signup endpoint
func (a *App) SignUpHandler(w http.ResponseWriter, r *http.Request) {
	var payload SignUpPayload
	err := json.NewDecoder(r.Body).Decode(&payload)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	validator := validator.New()
	err = validator.Struct(payload)
	fmt.Println(payload)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	user := User{
		Username: payload.Username,
		Password: payload.Password,
		Email:    payload.Email,
	}

	_, dberr := a.createUserDB(r.Context(), &user)

	if dberr != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(dberr.Error()))
	}
}

// LoginRequest holds the payload for a login request
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// Token holds the data for the token generation
type Token struct {
	ID             int    `json:"id"`
	Username       string `json:"username"`
	Email          string `json:"email"`
	StandardClaims *jwt.StandardClaims
}

func (t *Token) Valid() error {
	return nil
}

type LoginResponse struct {
	ID        int       `json:"id"`
	Username  string    `json:"user"`
	Email     string    `json:"email"`
	CreatedOn time.Time `json:"createdOn"`
	LastLogin time.Time `json:"lastLogin"`
}

// LoginHandler handles login requests
func (a *App) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var payload LoginRequest
	err := json.NewDecoder(r.Body).Decode(&payload)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
	}

	user, dbErr := a.getUserByEmailDB(r.Context(), payload.Email)

	if dbErr != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Incorrect user or password"))
	}

	if !comparePasswordHashes(payload.Password, user.Password) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Incorrect user or password"))
	}

	expirationTime := time.Now().Add(time.Minute * 30)
	token := &Token{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		StandardClaims: &jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	responseToken := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), token)
	tokenString, error := responseToken.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if error != nil {
		fmt.Println(error)
	}

	resp := LoginResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		LastLogin: user.LastLogin,
		CreatedOn: user.CreatedOn,
	}
	cookie := http.Cookie{
		Name:    "dport",
		Value:   tokenString,
		Expires: expirationTime,
	}

	http.SetCookie(w, &cookie)
	w.Header().Add("Content-Type", "application/json")
	encErr := json.NewEncoder(w).Encode(resp)

	if encErr != nil {
		fmt.Println(encErr)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(encErr.Error()))
	}
}

func comparePasswordHashes(password string, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return false
	}
	return true
}
