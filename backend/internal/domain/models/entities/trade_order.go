package entities

import (
	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/vo"
	"github.com/shopspring/decimal"
)

// TradeOrder is a buy or a sell order on a card; side-only columns stay null on the other side.
type TradeOrder struct {
	ID          string `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TradeCardID string `gorm:"type:uuid;not null;index"`
	// TradeDate repeats the card's session date so a fill can be dated without loading the card.
	TradeDate string `gorm:"size:10;not null"`
	Side      string `gorm:"size:8;not null"`
	Sequence  int    `gorm:"not null"`
	Symbol    string `gorm:"size:16;not null;index"`
	Name      string `gorm:"size:64;not null;default:''"`

	LimitPrice           decimal.Decimal     `gorm:"type:numeric(18,4);not null"`
	MaxChasePrice        decimal.NullDecimal `gorm:"type:numeric(18,4)"`
	Shares               int                 `gorm:"not null"`
	EstimatedAmount      decimal.NullDecimal `gorm:"type:numeric(18,4)"`
	StopLossPrice        decimal.NullDecimal `gorm:"type:numeric(18,4)"`
	TakeProfitPrice      decimal.NullDecimal `gorm:"type:numeric(18,4)"`
	ExitReason           string              `gorm:"size:16;not null;default:''"`
	Reason               string              `gorm:"type:text;not null;default:''"`
	IsMonthEndAllocation bool                `gorm:"not null;default:false"`
	IsSkipped            bool                `gorm:"not null;default:false"`
	// SkippedBy is the member name or "owner" that skipped the order.
	SkippedBy string `gorm:"size:160;not null;default:''"`

	// Restrict, not cascade: a fill is the record of money that moved and is never deleted with its order.
	Fills []OrderFill `gorm:"foreignKey:TradeOrderID;constraint:OnDelete:RESTRICT"`
}

func (tradeOrder TradeOrder) TableName() string {
	return "TradeOrders"
}

// ToLedgerEntryVo pairs one of this order's fills with the order facts the ledger replays it with.
func (tradeOrder TradeOrder) ToLedgerEntryVo(fill OrderFill) vo.LedgerEntryVo {
	return vo.LedgerEntryVo{
		FillID:          fill.ID,
		Side:            vo.OrderSideVo(tradeOrder.Side),
		Symbol:          tradeOrder.Symbol,
		Name:            tradeOrder.Name,
		TradeDate:       tradeOrder.TradeDate,
		StopLossPrice:   tradeOrder.StopLossPrice,
		TakeProfitPrice: tradeOrder.TakeProfitPrice,
		Shares:          fill.Shares,
		Amount:          fill.Amount,
		Fee:             fill.Fee,
		FilledAt:        fill.FilledAt,
		RecordedAt:      fill.RecordedAt,
	}
}

func (tradeOrder TradeOrder) ToLedgerEntryVos() []vo.LedgerEntryVo {
	entries := make([]vo.LedgerEntryVo, 0, len(tradeOrder.Fills))
	for _, fill := range tradeOrder.Fills {
		entries = append(entries, tradeOrder.ToLedgerEntryVo(fill))
	}

	return entries
}
