package db

import (
	"database/sql"
	"fmt"
)

// DB is a wrapper for sql.DB
type DB struct {
	*sql.DB
}

// New returns a database instance holding a connection to a
// postgres database
func New(
	host string,
	port string,
	user string,
	password string,
	dbname string,
) *DB {
	dbConfig := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := sql.Open("postgres", dbConfig)

	if err != nil {
		panic(err)
	}

	return &DB{db}
}
