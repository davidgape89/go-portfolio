package main

import (
	"context"
	"database/sql"
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

func (a *App) pingDB() {
	ctx, stop := context.WithCancel(context.Background())
	defer stop()

	pctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	if err := a.db.PingContext(pctx); err != nil {
		panic(err)
	}
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
			fmt.Println(err)
		}
	}

	if err != nil {
		return &User{}, err
	}

	return &user, nil
}
