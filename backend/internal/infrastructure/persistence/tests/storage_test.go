package persistence_test

import (
	"context"
	"errors"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/entities"
	"github.com/CodeMachine0121/trading-board/backend/internal/infrastructure/persistence"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// testDatabaseNameSuffix is required on the test database name because these tests empty every table.
const testDatabaseNameSuffix = "_test"

// newTestDatabase connects through TEST_POSTGRES_DSN, migrates and empties the tables; it skips when the
// variable is unset, which CI refuses to accept as a pass.
func newTestDatabase(t *testing.T) *gorm.DB {
	dataSourceName := os.Getenv("TEST_POSTGRES_DSN")
	if dataSourceName == "" {
		t.Skip("TEST_POSTGRES_DSN is not set; skipping storage tests")
	}
	require.True(t, strings.Contains(dataSourceName, testDatabaseNameSuffix+" ") ||
		strings.HasSuffix(dataSourceName, testDatabaseNameSuffix),
		"TEST_POSTGRES_DSN must name a database ending in %s: these tests empty every table", testDatabaseNameSuffix)

	database, openError := persistence.NewDatabase(dataSourceName)
	require.NoError(t, openError)
	database.Logger = logger.Discard
	t.Cleanup(func() {
		if connection, connectionError := database.DB(); connectionError == nil {
			_ = connection.Close()
		}
	})

	_, migrateError := persistence.NewSchemaMigrator(database).Migrate()
	require.NoError(t, migrateError)
	emptied := database.Session(&gorm.Session{AllowGlobalUpdate: true})
	require.NoError(t, emptied.Delete(&entities.OrderFill{}).Error)
	require.NoError(t, emptied.Delete(&entities.TradeOrder{}).Error)
	require.NoError(t, emptied.Delete(&entities.TradeCard{}).Error)

	return database
}

// closedDatabase is a connection whose pool is already closed, so every call fails the way an outage would.
func closedDatabase(t *testing.T) *gorm.DB {
	database := newTestDatabase(t)
	connection, connectionError := database.DB()
	require.NoError(t, connectionError)
	require.NoError(t, connection.Close())

	return database
}

func aCard(tradeDate string, symbols ...string) entities.TradeCard {
	card := entities.TradeCard{TradeDate: tradeDate, Status: "trade", PublishedAt: time.Date(2026, 11, 25, 12, 0, 0, 0, time.UTC)}
	for index, symbol := range symbols {
		card.Orders = append(card.Orders, entities.TradeOrder{
			TradeDate: tradeDate, Side: "buy", Sequence: index, Symbol: symbol, Shares: 90,
			LimitPrice: decimal.RequireFromString("52.5"), StopLossPrice: decimal.NewNullDecimal(decimal.RequireFromString("48")),
		})
	}

	return card
}

func aFill(orderID string, shares int, recordedAt time.Time) entities.OrderFill {
	return entities.OrderFill{TradeOrderID: orderID, Price: decimal.RequireFromString("52.5"), Shares: shares,
		Fee: decimal.RequireFromString("4"), Amount: decimal.RequireFromString("4725"),
		FilledAt: recordedAt, RecordedAt: recordedAt}
}

func TestTradeCardRepositoryStoresACardWithItsOrders(t *testing.T) {
	database := newTestDatabase(t)
	repository := persistence.NewTradeCardRepository(database)

	created, createError := repository.Create(t.Context(), aCard("2026-11-26", "2303", "3006"))
	found, isFound, findError := repository.FindByTradeDate(t.Context(), "2026-11-26")

	require.NoError(t, createError)
	require.NoError(t, findError)
	require.True(t, isFound)
	assert.NotEmpty(t, created.ID)
	assert.Equal(t, created.ID, found.ID)
	require.Len(t, found.Orders, 2)
	assert.Equal(t, []string{"2303", "3006"}, []string{found.Orders[0].Symbol, found.Orders[1].Symbol})
	assert.NotEmpty(t, found.Orders[0].ID)
	assert.Equal(t, "52.5", found.Orders[0].LimitPrice.String())
	assert.Equal(t, "48", found.Orders[0].StopLossPrice.Decimal.String())
	assert.False(t, found.Orders[0].MaxChasePrice.Valid)
}

func TestTradeCardRepositoryKeepsOneCardPerSessionDate(t *testing.T) {
	repository := persistence.NewTradeCardRepository(newTestDatabase(t))
	_, firstError := repository.Create(t.Context(), aCard("2026-11-26", "2303"))

	_, secondError := repository.Create(t.Context(), aCard("2026-11-26", "3006"))

	require.NoError(t, firstError)
	assert.Error(t, secondError)
}

func TestTradeCardRepositoryFindsLatestAndEarliestAndListsAMonth(t *testing.T) {
	repository := persistence.NewTradeCardRepository(newTestDatabase(t))
	for _, tradeDate := range []string{"2026-11-25", "2026-12-01", "2026-10-31", "2026-11-26"} {
		_, createError := repository.Create(t.Context(), aCard(tradeDate, "2303"))
		require.NoError(t, createError)
	}

	latest, hasLatest, latestError := repository.FindLatest(t.Context())
	earliest, hasEarliest, earliestError := repository.FindEarliestTradeDate(t.Context())
	november, listError := repository.ListByMonth(t.Context(), "2026-11")

	require.NoError(t, latestError)
	require.NoError(t, earliestError)
	require.NoError(t, listError)
	assert.True(t, hasLatest)
	assert.Equal(t, "2026-12-01", latest.TradeDate)
	assert.True(t, hasEarliest)
	assert.Equal(t, "2026-10-31", earliest)
	require.Len(t, november, 2)
	assert.Equal(t, []string{"2026-11-26", "2026-11-25"}, []string{november[0].TradeDate, november[1].TradeDate})
	assert.Len(t, november[0].Orders, 1)
}

func TestTradeCardRepositoryAnswersNoneWhenEmpty(t *testing.T) {
	repository := persistence.NewTradeCardRepository(newTestDatabase(t))

	_, hasLatest, latestError := repository.FindLatest(t.Context())
	_, hasCard, cardError := repository.FindByTradeDate(t.Context(), "2026-11-26")
	earliest, hasEarliest, earliestError := repository.FindEarliestTradeDate(t.Context())
	cards, listError := repository.ListByMonth(t.Context(), "2026-11")

	require.NoError(t, errors.Join(latestError, cardError, earliestError, listError))
	assert.False(t, hasLatest)
	assert.False(t, hasCard)
	assert.False(t, hasEarliest)
	assert.Empty(t, earliest)
	assert.Empty(t, cards)
}

func TestTradeCardRepositoryDeletesACardWithItsOrders(t *testing.T) {
	database := newTestDatabase(t)
	repository := persistence.NewTradeCardRepository(database)
	created, createError := repository.Create(t.Context(), aCard("2026-11-26", "2303"))
	require.NoError(t, createError)

	deleteError := repository.Delete(t.Context(), created.ID)

	require.NoError(t, deleteError)
	_, isFound, _ := repository.FindByTradeDate(t.Context(), "2026-11-26")
	assert.False(t, isFound)
	var remainingOrders int64
	require.NoError(t, database.Model(&entities.TradeOrder{}).Count(&remainingOrders).Error)
	assert.Zero(t, remainingOrders)
}

func TestTradeCardRepositoryCannotDeleteACardWhoseOrderHasAFill(t *testing.T) {
	database := newTestDatabase(t)
	repository := persistence.NewTradeCardRepository(database)
	created, _ := repository.Create(t.Context(), aCard("2026-11-26", "2303"))
	_, fillError := persistence.NewOrderFillRepository(database).Create(t.Context(), aFill(created.Orders[0].ID, 10, time.Now()))
	require.NoError(t, fillError)

	deleteError := repository.Delete(t.Context(), created.ID)

	assert.Error(t, deleteError)
}

func TestTradeOrderRepositoryFindsAnOrderWithItsFillsInReportedOrder(t *testing.T) {
	database := newTestDatabase(t)
	created, _ := persistence.NewTradeCardRepository(database).Create(t.Context(), aCard("2026-11-26", "2303", "3006"))
	fillRepository := persistence.NewOrderFillRepository(database)
	orderID := created.Orders[0].ID
	second, _ := fillRepository.Create(t.Context(), aFill(orderID, 50, time.Date(2026, 11, 26, 2, 0, 0, 0, time.UTC)))
	first, _ := fillRepository.Create(t.Context(), aFill(orderID, 40, time.Date(2026, 11, 26, 1, 0, 0, 0, time.UTC)))
	repository := persistence.NewTradeOrderRepository(database)

	order, isFound, findError := repository.FindByID(t.Context(), orderID)
	withFills, listError := repository.ListWithFills(t.Context())

	require.NoError(t, findError)
	require.NoError(t, listError)
	require.True(t, isFound)
	require.Len(t, order.Fills, 2)
	assert.Equal(t, []string{first.ID, second.ID}, []string{order.Fills[0].ID, order.Fills[1].ID})
	assert.Equal(t, "2026-11-26", order.TradeDate)
	require.Len(t, withFills, 1)
	assert.Equal(t, orderID, withFills[0].ID)
	assert.Len(t, withFills[0].Fills, 2)
}

func TestTradeOrderRepositoryAnswersNoneForUnknownOrMalformedIdentities(t *testing.T) {
	repository := persistence.NewTradeOrderRepository(newTestDatabase(t))

	for _, orderID := range []string{"not-a-uuid", "6f1c1e0e-8f9e-4c7e-9a43-0d3b6f0b1c2d"} {
		_, isFound, findError := repository.FindByID(t.Context(), orderID)

		require.NoError(t, findError)
		assert.False(t, isFound, orderID)
	}
}

func TestTradeOrderRepositoryMarksAnOrderSkipped(t *testing.T) {
	database := newTestDatabase(t)
	created, _ := persistence.NewTradeCardRepository(database).Create(t.Context(), aCard("2026-11-26", "2303"))
	repository := persistence.NewTradeOrderRepository(database)

	markError := repository.MarkSkipped(t.Context(), created.Orders[0].ID)

	require.NoError(t, markError)
	order, _, _ := repository.FindByID(t.Context(), created.Orders[0].ID)
	assert.True(t, order.IsSkipped)
}

func TestRepositoriesReportStorageFailures(t *testing.T) {
	database := closedDatabase(t)
	cardRepository := persistence.NewTradeCardRepository(database)
	orderRepository := persistence.NewTradeOrderRepository(database)
	fillRepository := persistence.NewOrderFillRepository(database)

	_, _, findError := cardRepository.FindByTradeDate(t.Context(), "2026-11-26")
	_, listError := cardRepository.ListByMonth(t.Context(), "2026-11")
	_, createError := cardRepository.Create(t.Context(), aCard("2026-11-26"))
	deleteError := cardRepository.Delete(t.Context(), "6f1c1e0e-8f9e-4c7e-9a43-0d3b6f0b1c2d")
	_, _, orderError := orderRepository.FindByID(t.Context(), "6f1c1e0e-8f9e-4c7e-9a43-0d3b6f0b1c2d")
	_, withFillsError := orderRepository.ListWithFills(t.Context())
	markError := orderRepository.MarkSkipped(t.Context(), "6f1c1e0e-8f9e-4c7e-9a43-0d3b6f0b1c2d")
	_, fillError := fillRepository.Create(t.Context(), aFill("6f1c1e0e-8f9e-4c7e-9a43-0d3b6f0b1c2d", 1, time.Now()))
	transactionError := persistence.NewTransactionRepository(database).Exclusively(t.Context(),
		func(context.Context) error { return nil })
	_, migrateError := persistence.NewSchemaMigrator(database).Migrate()

	for _, storageError := range []error{findError, listError, createError, deleteError, orderError, withFillsError,
		markError, fillError, transactionError, migrateError} {
		assert.Error(t, storageError)
	}
}

func TestTradeCardRepositoryRefusesAnUnreadableMonth(t *testing.T) {
	_, listError := persistence.NewTradeCardRepository(newTestDatabase(t)).ListByMonth(t.Context(), "Nov")

	assert.Error(t, listError)
}

func TestTransactionRepositoryCommitsOrRollsBackTogether(t *testing.T) {
	database := newTestDatabase(t)
	transactionRepository := persistence.NewTransactionRepository(database)
	cardRepository := persistence.NewTradeCardRepository(database)
	failure := errors.New("refused")

	rolledBack := transactionRepository.Exclusively(t.Context(), func(transactionContext context.Context) error {
		_, createError := cardRepository.Create(transactionContext, aCard("2026-11-25", "2303"))
		require.NoError(t, createError)

		return failure
	})
	committed := transactionRepository.Exclusively(t.Context(), func(transactionContext context.Context) error {
		return transactionRepository.Exclusively(transactionContext, func(innerContext context.Context) error {
			_, createError := cardRepository.Create(innerContext, aCard("2026-11-26", "2303"))

			return createError
		})
	})

	assert.ErrorIs(t, rolledBack, failure)
	require.NoError(t, committed)
	_, hasRolledBack, _ := cardRepository.FindByTradeDate(t.Context(), "2026-11-25")
	_, hasCommitted, _ := cardRepository.FindByTradeDate(t.Context(), "2026-11-26")
	assert.False(t, hasRolledBack)
	assert.True(t, hasCommitted)
}

func TestTransactionRepositoryRunsWritesOneAtATime(t *testing.T) {
	transactionRepository := persistence.NewTransactionRepository(newTestDatabase(t))
	firstHolds := make(chan struct{})
	releaseFirst := make(chan struct{})
	var order []string
	var orderLock sync.Mutex
	record := func(step string) {
		orderLock.Lock()
		defer orderLock.Unlock()
		order = append(order, step)
	}

	var waitGroup sync.WaitGroup
	waitGroup.Add(2)
	go func() {
		defer waitGroup.Done()
		_ = transactionRepository.Exclusively(t.Context(), func(context.Context) error {
			close(firstHolds)
			<-releaseFirst
			record("first done")

			return nil
		})
	}()
	<-firstHolds
	go func() {
		defer waitGroup.Done()
		_ = transactionRepository.Exclusively(t.Context(), func(context.Context) error {
			record("second ran")

			return nil
		})
	}()
	time.Sleep(200 * time.Millisecond)
	record("released")
	close(releaseFirst)
	waitGroup.Wait()

	assert.Equal(t, []string{"released", "first done", "second ran"}, order)
}

func TestSchemaMigratorSyncsEveryTableUnderItsLock(t *testing.T) {
	schemaMigrator := persistence.NewSchemaMigrator(newTestDatabase(t))

	var migratedTables []string
	exclusiveError := schemaMigrator.Exclusively(t.Context(), func() error {
		tables, migrateError := schemaMigrator.Migrate()
		migratedTables = tables

		return migrateError
	})

	require.NoError(t, exclusiveError)
	assert.Equal(t, []string{"TradeCards", "TradeOrders", "OrderFills"}, migratedTables)
}

func TestSchemaMigratorReportsALockFailure(t *testing.T) {
	exclusiveError := persistence.NewSchemaMigrator(closedDatabase(t)).Exclusively(t.Context(), func() error { return nil })

	assert.Error(t, exclusiveError)
}

func TestNewDatabaseRefusesAnUnreadableAddress(t *testing.T) {
	_, openError := persistence.NewDatabase("host=127.0.0.1 port=1 user=nobody dbname=none_test sslmode=disable connect_timeout=1")

	assert.Error(t, openError)
}

// impatientDatabase gives up waiting for any lock after 100ms, so a lock held elsewhere surfaces as an error.
func impatientDatabase(t *testing.T) *gorm.DB {
	database, openError := persistence.NewDatabase(os.Getenv("TEST_POSTGRES_DSN") + " options='-c lock_timeout=100'")
	require.NoError(t, openError)
	database.Logger = logger.Discard
	t.Cleanup(func() {
		if connection, connectionError := database.DB(); connectionError == nil {
			_ = connection.Close()
		}
	})

	return database
}

func TestExclusiveWorkFailsWhenTheLockCannotBeHad(t *testing.T) {
	database := newTestDatabase(t)
	impatient := impatientDatabase(t)
	holding := make(chan struct{})
	release := make(chan struct{})
	finished := make(chan struct{})

	go func() {
		defer close(finished)
		_ = persistence.NewTransactionRepository(database).Exclusively(t.Context(), func(context.Context) error {
			return persistence.NewSchemaMigrator(database).Exclusively(t.Context(), func() error {
				close(holding)
				<-release

				return nil
			})
		})
	}()
	<-holding

	transactionError := persistence.NewTransactionRepository(impatient).Exclusively(t.Context(),
		func(context.Context) error { return nil })
	migrationError := persistence.NewSchemaMigrator(impatient).Exclusively(t.Context(), func() error { return nil })
	close(release)
	<-finished

	assert.ErrorContains(t, transactionError, "wait for other writes")
	assert.ErrorContains(t, migrationError, "wait for other migrations")
}
