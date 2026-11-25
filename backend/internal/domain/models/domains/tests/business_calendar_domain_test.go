package domains_test

import (
	"testing"
	"time"

	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/domains"
	"github.com/stretchr/testify/assert"
)

func TestBusinessCalendarDomainReckonsInTheBusinessTimeZone(t *testing.T) {
	// 16:30 UTC on the 25th is already the 26th in Taipei.
	instant := time.Date(2026, 11, 25, 16, 30, 0, 0, time.UTC)

	assert.Equal(t, "2026-11-26", calendar.DateOf(instant))
	assert.Equal(t, "2026-11", calendar.MonthOf(instant))
	assert.Equal(t, 26, calendar.DayOfMonth(instant))
}

func TestBusinessCalendarDomainFallsBackToUtcWithoutAZone(t *testing.T) {
	utcCalendar := domains.NewBusinessCalendarDomain(nil)

	assert.Equal(t, "2026-11-25", utcCalendar.DateOf(time.Date(2026, 11, 25, 16, 30, 0, 0, time.UTC)))
}

func TestBusinessCalendarDomainRecognisesDatesAndMonths(t *testing.T) {
	cases := []struct {
		text    string
		isDate  bool
		isMonth bool
	}{
		{text: "2026-11-26", isDate: true, isMonth: false},
		{text: "2026-11", isDate: false, isMonth: true},
		{text: "2026-13", isDate: false, isMonth: false},
		{text: "2026-2-01", isDate: false, isMonth: false},
		{text: "", isDate: false, isMonth: false},
	}

	for _, testCase := range cases {
		t.Run(testCase.text, func(t *testing.T) {
			assert.Equal(t, testCase.isDate, calendar.IsDate(testCase.text))
			assert.Equal(t, testCase.isMonth, calendar.IsMonth(testCase.text))
		})
	}
}

func TestBusinessCalendarDomainCountsMonthsAndDays(t *testing.T) {
	assert.Equal(t, "2027-01", calendar.NextMonth("2026-12"))
	assert.Equal(t, 5, calendar.DaysBetween("2026-11-20", "2026-11-25"))
}
