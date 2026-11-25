package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/CodeMachine0121/trading-board/backend/internal/config"
	"github.com/CodeMachine0121/trading-board/backend/internal/infrastructure/persistence"
	"github.com/joho/godotenv"
)

// The migrate command syncs the schema from the entities; the deployment runs it before the server.
func main() {
	if loadError := godotenv.Load(); loadError != nil {
		log.Println("no .env file loaded, falling back to process environment")
	}

	applicationConfig := config.Load()

	database, databaseError := persistence.NewDatabase(applicationConfig.Database.DataSourceName())
	if databaseError != nil {
		log.Fatalf("failed to connect to database: %v", databaseError)
	}

	log.Printf("migrating database %q on %s:%s",
		applicationConfig.Database.Database, applicationConfig.Database.Host, applicationConfig.Database.Port)

	schemaMigrator := persistence.NewSchemaMigrator(database)
	if migrationError := schemaMigrator.Exclusively(context.Background(), func() error {
		migratedTables, migrateError := schemaMigrator.Migrate()
		if migrateError != nil {
			return fmt.Errorf("migration failed: %w", migrateError)
		}

		log.Printf("migration applied to %d table(s): %s", len(migratedTables), strings.Join(migratedTables, ", "))

		return nil
	}); migrationError != nil {
		log.Fatalf("%v", migrationError)
	}
}
