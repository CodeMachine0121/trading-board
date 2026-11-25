package service

import (
	"context"

	domaininterface "github.com/CodeMachine0121/trading-board/backend/internal/domain/interface"
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/domains"
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/vo"
)

// TradeLedgerLoader builds the ledger from the whole fill log; every service that needs a derived figure goes through it.
type TradeLedgerLoader struct {
	tradeOrderRepository domaininterface.ITradeOrderRepository
	tradeCardRepository  domaininterface.ITradeCardRepository
	clockProxy           domaininterface.IClockProxy
	calendar             domains.BusinessCalendarDomain
	budgetPolicy         vo.BudgetPolicyVo
}

func NewTradeLedgerLoader(
	tradeOrderRepository domaininterface.ITradeOrderRepository,
	tradeCardRepository domaininterface.ITradeCardRepository,
	clockProxy domaininterface.IClockProxy,
	calendar domains.BusinessCalendarDomain,
	budgetPolicy vo.BudgetPolicyVo,
) *TradeLedgerLoader {
	return &TradeLedgerLoader{
		tradeOrderRepository: tradeOrderRepository,
		tradeCardRepository:  tradeCardRepository,
		clockProxy:           clockProxy,
		calendar:             calendar,
		budgetPolicy:         budgetPolicy,
	}
}

func (ledgerLoader *TradeLedgerLoader) Load(executionContext context.Context) (domains.TradeLedgerDomain, error) {
	ordersWithFills, listError := ledgerLoader.tradeOrderRepository.ListWithFills(executionContext)
	if listError != nil {
		return domains.TradeLedgerDomain{}, listError
	}

	earliestCardDate, _, earliestError := ledgerLoader.tradeCardRepository.FindEarliestTradeDate(executionContext)
	if earliestError != nil {
		return domains.TradeLedgerDomain{}, earliestError
	}

	entries := []vo.LedgerEntryVo{}
	for _, order := range ordersWithFills {
		entries = append(entries, order.ToLedgerEntryVos()...)
	}

	return domains.NewTradeLedgerDomain(
		entries, earliestCardDate, ledgerLoader.calendar, ledgerLoader.budgetPolicy, ledgerLoader.clockProxy.Now()), nil
}
