package config_test

import (
	"testing"
	"time"

	"github.com/CodeMachine0121/trading-board/backend/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestLoadUsesTheDefaultsWhenNothingIsSet(t *testing.T) {
	for _, name := range []string{"SERVER_PORT", "CORS_ALLOWED_ORIGINS", "POSTGRES_HOST", "POSTGRES_DATABASE",
		"AUTH_ACCESS_TOKEN_LIFETIME_MINUTES", "AUTH_REFRESH_TOKEN_LIFETIME_DAYS", "AUTH_AUTHORIZATION_CODE_LIFETIME_SECONDS",
		"OAUTH_CLIENT_ID", "OAUTH_REDIRECT_URIS", "AUTH_REFRESH_COOKIE_DOMAIN", "AUTH_REFRESH_COOKIE_SECURE",
		"FEE_DISCOUNT", "FEE_MINIMUM", "BUDGET_MONTHLY_ALLOCATION",
		"BUDGET_MONTH_END_START_DAY", "BUSINESS_TIME_ZONE", "OWNER_EMAIL", "OWNER_PASSWORD_HASH"} {
		t.Setenv(name, "")
	}

	applicationConfig := config.Load()

	assert.Equal(t, "8080", applicationConfig.ServerPort)
	assert.Equal(t, []string{"http://localhost:3000"}, applicationConfig.CorsAllowedOrigins)
	assert.Equal(t, "trading_board", applicationConfig.Database.Database)
	assert.Equal(t, 15*time.Minute, applicationConfig.AuthorizationPolicy.AccessTokenLifetime)
	assert.Equal(t, 30*24*time.Hour, applicationConfig.AuthorizationPolicy.RefreshTokenLifetime)
	assert.Equal(t, 60*time.Second, applicationConfig.AuthorizationPolicy.AuthorizationCodeLifetime)
	assert.Equal(t, "trading-board-web", applicationConfig.OAuthClient.ClientID)
	assert.Equal(t, []string{"http://localhost:3000/auth/callback"}, applicationConfig.OAuthClient.RedirectUris)
	assert.Equal(t, config.RefreshCookieConfig{Domain: "", IsSecure: true}, applicationConfig.RefreshCookie)
	assert.Equal(t, "0.6", applicationConfig.FeeSchedule.Discount.String())
	assert.Equal(t, "1", applicationConfig.FeeSchedule.MinimumFee.String())
	assert.Equal(t, "0.001425", applicationConfig.FeeSchedule.BrokerFeeRate.String())
	assert.Equal(t, "0.003", applicationConfig.FeeSchedule.TransactionTaxRate.String())
	assert.Equal(t, "10000", applicationConfig.BudgetPolicy.MonthlyAllocation.String())
	assert.Equal(t, 22, applicationConfig.BudgetPolicy.MonthEndStartDay)
	assert.Equal(t, "Asia/Taipei", applicationConfig.BusinessTimeZone.String())
	assert.Empty(t, applicationConfig.OwnerAccount.Email)
	assert.Empty(t, applicationConfig.OwnerAccount.PasswordHash)
}

func TestLoadReadsTheEnvironment(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://a.example, https://b.example")
	t.Setenv("POSTGRES_HOST", "postgres.go-trading.svc.cluster.local")
	t.Setenv("POSTGRES_PASSWORD", "pw")
	t.Setenv("OWNER_EMAIL", "owner@example.com")
	t.Setenv("OWNER_PASSWORD_HASH", "$2a$10$fake")
	t.Setenv("OAUTH_REDIRECT_URIS", "https://board.example/auth/callback, http://localhost:3000/auth/callback")
	t.Setenv("AUTH_ACCESS_TOKEN_LIFETIME_MINUTES", "5")
	t.Setenv("AUTH_REFRESH_TOKEN_LIFETIME_DAYS", "7")
	t.Setenv("AUTH_REFRESH_COOKIE_DOMAIN", "board.example")
	t.Setenv("AUTH_REFRESH_COOKIE_SECURE", "false")
	t.Setenv("FEE_DISCOUNT", "0.28")
	t.Setenv("BUDGET_MONTH_END_START_DAY", "20")

	applicationConfig := config.Load()

	assert.Equal(t, []string{"https://a.example", "https://b.example"}, applicationConfig.CorsAllowedOrigins)
	assert.Equal(t, "owner@example.com", applicationConfig.OwnerAccount.Email)
	assert.Equal(t, "$2a$10$fake", applicationConfig.OwnerAccount.PasswordHash)
	assert.Equal(t, []string{"https://board.example/auth/callback", "http://localhost:3000/auth/callback"},
		applicationConfig.OAuthClient.RedirectUris)
	assert.Equal(t, 5*time.Minute, applicationConfig.AuthorizationPolicy.AccessTokenLifetime)
	assert.Equal(t, 7*24*time.Hour, applicationConfig.AuthorizationPolicy.RefreshTokenLifetime)
	assert.Equal(t, config.RefreshCookieConfig{Domain: "board.example", IsSecure: false}, applicationConfig.RefreshCookie)
	assert.Equal(t, "0.28", applicationConfig.FeeSchedule.Discount.String())
	assert.Equal(t, 20, applicationConfig.BudgetPolicy.MonthEndStartDay)
	assert.Equal(t, "host=postgres.go-trading.svc.cluster.local port=5432 user=postgres password=pw dbname=trading_board sslmode=disable",
		applicationConfig.Database.DataSourceName())
}

func TestLoadFallsBackOnUnreadableValues(t *testing.T) {
	t.Setenv("AUTH_ACCESS_TOKEN_LIFETIME_MINUTES", "-3")
	t.Setenv("AUTH_REFRESH_COOKIE_SECURE", "maybe")
	t.Setenv("FEE_DISCOUNT", "cheap")
	t.Setenv("BUSINESS_TIME_ZONE", "Mars/Olympus")

	applicationConfig := config.Load()

	assert.Equal(t, 15*time.Minute, applicationConfig.AuthorizationPolicy.AccessTokenLifetime)
	assert.True(t, applicationConfig.RefreshCookie.IsSecure)
	assert.Equal(t, "0.6", applicationConfig.FeeSchedule.Discount.String())
	assert.Equal(t, "Asia/Taipei", applicationConfig.BusinessTimeZone.String())
}
