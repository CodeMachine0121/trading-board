package service

import (
	"context"
	"fmt"

	domaininterface "github.com/CodeMachine0121/trading-board/backend/internal/domain/interface"
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/domains"
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/dto"
)

// TradeCardService publishes and reads Daily Trade Cards.
type TradeCardService struct {
	tradeCardRepository   domaininterface.ITradeCardRepository
	transactionRepository domaininterface.ITransactionRepository
	ledgerLoader          *TradeLedgerLoader
	clockProxy            domaininterface.IClockProxy
	calendar              domains.BusinessCalendarDomain
}

func NewTradeCardService(
	tradeCardRepository domaininterface.ITradeCardRepository,
	transactionRepository domaininterface.ITransactionRepository,
	ledgerLoader *TradeLedgerLoader,
	clockProxy domaininterface.IClockProxy,
	calendar domains.BusinessCalendarDomain,
) *TradeCardService {
	return &TradeCardService{
		tradeCardRepository:   tradeCardRepository,
		transactionRepository: transactionRepository,
		ledgerLoader:          ledgerLoader,
		clockProxy:            clockProxy,
		calendar:              calendar,
	}
}

// PublishCard replaces an earlier card of the same session date only while none of its orders has a fill.
func (cardService *TradeCardService) PublishCard(
	executionContext context.Context, publishDto dto.TradeCardPublishDto,
) (dto.TradeCardDto, error) {
	publishedCard, validationError := domains.NewPublishedTradeCardDomain(
		publishDto, cardService.calendar, cardService.clockProxy.Now())
	if validationError != nil {
		return dto.TradeCardDto{}, validationError
	}

	var publishedDto dto.TradeCardDto
	publishError := cardService.transactionRepository.Exclusively(executionContext,
		func(transactionContext context.Context) error {
			existingCard, hasExisting, findError := cardService.tradeCardRepository.FindByTradeDate(
				transactionContext, publishedCard.TradeDate())
			if findError != nil {
				return findError
			}
			if hasExisting && domains.NewTradeCardDomain(existingCard, cardService.calendar).HasFills() {
				return domains.Conflict("cardHasFills", fmt.Sprintf(
					"%s 的交易卡已有成交，不能再覆蓋", publishedCard.TradeDate()))
			}

			ledger, loadError := cardService.ledgerLoader.Load(transactionContext)
			if loadError != nil {
				return loadError
			}
			if holdingError := ledger.RequireHoldings(publishedCard.SoldSharesBySymbol()); holdingError != nil {
				return holdingError
			}

			if hasExisting {
				if deleteError := cardService.tradeCardRepository.Delete(transactionContext, existingCard.ID); deleteError != nil {
					return deleteError
				}
			}

			createdCard, createError := cardService.tradeCardRepository.Create(transactionContext, publishedCard.ToEntity())
			if createError != nil {
				return createError
			}
			publishedDto = domains.NewTradeCardDomain(createdCard, cardService.calendar).ToDto()

			return nil
		})
	if publishError != nil {
		return dto.TradeCardDto{}, publishError
	}

	return publishedDto, nil
}

func (cardService *TradeCardService) GetCard(
	executionContext context.Context, tradeDate string,
) (dto.TradeCardDto, error) {
	if !cardService.calendar.IsDate(tradeDate) {
		return dto.TradeCardDto{}, domains.ValidationFailed("invalidRequest", "交易日須為 YYYY-MM-DD")
	}

	card, isFound, findError := cardService.tradeCardRepository.FindByTradeDate(executionContext, tradeDate)
	if findError != nil {
		return dto.TradeCardDto{}, findError
	}
	if !isFound {
		return dto.TradeCardDto{}, domains.NotFound("cardNotFound", fmt.Sprintf("找不到 %s 的交易卡", tradeDate))
	}

	return domains.NewTradeCardDomain(card, cardService.calendar).ToDto(), nil
}

func (cardService *TradeCardService) GetLatestCard(executionContext context.Context) (dto.TradeCardDto, error) {
	card, isFound, findError := cardService.tradeCardRepository.FindLatest(executionContext)
	if findError != nil {
		return dto.TradeCardDto{}, findError
	}
	if !isFound {
		return dto.TradeCardDto{}, domains.NotFound("cardNotFound", "還沒有任何交易卡")
	}

	return domains.NewTradeCardDomain(card, cardService.calendar).ToDto(), nil
}

// ListCards reads the current month when no month is given.
func (cardService *TradeCardService) ListCards(
	executionContext context.Context, month string,
) ([]dto.TradeCardSummaryDto, error) {
	if month == "" {
		month = cardService.calendar.MonthOf(cardService.clockProxy.Now())
	}
	if !cardService.calendar.IsMonth(month) {
		return nil, domains.ValidationFailed("invalidRequest", "month 須為 YYYY-MM")
	}

	cards, listError := cardService.tradeCardRepository.ListByMonth(executionContext, month)
	if listError != nil {
		return nil, listError
	}

	summaryDtos := make([]dto.TradeCardSummaryDto, 0, len(cards))
	for _, card := range cards {
		summaryDtos = append(summaryDtos, domains.NewTradeCardDomain(card, cardService.calendar).ToSummaryDto())
	}

	return summaryDtos, nil
}
