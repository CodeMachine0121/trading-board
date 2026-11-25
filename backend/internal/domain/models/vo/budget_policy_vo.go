package vo

import "github.com/shopspring/decimal"

// BudgetPolicyVo holds the monthly allocation and the day of the month the month-end mode starts.
type BudgetPolicyVo struct {
	MonthlyAllocation decimal.Decimal
	MonthEndStartDay  int
}
