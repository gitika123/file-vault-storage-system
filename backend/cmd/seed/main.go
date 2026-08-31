package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/balkanid/file-vault/internal/auth/password"
	"github.com/balkanid/file-vault/internal/platform/config"
	"github.com/balkanid/file-vault/internal/platform/database"
)

type demoUser struct {
	email       string
	displayName string
	role        string
	envName     string
}

var demoUsers = []demoUser{
	{email: "alice@example.com", displayName: "Alice Owner", role: "user", envName: "SEED_ALICE_PASSWORD"},
	{email: "bob@example.com", displayName: "Bob Collaborator", role: "user", envName: "SEED_BOB_PASSWORD"},
	{email: "admin@example.com", displayName: "Admin Reviewer", role: "admin", envName: "SEED_ADMIN_PASSWORD"},
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load configuration: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	db, err := database.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer db.Close()

	for _, user := range demoUsers {
		plainPassword := os.Getenv(user.envName)
		if len(plainPassword) < 12 {
			log.Fatalf("%s must be set and contain at least 12 characters", user.envName)
		}
		hash, err := password.Hash(plainPassword)
		if err != nil {
			log.Fatalf("hash %s password: %v", user.email, err)
		}
		_, err = db.ExecContext(ctx, `
			INSERT INTO users (email, display_name, password_hash, role)
			VALUES ($1, $2, $3, $4::user_role)
			ON CONFLICT (email) DO UPDATE SET
				display_name = EXCLUDED.display_name,
				password_hash = EXCLUDED.password_hash,
				role = EXCLUDED.role,
				disabled_at = NULL,
				updated_at = now()`, user.email, user.displayName, hash, user.role)
		if err != nil {
			log.Fatalf("seed %s: %v", user.email, err)
		}
	}
	fmt.Println("seeded demo users: alice@example.com, bob@example.com, admin@example.com")
}
