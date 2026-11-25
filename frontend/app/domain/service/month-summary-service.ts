import type { IMonthSummaryProxy } from '~/domain/interface/i-month-summary-proxy'
import type { BudgetStripDto } from '~/domain/models/dto/budget-strip-dto'
import { TaipeiCalendarDomain } from '~/domain/models/domains/taipei-calendar-domain'
import { MonthSummaryPageDto } from '~/domain/models/dto/month-summary-page-dto'

/** 月報：本月預算那一條，以及某個月的完整月報。 */
export class MonthSummaryService {
  constructor(private readonly monthSummaryProxy: IMonthSummaryProxy) {}

  /** 本月（台北）的預算；板子還沒有本月的資料時回 `null`。 */
  async fetchBudgetStrip(): Promise<BudgetStripDto | null> {
    const currentMonth = new TaipeiCalendarDomain(new Date()).toMonthDto().value
    const summary = await this.monthSummaryProxy.fetchMonthSummary(currentMonth)

    return summary === null ? null : summary.toDomain().toBudgetStripDto()
  }

  async fetchMonthSummary(month?: string): Promise<MonthSummaryPageDto> {
    const monthDto = new TaipeiCalendarDomain(new Date()).toMonthDto(month)
    const summary = await this.monthSummaryProxy.fetchMonthSummary(monthDto.value)

    return new MonthSummaryPageDto(monthDto, summary === null ? null : summary.toDomain().toDto())
  }
}
