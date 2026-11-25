package clock_test

import (
	"testing"
	"time"

	"github.com/CodeMachine0121/trading-board/backend/internal/infrastructure/clock"
	"github.com/stretchr/testify/assert"
)

func TestSystemClockProxyReadsTheWallClock(t *testing.T) {
	before := time.Now()

	now := clock.NewSystemClockProxy().Now()

	assert.False(t, now.Before(before))
	assert.False(t, now.After(time.Now()))
}
