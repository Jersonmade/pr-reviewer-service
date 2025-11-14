package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Jersonmade/pr-reviewer-service/internal/handlers"
	"github.com/Jersonmade/pr-reviewer-service/internal/services"
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

	defer func() {
		if err := store.Close(); err != nil {
			log.Printf("failed to close store: %v", err)
		}
	}()

	log.Println("Successfully connected to PostgreSQL")

	userService := services.NewUserService(store)
	teamService := services.NewTeamService(store)
	prService := services.NewPRService(store, userService)

	userHandler := handlers.NewUserHandler(userService, prService)
	teamHandler := handlers.NewTeamHandler(teamService)
	prHandler := handlers.NewPRHandler(prService)

	mux := http.NewServeMux()

	mux.HandleFunc("/team/add", teamHandler.AddTeam)
	mux.HandleFunc("/team/get", teamHandler.GetTeam)

	mux.HandleFunc("/users/setIsActive", userHandler.SetUserActive)
	mux.HandleFunc("/users/getReview", userHandler.GetUserReviews)

	mux.HandleFunc("/pullRequest/create", prHandler.CreatePR)
	mux.HandleFunc("/pullRequest/merge", prHandler.MergePR)
	mux.HandleFunc("/pullRequest/reassign", prHandler.ReassignReviewer)

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{"status":"ok"}`))
		if err != nil {
			log.Printf("failed to write response: %v", err)
		}
	})

	port := getEnv("PORT", "8080")

	log.Printf("Server starting on :%s...\n", port)

	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Printf("Server failed: %v", err)
		return
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
