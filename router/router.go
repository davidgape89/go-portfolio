package router

import (
	"go-portfolio/db"

	"github.com/gorilla/mux"
)

// Router holds an instance of the router
type Router struct {
	db *db.DB
	mux.Router
}

// New returns an instance of a set up Router
func New(db *db.DB) *Router {
	r := &Router{
		db:     db,
		Router: *mux.NewRouter(),
	}

	// r.HandleFunc("/signup", r.SignUpHandler).Methods("POST")
	r.HandleFunc("/login", r.LoginHandler).Methods("POST")
	r.HandleFunc("/logout", r.LogoutHandler).Methods("POST")
	r.HandleFunc("/posts", r.GetPostsHandler).Methods("GET")
	r.HandleFunc("/posts/{id}", r.GetPostHandler).Methods("GET")
	r.HandleFunc("/posts", Auth(r.NewPostHandler)).Methods("POST")
	r.HandleFunc("/posts/{id}", Auth(r.DeletePostHandler)).Methods("DELETE")
	r.HandleFunc("/posts/{id}", Auth(r.UpdatePostHandler)).Methods("PUT")

	return r
}
