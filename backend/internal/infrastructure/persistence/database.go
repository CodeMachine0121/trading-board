package persistence

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const slowQueryThreshold = 200 * time.Millisecond

// NewDatabase opens the PostgreSQL connection without touching the schema; "record not found" is an
// expected answer here, so it stays out of the log.
func NewDatabase(dataSourceName string) (*gorm.DB, error) {
	database, openError := gorm.Open(postgres.Open(dataSourceName), &gorm.Config{
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger.Config{
				SlowThreshold:             slowQueryThreshold,
				LogLevel:                  logger.Warn,
				IgnoreRecordNotFoundError: true,
			},
		),
	})
	if openError != nil {
		return nil, fmt.Errorf("open postgres connection: %w", openError)
	}

	return database, nil
}
