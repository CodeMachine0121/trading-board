package dto

type TradeCardSummaryDto struct {
	TradeDate        string `json:"tradeDate"`
	Status           string `json:"status"`
	BuyCount         int    `json:"buyCount"`
	SellCount        int    `json:"sellCount"`
	FilledOrderCount int    `json:"filledOrderCount"`
	PublishedBy      string `json:"publishedBy"`
}
