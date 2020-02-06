package router

import (
	"encoding/json"
	"go-portfolio/db"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"gopkg.in/go-playground/validator.v9"
)

func (router *Router) DeletePostHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	rows, dbErr := router.db.DeletePostByIdDB(r.Context(), id)

	if dbErr != nil || rows == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
}

func (router *Router) GetPostHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	post, dbErr := router.db.GetPostByIdDB(r.Context(), id)

	if dbErr != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	if encErr := json.NewEncoder(w).Encode(*post); encErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (router *Router) NewPostHandler(w http.ResponseWriter, r *http.Request) {
	var post db.Post

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

	if dbErr := router.db.StorePostDB(r.Context(), &post); dbErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(dbErr.Error()))
	}
}

func (router *Router) UpdatePostHandler(w http.ResponseWriter, r *http.Request) {
	var post db.Post

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

	rows, dbErr := router.db.UpdatePostDB(r.Context(), &post)

	if dbErr != nil || rows == 0 {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(dbErr.Error()))
	}
}

type PostsResponse struct {
	posts []db.Post `json:"posts"`
}

func (router *Router) GetPostsHandler(w http.ResponseWriter, r *http.Request) {
	posts, err := router.db.GetPostsDB(r.Context(), db.Posted)

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
