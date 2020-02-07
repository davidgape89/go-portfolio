package main

import (
	"fmt"
	"go-portfolio/db"
	"go-portfolio/router"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/handlers"
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

	allowedHeaders := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type"})
	allowedOrigins := handlers.AllowedOrigins([]string{os.Getenv("ORIGIN_ALLOWED")})
	allowedMethods := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})

	server := &http.Server{
		Handler:      handlers.CORS(allowedHeaders, allowedOrigins, allowedMethods)(r),
		Addr:         address,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	fmt.Println("Running App on ", address)

	log.Fatal(server.ListenAndServe())
}
