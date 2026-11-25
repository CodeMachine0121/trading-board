package vo

// ExitReasonVo is why the agent team advises a sell.
type ExitReasonVo string

const (
	ExitReasonTakeProfit   ExitReasonVo = "takeProfit"
	ExitReasonStopLoss     ExitReasonVo = "stopLoss"
	ExitReasonTrailingStop ExitReasonVo = "trailingStop"
	ExitReasonTimeStop     ExitReasonVo = "timeStop"
)
