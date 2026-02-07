package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"user-management/internal/config"
	"user-management/pkg/database"

	"github.com/pressly/goose/v3"
)

func main() {
	var (
		command string
		step    int64
	)

	flag.StringVar(&command, "command", "up", "Migration command (up, down, status, create, reset)")
	flag.Int64Var(&step, "step", 0, "Number of migrations to apply/rollback")
	flag.Parse()

	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Connect to database
	db, err := database.NewDatabase(&cfg.Database)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Error closing database connection: %v", err)
		}
	}()

	sqlDB, err := db.DB.DB()
	if err != nil {
		log.Fatal("Failed to get SQL DB:", err)
	}

	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatal("Failed to set dialect:", err)
	}

	migrationsDir := "internal/migration/migrations"

	switch command {
	case "up":
		if step > 0 {
			current, err := goose.GetDBVersion(sqlDB)
			if err != nil {
				log.Fatal(err)
			}
			target := current + step
			log.Printf("Migrating UP from %d → %d\n", current, target)
			if err := goose.UpTo(sqlDB, migrationsDir, target); err != nil {
				log.Fatal(err)
			}
		} else {
			if err := goose.Up(sqlDB, migrationsDir); err != nil {
				log.Fatal(err)
			}
		}
		log.Println("✅ Migrations applied successfully")

	case "down":
		if step > 0 {
			current, err := goose.GetDBVersion(sqlDB)
			if err != nil {
				log.Fatal(err)
			}
			target := current - step
			if target < 0 {
				target = 0
			}
			log.Printf("Migrating DOWN from %d → %d\n", current, target)
			if err := goose.DownTo(sqlDB, migrationsDir, target); err != nil {
				log.Fatal(err)
			}
		} else {
			if err := goose.Down(sqlDB, migrationsDir); err != nil {
				log.Fatal(err)
			}
		}
		log.Println("✅ Migrations rolled back successfully")

	case "status":
		if err := goose.Status(sqlDB, migrationsDir); err != nil {
			log.Fatal(err)
		}

	case "create":
		name := flag.Arg(0)
		if name == "" {
			log.Fatal("Migration name is required")
		}
		if err := goose.Create(sqlDB, migrationsDir, name, "sql"); err != nil {
			log.Fatal(err)
		}
		log.Printf("✅ Migration file created: %s\n", name)

	case "reset":
		if err := goose.Reset(sqlDB, migrationsDir); err != nil {
			log.Fatal(err)
		}
		if err := goose.Up(sqlDB, migrationsDir); err != nil {
			log.Fatal(err)
		}
		log.Println("✅ Database reset successfully")

	default:
		fmt.Println("Available commands: up, down, status, create, reset")
		os.Exit(1)
	}
}
