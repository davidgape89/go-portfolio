package db

import (
	"context"
	"database/sql"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// User type for database
type User struct {
	ID          int          `json:"id"`
	Username    string       `json:"username"`
	Password    string       `json:"password"`
	Email       string       `json:"email"`
	CreatedOn   time.Time    `json:"createdOn"`
	LastLogin   JSONNullTime `json:"lastLogin"`
	Description string       `json:"description"`
}

func (db *DB) CreateUserDB(ctx context.Context, user *User) (sql.Result, error) {
	const insertQuery string = "INSERT INTO users (username, password, email, created_on) " +
		"VALUES ($1, $2, $3, $4)"
	createdOn := time.Now()

	// Hash password
	pass, hashErr := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)

	if hashErr != nil {
		return nil, hashErr
	}

	user.Password = string(pass)

	res, err := db.ExecContext(ctx, insertQuery, user.Username, user.Password, user.Email, createdOn)

	if err != nil {
		return nil, err
	}

	return res, nil
}

func (db *DB) GetUserByEmailDB(ctx context.Context, email string) (*User, error) {
	const userQuery string = "SELECT * FROM users WHERE email = $1;"

	res, err := db.QueryContext(ctx, userQuery, email)
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

func (db *DB) UpdateUserLastLoginByIDDB(id int, time time.Time) error {
	const updateQuery string = `UPDATE users SET last_login=$1 WHERE id=$2`

	_, err := db.Exec(updateQuery, time, id)

	return err
}
