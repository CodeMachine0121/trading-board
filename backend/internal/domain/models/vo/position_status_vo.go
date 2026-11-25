package vo

// PositionStatusVo: a holding is open until its shares reach zero.
type PositionStatusVo string

const (
	PositionStatusOpen   PositionStatusVo = "open"
	PositionStatusClosed PositionStatusVo = "closed"
)
