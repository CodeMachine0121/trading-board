package entities

import "time"

// TradeCard is one Daily Trade Card; there is at most one per session date.
type TradeCard struct {
	ID          string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TradeDate   string    `gorm:"size:10;not null;uniqueIndex"`
	Status      string    `gorm:"size:16;not null"`
	Note        string    `gorm:"type:text;not null;default:''"`
	PublishedAt time.Time `gorm:"type:timestamptz;not null"`

	// Deleting a card takes its orders with it; replacing a card is only allowed while no order has a fill.
	Orders []TradeOrder `gorm:"foreignKey:TradeCardID;constraint:OnDelete:CASCADE"`
}

func (tradeCard TradeCard) TableName() string {
	return "TradeCards"
}
