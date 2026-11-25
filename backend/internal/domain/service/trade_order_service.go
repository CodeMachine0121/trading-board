package service

import (
	"context"

	domaininterface "github.com/CodeMachine0121/trading-board/backend/internal/domain/interface"
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/domains"
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/dto"
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/entities"
)

// TradeOrderService records what happened to an order: a fill, or the owner skipping it.
type TradeOrderService struct {
	tradeOrderRepository  domaininterface.ITradeOrderRepository
	orderFillRepository   domaininterface.IOrderFillRepository
	transactionRepository domaininterface.ITransactionRepository
	ledgerLoader          *TradeLedgerLoader
	clockProxy            domaininterface.IClockProxy
	calendar              domains.BusinessCalendarDomain
	feeSchedule           domains.FeeScheduleDomain
}

func NewTradeOrderService(
	tradeOrderRepository domaininterface.ITradeOrderRepository,
	orderFillRepository domaininterface.IOrderFillRepository,
	transactionRepository domaininterface.ITransactionRepository,
	ledgerLoader *TradeLedgerLoader,
	clockProxy domaininterface.IClockProxy,
	calendar domains.BusinessCalendarDomain,
	feeSchedule domains.FeeScheduleDomain,
) *TradeOrderService {
	return &TradeOrderService{
		tradeOrderRepository:  tradeOrderRepository,
		orderFillRepository:   orderFillRepository,
		transactionRepository: transactionRepository,
		ledgerLoader:          ledgerLoader,
		clockProxy:            clockProxy,
		calendar:              calendar,
		feeSchedule:           feeSchedule,
	}
}

// ReportFill stores the fill only when the order, the holding and the monthly budget all allow it.
func (orderService *TradeOrderService) ReportFill(
	executionContext context.Context, orderID string, reportDto dto.OrderFillReportDto,
) (dto.OrderDto, error) {
	var orderDto dto.OrderDto
	reportError := orderService.transactionRepository.Exclusively(executionContext,
		func(transactionContext context.Context) error {
			order, findError := orderService.requireOrder(transactionContext, orderID)
			if findError != nil {
				return findError
			}

			orderDomain := domains.NewTradeOrderDomain(order, orderService.calendar)
			fill, fillError := orderDomain.ReportFill(reportDto, orderService.feeSchedule, orderService.clockProxy.Now())
			if fillError != nil {
				return fillError
			}

			ledger, loadError := orderService.ledgerLoader.Load(transactionContext)
			if loadError != nil {
				return loadError
			}
			if admitError := ledger.Admit(order.ToLedgerEntryVo(fill)); admitError != nil {
				return admitError
			}

			if _, createError := orderService.orderFillRepository.Create(transactionContext, fill); createError != nil {
				return createError
			}

			// Read back so the answer carries the identity the stored fill was given.
			storedOrder, _, reloadError := orderService.tradeOrderRepository.FindByID(transactionContext, orderID)
			if reloadError != nil {
				return reloadError
			}
			orderDto = domains.NewTradeOrderDomain(storedOrder, orderService.calendar).ToDto()

			return nil
		})
	if reportError != nil {
		return dto.OrderDto{}, reportError
	}

	return orderDto, nil
}

func (orderService *TradeOrderService) SkipOrder(
	executionContext context.Context, orderID string, skipDto dto.OrderSkipDto,
) (dto.OrderDto, error) {
	var orderDto dto.OrderDto
	skipError := orderService.transactionRepository.Exclusively(executionContext,
		func(transactionContext context.Context) error {
			order, findError := orderService.requireOrder(transactionContext, orderID)
			if findError != nil {
				return findError
			}

			orderDomain := domains.NewTradeOrderDomain(order, orderService.calendar)
			if skipError := orderDomain.Skip(skipDto); skipError != nil {
				return skipError
			}
			if markError := orderService.tradeOrderRepository.MarkSkipped(
				transactionContext, orderID, skipDto.SkippedBy); markError != nil {
				return markError
			}
			orderDto = orderDomain.ToDto()

			return nil
		})
	if skipError != nil {
		return dto.OrderDto{}, skipError
	}

	return orderDto, nil
}

// requireOrder answers not found for an order that does not exist, so every use case refuses it the same way.
func (orderService *TradeOrderService) requireOrder(
	executionContext context.Context, orderID string,
) (entities.TradeOrder, error) {
	order, isFound, findError := orderService.tradeOrderRepository.FindByID(executionContext, orderID)
	if findError != nil {
		return entities.TradeOrder{}, findError
	}
	if !isFound {
		return entities.TradeOrder{}, domains.NotFound("orderNotFound", "找不到這筆委託")
	}

	return order, nil
}
