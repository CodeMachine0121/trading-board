package domains

import "time"

const (
	dateLayout  = "2006-01-02"
	monthLayout = "2006-01"
)

// BusinessCalendarDomain reckons every date and month in the business time zone (Asia/Taipei).
type BusinessCalendarDomain struct {
	location *time.Location
}

// NewBusinessCalendarDomain falls back to UTC when no zone is given rather than failing every date.
func NewBusinessCalendarDomain(location *time.Location) BusinessCalendarDomain {
	if location == nil {
		location = time.UTC
	}

	return BusinessCalendarDomain{location: location}
}

func (calendar BusinessCalendarDomain) Location() *time.Location {
	return calendar.location
}

func (calendar BusinessCalendarDomain) DateOf(instant time.Time) string {
	return instant.In(calendar.location).Format(dateLayout)
}

func (calendar BusinessCalendarDomain) MonthOf(instant time.Time) string {
	return instant.In(calendar.location).Format(monthLayout)
}

func (calendar BusinessCalendarDomain) DayOfMonth(instant time.Time) int {
	return instant.In(calendar.location).Day()
}

func (calendar BusinessCalendarDomain) IsDate(text string) bool {
	parsed, parseError := time.Parse(dateLayout, text)

	return parseError == nil && parsed.Format(dateLayout) == text
}

func (calendar BusinessCalendarDomain) IsMonth(text string) bool {
	parsed, parseError := time.Parse(monthLayout, text)

	return parseError == nil && parsed.Format(monthLayout) == text
}

// NextMonth expects a month already checked with IsMonth.
func (calendar BusinessCalendarDomain) NextMonth(month string) string {
	parsed, _ := time.Parse(monthLayout, month)

	return parsed.AddDate(0, 1, 0).Format(monthLayout)
}

// DaysBetween counts calendar days from one date to a later one; both are checked dates.
func (calendar BusinessCalendarDomain) DaysBetween(fromDate string, toDate string) int {
	from, _ := time.Parse(dateLayout, fromDate)
	to, _ := time.Parse(dateLayout, toDate)

	return int(to.Sub(from).Hours() / 24)
}
