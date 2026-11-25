package vo

// OrderSideVo tells a buy order from a sell order; both live in the same order table.
type OrderSideVo string

const (
	OrderSideBuy  OrderSideVo = "buy"
	OrderSideSell OrderSideVo = "sell"
)
