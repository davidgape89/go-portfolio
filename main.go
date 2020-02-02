package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

// App holds the router and the db instances
type App struct {
	db     *sql.DB
	router *mux.Router
}

func main() {
	envErr := godotenv.Load()

	if envErr != nil {
		log.Fatal(envErr)
	}

	a := App{}
	address := ":8080"

	a.db = NewDataBase()
	defer a.db.Close()

	a.SetUpRouter()

	fmt.Println("Running App on ", address)

	server := &http.Server{
		Handler:      a.router,
		Addr:         address,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(server.ListenAndServe())
}
