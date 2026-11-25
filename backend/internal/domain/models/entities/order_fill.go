package entities

import (
	"time"

	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/dto"
	"github.com/shopspring/decimal"
)

// OrderFill is one reported fill; fills are append-only in v1.
type OrderFill struct {
	ID           string          `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TradeOrderID string          `gorm:"type:uuid;not null;index"`
	Price        decimal.Decimal `gorm:"type:numeric(18,4);not null"`
	Shares       int             `gorm:"not null"`
	Fee          decimal.Decimal `gorm:"type:numeric(18,4);not null"`
	Amount       decimal.Decimal `gorm:"type:numeric(18,4);not null"`
	FilledAt     time.Time       `gorm:"type:timestamptz;not null"`
	// RecordedAt orders the ledger: holdings are replayed in the order fills were reported.
	RecordedAt time.Time `gorm:"type:timestamptz;not null;index"`
}

func (orderFill OrderFill) TableName() string {
	return "OrderFills"
}

// ToDto shows the fill time in the business time zone so its offset is the one the owner lives in.
func (orderFill OrderFill) ToDto(location *time.Location) dto.OrderFillDto {
	return dto.OrderFillDto{
		ID:       orderFill.ID,
		Price:    orderFill.Price,
		Shares:   orderFill.Shares,
		Fee:      orderFill.Fee,
		Amount:   orderFill.Amount,
		FilledAt: orderFill.FilledAt.In(location),
	}
}
