package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"
)

// PostStatus is an enum for a post status
type PostStatus string

// Posted and Hidden statuses for a post
const (
	Posted PostStatus = "posted"
	Hidden PostStatus = "hidden"
)

// Post data structure
type Post struct {
	ID         int          `json:"id"`
	UserID     int          `json:"userId" validate:"required,min=1"`
	Title      string       `json:"title" validate:"required"`
	Content    string       `json:"content" validate:"required"`
	Status     PostStatus   `json:"status" validate:"required"`
	CreateTime time.Time    `json:"createTime"`
	UpdateTime JSONNullTime `json:"updateTime"`
}

func (db *DB) GetPostsDB(ctx context.Context, status PostStatus) (*[]Post, error) {
	const postsQuery string = "SELECT * FROM posts WHERE status = $1"
	res, err := db.QueryContext(ctx, postsQuery, status)
	defer res.Close()

	var posts []Post

	if err != nil {
		return &posts, err
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
			return &posts, scanErr
		}
		posts = append(posts, post)
	}

	return &posts, nil
}

func (db *DB) GetPostByIdDB(ctx context.Context, id int) (*Post, error) {
	var post Post
	const postQuery = "SELECT * FROM posts WHERE id = $1;"

	res, err := db.QueryContext(ctx, postQuery, id)
	defer res.Close()

	if err != nil {
		return &post, err
	}

	if !res.Next() {
		return &post, errors.New("No results found")
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
		return &post, scanErr
	}

	return &post, nil
}

func (db *DB) DeletePostByIdDB(ctx context.Context, id int) (int64, error) {
	const deleteQuery = "DELETE FROM posts WHERE id = $1;"

	resp, err := db.ExecContext(ctx, deleteQuery, id)

	if err != nil {
		return 0, err
	}

	return resp.RowsAffected()
}

func (db *DB) StorePostDB(ctx context.Context, post *Post) error {
	const insertQuery string = "INSERT INTO posts (user_id, title, content, status, create_time) " +
		"VALUES ($1, $2, $3, $4, $5)"
	post.CreateTime = time.Now()

	_, err := db.ExecContext(
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

func (db *DB) UpdatePostDB(ctx context.Context, post *Post) (int64, error) {
	const updateQuery string = "UPDATE posts SET title = $1, content = $2, status = $3, update_time = $4 " +
		"WHERE id = $5"
	post.UpdateTime = JSONNullTime{NullTime: sql.NullTime{Time: time.Now(), Valid: true}}

	resp, err := db.ExecContext(
		ctx,
		updateQuery,
		post.Title,
		post.Content,
		post.Status,
		post.UpdateTime,
		post.ID,
	)

	if err != nil {
		return 0, err
	}

	return resp.RowsAffected()
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
