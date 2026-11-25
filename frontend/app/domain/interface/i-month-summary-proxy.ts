import type { MonthSummary } from '~/domain/models/entities/month-summary'

/** 板子算出來的月報。那個月板子沒有資料時回 `null`。 */
export interface IMonthSummaryProxy {
  fetchMonthSummary(month: string): Promise<MonthSummary | null>
}
