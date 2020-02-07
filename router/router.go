package router

import (
	"go-portfolio/db"

	"github.com/gorilla/mux"
)

type Router struct {
	db *db.DB
	mux.Router
}

func New(db *db.DB) *Router {
	r := &Router{
		db:     db,
		Router: *mux.NewRouter(),
	}

	r.HandleFunc("/signup", r.SignUpHandler).Methods("POST")
	r.HandleFunc("/login", r.LoginHandler).Methods("POST")
	r.HandleFunc("/logout", r.LogoutHandler).Methods("POST")
	r.HandleFunc("/posts", r.GetPostsHandler).Methods("GET")
	r.HandleFunc("/posts/{id}", r.GetPostHandler).Methods("GET")
	r.HandleFunc("/posts", Auth(r.NewPostHandler)).Methods("POST")
	r.HandleFunc("/posts/{id}", Auth(r.DeletePostHandler)).Methods("DELETE")
	r.HandleFunc("/posts/{id}", Auth(r.UpdatePostHandler)).Methods("PUT")

	return r
}
