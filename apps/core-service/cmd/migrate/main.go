package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/golang-migrate/migrate/v4"

	// Side-effect imports: register the postgres database driver and the file
	// source with golang-migrate so migrate.New can resolve both schemes.
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"rekberkuy/core-service/config"
)

const defaultMigrationsPath = "db/migrations"

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		printUsage()
		os.Exit(2)
	}

	// Extract the optional -path flag wherever it appears, leaving positional
	// args (and negative numbers like "-1" for steps) untouched.
	migrationsPath := defaultMigrationsPath
	positional := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "-path" || arg == "--path":
			if i+1 >= len(args) {
				log.Fatal("❌ -path requires a value")
			}
			migrationsPath = args[i+1]
			i++
		case strings.HasPrefix(arg, "-path="):
			migrationsPath = strings.TrimPrefix(arg, "-path=")
		case strings.HasPrefix(arg, "--path="):
			migrationsPath = strings.TrimPrefix(arg, "--path=")
		default:
			positional = append(positional, arg)
		}
	}

	if len(positional) == 0 {
		printUsage()
		os.Exit(2)
	}

	command := positional[0]
	if command == "help" || command == "-h" || command == "--help" {
		printUsage()
		return
	}
	rest := positional[1:]

	cfg := config.LoadConfig()

	m, err := migrate.New(fmt.Sprintf("file://%s", migrationsPath), cfg.Database.URL)
	if err != nil {
		log.Fatalf("❌ Failed to initialise migrate: %v", err)
	}
	defer func() {
		srcErr, dbErr := m.Close()
		if srcErr != nil {
			log.Printf("⚠️  migrate source close error: %v", srcErr)
		}
		if dbErr != nil {
			log.Printf("⚠️  migrate database close error: %v", dbErr)
		}
	}()

	switch command {
	case "up":
		if err := m.Up(); err != nil {
			if errors.Is(err, migrate.ErrNoChange) {
				log.Println("✅ Database is already up to date.")
				return
			}
			log.Fatalf("❌ Migration UP failed: %v", err)
		}
		log.Println("✅ All migrations applied successfully (up).")
	case "down":
		if err := m.Down(); err != nil {
			if errors.Is(err, migrate.ErrNoChange) {
				log.Println("ℹ️  No migrations to roll back.")
				return
			}
			log.Fatalf("❌ Migration DOWN failed: %v", err)
		}
		log.Println("✅ All migrations rolled back (database at version 0).")
	case "fresh":
		if err := m.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			log.Fatalf("❌ Migration FRESH (down phase) failed: %v", err)
		}
		if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			log.Fatalf("❌ Migration FRESH (up phase) failed: %v", err)
		}
		log.Println("✅ Database rebuilt from scratch (fresh).")
	case "steps":
		if len(rest) < 1 {
			log.Fatal("❌ 'steps' requires a numeric argument (e.g. 'steps 1' or 'steps -1').")
		}
		n, err := strconv.Atoi(rest[0])
		if err != nil {
			log.Fatalf("❌ Invalid step count %q: %v", rest[0], err)
		}
		if err := m.Steps(n); err != nil {
			if errors.Is(err, migrate.ErrNoChange) {
				log.Println("ℹ️  No migrations to apply for the requested step count.")
				return
			}
			log.Fatalf("❌ Migration STEPS failed: %v", err)
		}
		log.Printf("✅ Applied %d migration step(s).", n)
	case "force":
		if len(rest) < 1 {
			log.Fatal("❌ 'force' requires a version argument (e.g. 'force 1').")
		}
		version, err := strconv.Atoi(rest[0])
		if err != nil || version < 0 {
			log.Fatalf("❌ Invalid version %q: must be a non-negative integer.", rest[0])
		}
		if err := m.Force(version); err != nil {
			log.Fatalf("❌ Migration FORCE failed: %v", err)
		}
		log.Printf("✅ Forced migration version to %d.", version)
	case "version":
		version, dirty, err := m.Version()
		if err != nil {
			if errors.Is(err, migrate.ErrNilVersion) {
				log.Println("ℹ️  No migrations have been applied yet (version 0).")
				return
			}
			log.Fatalf("❌ Failed to read migration version: %v", err)
		}
		state := "clean"
		if dirty {
			state = "DIRTY"
		}
		log.Printf("📍 Current migration version: %d (%s).", version, state)
	default:
		fmt.Fprintf(os.Stderr, "❌ Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(2)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `RekberKuy Core Service — database migration runner (golang-migrate)

Usage:
  go run ./cmd/migrate/main.go [flags] <command> [args]

Flags:
  -path string   path to the SQL migrations directory (default "%s")

Commands:
  up             Apply all pending migrations.
  down           Roll back ALL migrations (drops to version 0 — destructive).
  fresh          Roll back everything, then re-apply all migrations (destructive).
  steps <n>      Apply n migrations forward (n>0) or backward (n<0).
  force <v>      Force the database into version v (recover from a DIRTY state).
  version        Print the current applied migration version + dirty flag.
  help           Show this help.

Examples:
  go run ./cmd/migrate/main.go up
  go run ./cmd/migrate/main.go version
  go run ./cmd/migrate/main.go steps -1
  go run ./cmd/migrate/main.go -path db/migrations force 1
`, defaultMigrationsPath)
}
