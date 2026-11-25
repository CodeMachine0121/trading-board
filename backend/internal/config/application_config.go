package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
	// Embedded zone database: the image may carry none, and "Asia/Taipei" would otherwise fail to load.
	_ "time/tzdata"

	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/vo"
	"github.com/shopspring/decimal"
)

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	SslMode  string
}

func (databaseConfig DatabaseConfig) DataSourceName() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		databaseConfig.Host, databaseConfig.Port, databaseConfig.User, databaseConfig.Password,
		databaseConfig.Database, databaseConfig.SslMode)
}

// ApplicationConfig is every setting, read once from the environment at start-up.
type ApplicationConfig struct {
	ServerPort            string
	CorsAllowedOrigins    []string
	Database              DatabaseConfig
	AccessCredential      vo.AccessCredentialVo
	AccessTokenSigningKey string
	AccessTokenLifetime   time.Duration
	FeeSchedule           vo.FeeScheduleVo
	BudgetPolicy          vo.BudgetPolicyVo
	BusinessTimeZone      *time.Location
}

// Load falls back to each setting's default when the variable is unset or unreadable, and says so in the log.
func Load() ApplicationConfig {
	return ApplicationConfig{
		ServerPort:         stringWithDefault("SERVER_PORT", "8080"),
		CorsAllowedOrigins: listWithDefault("CORS_ALLOWED_ORIGINS", []string{"http://localhost:3000"}),
		Database: DatabaseConfig{
			Host:     stringWithDefault("POSTGRES_HOST", "localhost"),
			Port:     stringWithDefault("POSTGRES_PORT", "5432"),
			User:     stringWithDefault("POSTGRES_USER", "postgres"),
			Password: os.Getenv("POSTGRES_PASSWORD"),
			Database: stringWithDefault("POSTGRES_DATABASE", "trading_board"),
			SslMode:  stringWithDefault("POSTGRES_SSL_MODE", "disable"),
		},
		AccessCredential: vo.AccessCredentialVo{
			OwnerUsername: os.Getenv("OWNER_USERNAME"),
			OwnerPassword: os.Getenv("OWNER_PASSWORD"),
			AgentApiKey:   os.Getenv("AGENT_API_KEY"),
		},
		AccessTokenSigningKey: os.Getenv("AUTH_ACCESS_TOKEN_SIGNING_KEY"),
		AccessTokenLifetime:   time.Duration(positiveIntWithDefault("AUTH_ACCESS_TOKEN_LIFETIME_HOURS", 24)) * time.Hour,
		FeeSchedule: vo.FeeScheduleVo{
			BrokerFeeRate:      decimalWithDefault("FEE_BROKER_RATE", "0.001425"),
			Discount:           decimalWithDefault("FEE_DISCOUNT", "0.6"),
			MinimumFee:         decimalWithDefault("FEE_MINIMUM", "1"),
			TransactionTaxRate: decimalWithDefault("FEE_TRANSACTION_TAX_RATE", "0.003"),
		},
		BudgetPolicy: vo.BudgetPolicyVo{
			MonthlyAllocation: decimalWithDefault("BUDGET_MONTHLY_ALLOCATION", "10000"),
			MonthEndStartDay:  positiveIntWithDefault("BUDGET_MONTH_END_START_DAY", 22),
		},
		BusinessTimeZone: locationWithDefault("BUSINESS_TIME_ZONE", "Asia/Taipei"),
	}
}

func stringWithDefault(name string, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}

	return fallback
}

func listWithDefault(name string, fallback []string) []string {
	items := []string{}
	for _, item := range strings.Split(os.Getenv(name), ",") {
		if trimmed := strings.TrimSpace(item); trimmed != "" {
			items = append(items, trimmed)
		}
	}
	if len(items) == 0 {
		return fallback
	}

	return items
}

func positiveIntWithDefault(name string, fallback int) int {
	value, parseError := strconv.Atoi(stringWithDefault(name, strconv.Itoa(fallback)))
	if parseError != nil || value <= 0 {
		log.Printf("%s is not a positive whole number; using %d", name, fallback)
		return fallback
	}

	return value
}

func decimalWithDefault(name string, fallback string) decimal.Decimal {
	value, parseError := decimal.NewFromString(stringWithDefault(name, fallback))
	if parseError != nil {
		log.Printf("%s is not a number; using %s", name, fallback)
		return decimal.RequireFromString(fallback)
	}

	return value
}

func locationWithDefault(name string, fallback string) *time.Location {
	location, loadError := time.LoadLocation(stringWithDefault(name, fallback))
	if loadError != nil {
		log.Printf("%s is not a known time zone; using %s", name, fallback)
		location, _ = time.LoadLocation(fallback)
	}

	return location
}
