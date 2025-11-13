package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Jersonmade/pr-reviewer-service/internal/storage"
)

func main() {
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "postgres")
	dbName := getEnv("DB_NAME", "pr_reviewer_service")

	connString := fmt.Sprintf(
        "host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
        dbHost, dbPort, dbUser, dbPassword, dbName,
    )

	store, err := storage.NewPostgresStorage(connString)

	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	defer store.Close()

	log.Println("Successfully connected to PostgreSQL")

	mux := http.NewServeMux()

	port := getEnv("PORT", "8085")

    log.Printf("Server starting on :%s...\n", port)

    if err := http.ListenAndServe(":" + port, mux); err != nil {
        log.Fatal(err)
    }
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}