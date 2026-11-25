package vo

// BudgetModeVo is only reported; the agent team applies the relaxed month-end rules itself.
type BudgetModeVo string

const (
	BudgetModeStrict   BudgetModeVo = "strict"
	BudgetModeMonthEnd BudgetModeVo = "monthEnd"
)
