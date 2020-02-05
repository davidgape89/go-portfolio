package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/go-playground/validator.v9"
)

// SetUpRouter instantiates the router and all the routes
func (a *App) SetUpRouter() {
	r := mux.NewRouter()

	r.HandleFunc("/signup", a.SignUpHandler).Methods("POST")
	r.HandleFunc("/login", a.LoginHandler).Methods("POST")
	r.HandleFunc("/posts", a.GetPostsHandler).Methods("GET")
	r.HandleFunc("/posts", a.NewPostHandler).Methods("POST")
	r.HandleFunc("/posts/{id}", a.GetPostHandler).Methods("GET")
	r.HandleFunc("/posts/{id}", a.DeletePostHandler).Methods("DELETE")

	a.router = r
}

func (a *App) DeletePostHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	rows, dbErr := a.deletePostByIdDB(r.Context(), id)

	if dbErr != nil || rows == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
}

func (a *App) GetPostHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	post, dbErr := a.getPostByIdDB(r.Context(), id)

	if dbErr != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	if encErr := json.NewEncoder(w).Encode(post); encErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (a *App) NewPostHandler(w http.ResponseWriter, r *http.Request) {
	var post Post

	if encErr := json.NewDecoder(r.Body).Decode(&post); encErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(encErr.Error()))
		return
	}

	validator := validator.New()

	if valErr := validator.Struct(post); valErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(valErr.Error()))
		return
	}

	if dbErr := a.storePostDB(r.Context(), post); dbErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(dbErr.Error()))
	}
}

type PostsResponse struct {
	posts []Post `json:"posts"`
}

func (a *App) GetPostsHandler(w http.ResponseWriter, r *http.Request) {
	posts, err := a.getPostsDB(r.Context(), Posted)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	encErr := json.NewEncoder(w).Encode(posts)

	if encErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(encErr.Error()))
		return
	}
}

// SignUpRequest is the struct for the signup endpoint
type SignUpRequest struct {
	Email           string `json:"email" validate:"required,email"`
	Password        string `json:"password" validate:"required"`
	ConfirmPassword string `json:"confirmPassword" validate:"required,eqfield=Password"`
	Username        string `json:"username" validate:"required"`
}

// SignUpHandler handles the signup endpoint
func (a *App) SignUpHandler(w http.ResponseWriter, r *http.Request) {
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

// Token holds the data for the token generation
type Token struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
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
		return
	}

	user, dbErr := a.getUserByEmailDB(r.Context(), payload.Email)

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
		fmt.Println(encErr)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(encErr.Error()))
	}

	updateErr := a.updateUserLastLoginByIDDB(user.ID, time.Now())

	if updateErr != nil {
		fmt.Println("There was a problem updating the last log in time")
		fmt.Println(updateErr)
	}
}

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
			fmt.Println(tokenErr)
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

func generateTokenFromUser(user *User, expirationTime time.Time) (string, error) {
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
		fmt.Println(error)
		return "", error
	}

	return tokenString, nil
}
