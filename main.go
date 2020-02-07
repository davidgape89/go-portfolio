package main

import (
	"fmt"
	"go-portfolio/db"
	"go-portfolio/router"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	envErr := godotenv.Load()

	if envErr != nil {
		fmt.Println("Env file not found, running with env variables...")
	}

	address := fmt.Sprintf(
		"%s:%s",
		os.Getenv("ADDRESS"),
		os.Getenv("PORT"),
	)

	database := db.New(
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)
	defer database.Close()

	r := router.New(database)

	fmt.Println("Running App on ", address)

	server := &http.Server{
		Handler:      r,
		Addr:         address,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(server.ListenAndServe())
}
