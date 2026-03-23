//go:build mage

package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

// Dev starts the development server with live reload via Air.
func Dev() error {
	return sh.RunV("air")
}

// Build compiles the server binary to bin/cairnpost.
func Build() error {
	mg.Deps(Templ)
	return sh.RunV("go", "build", "-o", "bin/cairnpost", "./cmd/server")
}

// Run builds and runs the server binary.
func Run() error {
	mg.Deps(Build)
	return sh.RunV("./bin/cairnpost")
}

// Test runs all Go tests.
func Test() error {
	return sh.RunV("go", "test", "./...")
}

// Templ generates Go code from .templ files.
func Templ() error {
	return sh.RunV("templ", "generate")
}

// Tailwind builds the Tailwind CSS output file.
func Tailwind() error {
	return sh.RunV("npx", "@tailwindcss/cli",
		"-i", "web/static/css/input.css",
		"-o", "web/static/css/output.css")
}

// TailwindWatch builds Tailwind CSS in watch mode.
func TailwindWatch() error {
	return sh.RunV("npx", "@tailwindcss/cli",
		"-i", "web/static/css/input.css",
		"-o", "web/static/css/output.css",
		"--watch")
}

type DB mg.Namespace

// Up starts PostgreSQL via Docker Compose.
func (DB) Up() error {
	return sh.RunV("docker", "compose", "up", "-d", "db")
}

// Down stops the Docker Compose database.
func (DB) Down() error {
	return sh.RunV("docker", "compose", "down")
}

type Migrate mg.Namespace

// Up runs all pending database migrations.
func (Migrate) Up() error {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	return sh.RunV("migrate",
		"-path", "internal/database/migrations",
		"-database", dbURL,
		"up")
}

// Down rolls back the last database migration.
func (Migrate) Down() error {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	return sh.RunV("migrate",
		"-path", "internal/database/migrations",
		"-database", dbURL,
		"down", "1")
}

type Seed mg.Namespace

// Admin creates or updates an admin user. Set NAME, EMAIL, and PASSWORD env vars.
func (Seed) Admin() error {
	name := os.Getenv("NAME")
	email := os.Getenv("EMAIL")
	password := os.Getenv("PASSWORD")
	if name == "" || email == "" || password == "" {
		return fmt.Errorf("NAME, EMAIL, and PASSWORD env vars are required")
	}
	return sh.RunV("go", "run", "./cmd/seed",
		"--name", name, "--email", email, "--password", password)
}

// Demo seeds a full demo dataset. Set EMAIL and PASSWORD env vars.
func (Seed) Demo() error {
	email := os.Getenv("EMAIL")
	password := os.Getenv("PASSWORD")
	if email == "" || password == "" {
		return fmt.Errorf("EMAIL and PASSWORD env vars are required")
	}
	return sh.RunV("go", "run", "./cmd/seed",
		"--demo", "--email", email, "--password", password)
}

// Clean removes build artifacts.
func Clean() error {
	dirs := []string{"bin", "tmp"}
	for _, d := range dirs {
		if err := sh.Rm(d); err != nil {
			return err
		}
	}
	return nil
}

// Generate runs all code generation (templ + tailwind).
func Generate() error {
	mg.Deps(Templ, Tailwind)
	return nil
}

// Lint runs go vet on all packages.
func Lint() error {
	return sh.RunV("go", "vet", "./...")
}

// Check runs generate, lint, and test in sequence.
func Check() error {
	mg.SerialDeps(Generate, Lint, Test)
	return nil
}

// init registers the default target.
func init() {
	// Ensure binaries from go install are available
	if gopath := os.Getenv("GOPATH"); gopath != "" {
		os.Setenv("PATH", gopath+"/bin:"+os.Getenv("PATH"))
	}
}

// isInstalled checks if a binary is available on PATH.
func isInstalled(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}
