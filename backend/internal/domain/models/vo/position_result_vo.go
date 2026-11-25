package vo

// PositionResultVo: a closed holding is a win only when its total realized profit is above zero.
type PositionResultVo string

const (
	PositionResultWin  PositionResultVo = "win"
	PositionResultLoss PositionResultVo = "loss"
)
