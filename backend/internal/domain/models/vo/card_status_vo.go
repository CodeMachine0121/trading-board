package vo

// CardStatusVo: a trade card carries orders, a no-trade card ("今日不買") carries none.
type CardStatusVo string

const (
	CardStatusTrade   CardStatusVo = "trade"
	CardStatusNoTrade CardStatusVo = "noTrade"
)
