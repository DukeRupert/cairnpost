package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/dukerupert/cairnpost/internal/auth"
	"github.com/dukerupert/cairnpost/internal/database"
	"github.com/dukerupert/cairnpost/internal/model"
	"github.com/dukerupert/cairnpost/internal/repository"
)

func main() {
	name := flag.String("name", "", "User name (required)")
	email := flag.String("email", "", "User email (required)")
	password := flag.String("password", "", "User password (required)")
	orgSlug := flag.String("org", "", "Org slug (defaults to ORG_SLUG env var)")
	flag.Parse()

	if *name == "" || *email == "" || *password == "" {
		fmt.Fprintln(os.Stderr, "Usage: seed --name NAME --email EMAIL --password PASSWORD [--org SLUG]")
		os.Exit(1)
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	slug := *orgSlug
	if slug == "" {
		slug = os.Getenv("ORG_SLUG")
	}
	if slug == "" {
		slug = "demo"
	}

	db, err := database.Connect(dbURL)
	if err != nil {
		log.Fatalf("connecting to database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	orgRepo := repository.NewOrgRepository(db)
	userRepo := repository.NewUserRepository(db)

	// Resolve org
	org, err := orgRepo.GetBySlug(ctx, slug)
	if err != nil {
		log.Fatalf("org %q not found: %v", slug, err)
	}

	// Hash password
	hash, err := auth.HashPassword(*password)
	if err != nil {
		log.Fatalf("hashing password: %v", err)
	}

	// Check if user exists
	existingUser, err := userRepo.GetByEmail(ctx, org.ID, *email)
	if err == nil {
		// User exists — update password
		if err := userRepo.SetPasswordHash(ctx, org.ID, existingUser.ID, hash); err != nil {
			log.Fatalf("updating password: %v", err)
		}
		fmt.Printf("Updated password for %s (%s)\n", existingUser.Name, existingUser.Email)
		return
	}

	// Create new user
	u := model.User{
		OrgID:        org.ID,
		Name:         *name,
		Email:        *email,
		Role:         model.RoleAdmin,
		PasswordHash: &hash,
	}
	if err := userRepo.Create(ctx, &u); err != nil {
		log.Fatalf("creating user: %v", err)
	}
	fmt.Printf("Created admin user %s (%s) in org %s\n", u.Name, u.Email, org.Name)
}
