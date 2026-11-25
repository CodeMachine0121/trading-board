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

// RefreshCookieConfig says how the refresh-token cookie is scoped: Domain empty means the API host only.
type RefreshCookieConfig struct {
	Domain   string
	IsSecure bool
}

// ApplicationConfig is every setting, read once from the environment at start-up.
type ApplicationConfig struct {
	ServerPort            string
	CorsAllowedOrigins    []string
	Database              DatabaseConfig
	OwnerAccount          vo.OwnerAccountVo
	OAuthClient           vo.OAuthClientVo
	AuthorizationPolicy   vo.AuthorizationPolicyVo
	RefreshCookie         RefreshCookieConfig
	AccessTokenSigningKey string
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
		OwnerAccount: vo.OwnerAccountVo{
			Email:        os.Getenv("OWNER_EMAIL"),
			PasswordHash: os.Getenv("OWNER_PASSWORD_HASH"),
		},
		OAuthClient: vo.OAuthClientVo{
			ClientID:     stringWithDefault("OAUTH_CLIENT_ID", "trading-board-web"),
			RedirectUris: listWithDefault("OAUTH_REDIRECT_URIS", []string{"http://localhost:3000/auth/callback"}),
		},
		AuthorizationPolicy: vo.AuthorizationPolicyVo{
			AuthorizationCodeLifetime: time.Duration(
				positiveIntWithDefault("AUTH_AUTHORIZATION_CODE_LIFETIME_SECONDS", 60)) * time.Second,
			AccessTokenLifetime: time.Duration(
				positiveIntWithDefault("AUTH_ACCESS_TOKEN_LIFETIME_MINUTES", 15)) * time.Minute,
			RefreshTokenLifetime: time.Duration(
				positiveIntWithDefault("AUTH_REFRESH_TOKEN_LIFETIME_DAYS", 30)) * 24 * time.Hour,
		},
		RefreshCookie: RefreshCookieConfig{
			Domain:   strings.TrimSpace(os.Getenv("AUTH_REFRESH_COOKIE_DOMAIN")),
			IsSecure: boolWithDefault("AUTH_REFRESH_COOKIE_SECURE", true),
		},
		AccessTokenSigningKey: os.Getenv("AUTH_ACCESS_TOKEN_SIGNING_KEY"),
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

func boolWithDefault(name string, fallback bool) bool {
	value, parseError := strconv.ParseBool(stringWithDefault(name, strconv.FormatBool(fallback)))
	if parseError != nil {
		log.Printf("%s is not true or false; using %t", name, fallback)
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
