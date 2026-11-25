package dto

// TradeCardPublishDto is the card the agent team publishes for one session date.
type TradeCardPublishDto struct {
	TradeDate  string
	Status     string
	Note       string
	BuyOrders  []BuyOrderWriteDto
	SellOrders []SellOrderWriteDto
}
