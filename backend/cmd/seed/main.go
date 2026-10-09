package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
	"bookflow-go/internal/config"
)

// Optional seeder: applies migrations/0002_seed.sql against DATABASE_URL.
// Usage: go run ./cmd/seed
func main() {
	cfg := config.Load()
	sql, err := os.ReadFile("migrations/0002_seed.sql")
	if err != nil {
		sql, err = os.ReadFile("backend/migrations/0002_seed.sql")
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "seed file not found:", err)
		os.Exit(1)
	}
	conn, err := pgx.Connect(context.Background(), cfg.DatabaseURL)
	if err != nil {
		fmt.Fprintln(os.Stderr, "db connect:", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())
	if _, err := conn.Exec(context.Background(), string(sql)); err != nil {
		fmt.Fprintln(os.Stderr, "seed failed:", err)
		os.Exit(1)
	}
	fmt.Println("seed applied")
}
