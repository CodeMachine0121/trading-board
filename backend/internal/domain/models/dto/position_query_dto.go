package dto

// PositionQueryDto: an empty status means open; the month only narrows closed holdings, by close date.
type PositionQueryDto struct {
	Status string
	Month  string
}
