package vo

// OrderStatusVo is derived from filled versus ordered shares, except skipped, which the owner sets.
type OrderStatusVo string

const (
	OrderStatusPending         OrderStatusVo = "pending"
	OrderStatusPartiallyFilled OrderStatusVo = "partiallyFilled"
	OrderStatusFilled          OrderStatusVo = "filled"
	OrderStatusSkipped         OrderStatusVo = "skipped"
)
