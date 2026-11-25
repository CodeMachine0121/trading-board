package config_test

import (
	"testing"
	"time"

	"github.com/CodeMachine0121/trading-board/backend/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestLoadUsesTheDefaultsWhenNothingIsSet(t *testing.T) {
	for _, name := range []string{"SERVER_PORT", "CORS_ALLOWED_ORIGINS", "POSTGRES_HOST", "POSTGRES_DATABASE",
		"AUTH_ACCESS_TOKEN_LIFETIME_HOURS", "FEE_DISCOUNT", "FEE_MINIMUM", "BUDGET_MONTHLY_ALLOCATION",
		"BUDGET_MONTH_END_START_DAY", "BUSINESS_TIME_ZONE", "OWNER_USERNAME"} {
		t.Setenv(name, "")
	}

	applicationConfig := config.Load()

	assert.Equal(t, "8080", applicationConfig.ServerPort)
	assert.Equal(t, []string{"http://localhost:3000"}, applicationConfig.CorsAllowedOrigins)
	assert.Equal(t, "trading_board", applicationConfig.Database.Database)
	assert.Equal(t, 24*time.Hour, applicationConfig.AccessTokenLifetime)
	assert.Equal(t, "0.6", applicationConfig.FeeSchedule.Discount.String())
	assert.Equal(t, "1", applicationConfig.FeeSchedule.MinimumFee.String())
	assert.Equal(t, "0.001425", applicationConfig.FeeSchedule.BrokerFeeRate.String())
	assert.Equal(t, "0.003", applicationConfig.FeeSchedule.TransactionTaxRate.String())
	assert.Equal(t, "10000", applicationConfig.BudgetPolicy.MonthlyAllocation.String())
	assert.Equal(t, 22, applicationConfig.BudgetPolicy.MonthEndStartDay)
	assert.Equal(t, "Asia/Taipei", applicationConfig.BusinessTimeZone.String())
	assert.Empty(t, applicationConfig.AccessCredential.OwnerUsername)
}

func TestLoadReadsTheEnvironment(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://a.example, https://b.example")
	t.Setenv("POSTGRES_HOST", "postgres.go-trading.svc.cluster.local")
	t.Setenv("POSTGRES_PASSWORD", "pw")
	t.Setenv("OWNER_USERNAME", "james")
	t.Setenv("AUTH_ACCESS_TOKEN_LIFETIME_HOURS", "12")
	t.Setenv("FEE_DISCOUNT", "0.28")
	t.Setenv("BUDGET_MONTH_END_START_DAY", "20")

	applicationConfig := config.Load()

	assert.Equal(t, []string{"https://a.example", "https://b.example"}, applicationConfig.CorsAllowedOrigins)
	assert.Equal(t, "james", applicationConfig.AccessCredential.OwnerUsername)
	assert.Equal(t, 12*time.Hour, applicationConfig.AccessTokenLifetime)
	assert.Equal(t, "0.28", applicationConfig.FeeSchedule.Discount.String())
	assert.Equal(t, 20, applicationConfig.BudgetPolicy.MonthEndStartDay)
	assert.Equal(t, "host=postgres.go-trading.svc.cluster.local port=5432 user=postgres password=pw dbname=trading_board sslmode=disable",
		applicationConfig.Database.DataSourceName())
}

func TestLoadFallsBackOnUnreadableValues(t *testing.T) {
	t.Setenv("AUTH_ACCESS_TOKEN_LIFETIME_HOURS", "-3")
	t.Setenv("FEE_DISCOUNT", "cheap")
	t.Setenv("BUSINESS_TIME_ZONE", "Mars/Olympus")

	applicationConfig := config.Load()

	assert.Equal(t, 24*time.Hour, applicationConfig.AccessTokenLifetime)
	assert.Equal(t, "0.6", applicationConfig.FeeSchedule.Discount.String())
	assert.Equal(t, "Asia/Taipei", applicationConfig.BusinessTimeZone.String())
}
