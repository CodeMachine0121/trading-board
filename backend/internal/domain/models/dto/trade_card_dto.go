package dto

import "time"

// TradeCardDto is the full card of the shared contract.
type TradeCardDto struct {
	TradeDate   string     `json:"tradeDate"`
	Status      string     `json:"status"`
	Note        string     `json:"note"`
	BuyOrders   []OrderDto `json:"buyOrders"`
	SellOrders  []OrderDto `json:"sellOrders"`
	PublishedAt time.Time  `json:"publishedAt"`
}
