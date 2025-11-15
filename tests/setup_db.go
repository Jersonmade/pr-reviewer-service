package tests

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Jersonmade/pr-reviewer-service/internal/handlers"
	"github.com/Jersonmade/pr-reviewer-service/internal/services"
	"github.com/Jersonmade/pr-reviewer-service/internal/storage"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type TestEnvironment struct {
	Store       *storage.PostgresStorage
	TeamHandler *handlers.TeamHandler
	PRHandler   *handlers.PRHandler
	UserHandler *handlers.UserHandler
	Cleanup     func()
}

func SetupTestEnvironment(t *testing.T) *TestEnvironment {
	t.Helper()

	store, cleanup := setupTestDB(t)

	teamService := services.NewTeamService(store)
	userService := services.NewUserService(store)
	prService := services.NewPRService(store, userService)

	teamHandler := handlers.NewTeamHandler(teamService)
	prHandler := handlers.NewPRHandler(prService)
	userHandler := handlers.NewUserHandler(userService, prService)

	return &TestEnvironment{
		Store:       store,
		TeamHandler: teamHandler,
		PRHandler:   prHandler,
		UserHandler: userHandler,
		Cleanup:     cleanup,
	}
}

func setupTestDB(t *testing.T) (*storage.PostgresStorage, func()) {
	ctx := context.Background()

	wd, err := os.Getwd()

	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	initScriptPath := filepath.Join(wd, "..", "migrations", "001_init.sql")

	postgresContainer, err := postgres.Run(ctx,
		"postgres:15-alpine",
		postgres.WithDatabase("pr_reviewer_service"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		postgres.WithInitScripts(initScriptPath),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second)),
	)

	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}

	host, _ := postgresContainer.Host(ctx)
	port, _ := postgresContainer.MappedPort(ctx, "5432")
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port.Port(), "postgres", "postgres", "pr_reviewer_service")

	store, err := storage.NewPostgresStorage(connStr)

	if err != nil {
		t.Fatalf("failed to connect to test db: %v", err)
	}

	cleanup := func() {
		_ = store.Close()
		_ = postgresContainer.Terminate(ctx)
	}

	return store, cleanup
}
