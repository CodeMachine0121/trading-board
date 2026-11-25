package models

import "github.com/CodeMachine0121/trading-board/backend/internal/domain/models/dto"

// TradeCardRequest is the PUT body; its tradeDate is optional and, when present, must match the path.
type TradeCardRequest struct {
	TradeDate  string             `json:"tradeDate"`
	Status     string             `json:"status"`
	Note       string             `json:"note"`
	BuyOrders  []BuyOrderRequest  `json:"buyOrders"`
	SellOrders []SellOrderRequest `json:"sellOrders"`
}

func (cardRequest TradeCardRequest) ToPublishDto(tradeDate string) dto.TradeCardPublishDto {
	buyOrders := make([]dto.BuyOrderWriteDto, 0, len(cardRequest.BuyOrders))
	for _, buyOrder := range cardRequest.BuyOrders {
		buyOrders = append(buyOrders, buyOrder.ToWriteDto())
	}

	sellOrders := make([]dto.SellOrderWriteDto, 0, len(cardRequest.SellOrders))
	for _, sellOrder := range cardRequest.SellOrders {
		sellOrders = append(sellOrders, sellOrder.ToWriteDto())
	}

	return dto.TradeCardPublishDto{
		TradeDate:  tradeDate,
		Status:     cardRequest.Status,
		Note:       cardRequest.Note,
		BuyOrders:  buyOrders,
		SellOrders: sellOrders,
	}
}
