package domains

import (
	"time"

	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/entities"
)

const (
	// SignInFailuresConsidered is how many of the latest failures decide the lock.
	SignInFailuresConsidered = 5
	signInFailureWindow      = 15 * time.Minute
	signInLockDuration       = 15 * time.Minute
)

// SignInThrottleDomain locks sign-in for 15 minutes once 5 failures fall within 15 minutes. Attempts refused while
// locked are not recorded, so the lock runs from the fifth failure and is never extended by waiting attempts.
type SignInThrottleDomain struct {
	oldestConsidered time.Time
	newest           time.Time
	isFull           bool
}

// NewSignInThrottleDomain takes the latest failures in any order.
func NewSignInThrottleDomain(latestFailures []entities.SignInFailure) SignInThrottleDomain {
	if len(latestFailures) < SignInFailuresConsidered {
		return SignInThrottleDomain{}
	}

	throttle := SignInThrottleDomain{
		oldestConsidered: latestFailures[0].FailedAt, newest: latestFailures[0].FailedAt, isFull: true,
	}
	for _, failure := range latestFailures[:SignInFailuresConsidered] {
		if failure.FailedAt.Before(throttle.oldestConsidered) {
			throttle.oldestConsidered = failure.FailedAt
		}
		if failure.FailedAt.After(throttle.newest) {
			throttle.newest = failure.FailedAt
		}
	}

	return throttle
}

func (throttle SignInThrottleDomain) IsLockedAt(now time.Time) bool {
	return throttle.isFull &&
		throttle.newest.Sub(throttle.oldestConsidered) <= signInFailureWindow &&
		now.Before(throttle.newest.Add(signInLockDuration))
}
