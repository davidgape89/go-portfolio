package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"os"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// NewDataBase creates a new database instance
func NewDataBase() *sql.DB {
	var (
		host     = os.Getenv("DB_HOST")
		port     = os.Getenv("DB_PORT")
		user     = os.Getenv("DB_USER")
		password = os.Getenv("DB_PASSWORD")
		dbname   = os.Getenv("DB_NAME")
	)

	dbConfig := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := sql.Open("postgres", dbConfig)

	if err != nil {
		panic(err)
	}

	return db
}

// User type for database
type User struct {
	ID          int       `json:"id"`
	Username    string    `json:"username"`
	Password    string    `json:"password"`
	Email       string    `json:"email"`
	CreatedOn   time.Time `json:"createdOn"`
	LastLogin   time.Time `json:"lastLogin"`
	Description string    `json:"description"`
}

func (a *App) createUserDB(ctx context.Context, user *User) (sql.Result, error) {
	const insertQuery string = "INSERT INTO users (username, password, email, created_on) " +
		"VALUES ($1, $2, $3, $4)"
	createdOn := time.Now()

	// Hash password
	pass, hashErr := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)

	if hashErr != nil {
		return nil, hashErr
	}

	user.Password = string(pass)

	res, err := a.db.ExecContext(ctx, insertQuery, user.Username, user.Password, user.Email, createdOn)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return res, nil
}

func (a *App) getUserByEmailDB(ctx context.Context, email string) (*User, error) {
	const userQuery string = "SELECT * FROM users WHERE email = $1;"
	fmt.Println(email)
	res, err := a.db.QueryContext(ctx, userQuery, email)
	defer res.Close()
	if err != nil {
		return &User{}, err
	}

	var user User

	for res.Next() {
		scanErr := res.Scan(
			&user.ID,
			&user.Username,
			&user.Password,
			&user.Email,
			&user.CreatedOn,
			&user.LastLogin,
		)
		if scanErr != nil {
			return &User{}, scanErr
		}
	}

	return &user, nil
}

func (a *App) updateUserLastLoginByIDDB(id int, time time.Time) error {
	const updateQuery string = `UPDATE users SET last_login=$1 WHERE id=$2`

	_, err := a.db.Exec(updateQuery, time, id)

	return err
}

// POSTS

func (a *App) getPostsDB(ctx context.Context, status PostStatus) ([]Post, error) {
	const postsQuery string = "SELECT * FROM posts WHERE status = $1"
	res, err := a.db.QueryContext(ctx, postsQuery, status)
	defer res.Close()

	var posts []Post

	if err != nil {
		return posts, err
	}

	for res.Next() {
		var post Post
		scanErr := res.Scan(
			&post.ID,
			&post.UserID,
			&post.Title,
			&post.Content,
			&post.Status,
			&post.CreateTime,
			&post.UpdateTime,
		)
		if scanErr != nil {
			return posts, scanErr
		}
		posts = append(posts, post)
	}

	return posts, nil
}

func (a *App) getPostByIdDB(ctx context.Context, id int) (Post, error) {
	var post Post
	const postQuery = "SELECT * FROM posts WHERE id = $1;"

	res, err := a.db.QueryContext(ctx, postQuery, id)
	defer res.Close()

	if err != nil {
		return post, err
	}

	if !res.Next() {
		return post, errors.New("No results found")
	}

	scanErr := res.Scan(
		&post.ID,
		&post.UserID,
		&post.Title,
		&post.Content,
		&post.Status,
		&post.CreateTime,
		&post.UpdateTime,
	)

	if scanErr != nil {
		return post, scanErr
	}

	return post, nil
}

func (a *App) deletePostByIdDB(ctx context.Context, id int) (int64, error) {
	const deleteQuery = "DELETE FROM posts WHERE id = $1;"

	resp, err := a.db.ExecContext(ctx, deleteQuery, id)

	if err != nil {
		return 0, err
	}

	return resp.RowsAffected()
}

func (a *App) storePostDB(ctx context.Context, post Post) error {
	fmt.Println(post)
	const insertQuery string = "INSERT INTO posts (user_id, title, content, status, create_time) " +
		"VALUES ($1, $2, $3, $4, $5)"
	post.CreateTime = time.Now()

	_, err := a.db.ExecContext(
		ctx,
		insertQuery,
		post.UserID,
		post.Title,
		post.Content,
		post.Status,
		post.CreateTime,
	)

	if err != nil {
		return err
	}

	return nil
}

// JSONNullTime stores a time value that could be null
type JSONNullTime struct {
	sql.NullTime
}

// MarshalJSON method encodes the database value into a json value
func (v JSONNullTime) MarshalJSON() ([]byte, error) {
	if v.Valid {
		return json.Marshal(v.Time)
	}

	return json.Marshal(nil)
}

// UnmarshalJSON encodes a json value into a database value
func (v *JSONNullTime) UnmarshalJSON(data []byte) error {
	// Unmarshalling into a pointer will let us detect null
	var t *time.Time
	if err := json.Unmarshal(data, &t); err != nil {
		return err
	}
	if t != nil {
		v.Valid = true
		v.Time = *t
	} else {
		v.Valid = false
	}
	return nil
}

type PostStatus string

const (
	Posted PostStatus = "posted"
	Hidden PostStatus = "hidden"
)

type Post struct {
	ID         int          `json:"id"`
	UserID     int          `json:"userId" validate:"required,min=1"`
	Title      string       `json:"title" validate:"required"`
	Content    string       `json:"content" validate:"required"`
	Status     PostStatus   `json:"status" validate:"required"`
	CreateTime time.Time    `json:"createTime"`
	UpdateTime JSONNullTime `json:"updateTime"`
}

func (a *App) updateUserLastLoginByIDDB(id int, time time.Time) error {
	const updateQuery string = `UPDATE users SET last_login=$1 WHERE id=$2`

	_, err := a.db.Exec(updateQuery, time, id)

	return err
}
