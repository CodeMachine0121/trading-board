package _interface

import "time"

//go:generate go tool mockgen -source=i_clock_proxy.go -destination=mocks/mock_i_clock_proxy.go -package=mocks

// IClockProxy reads the time so "today" and "this month" are testable.
type IClockProxy interface {
	Now() time.Time
}
