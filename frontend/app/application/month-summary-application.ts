import type { MonthSummaryService } from '~/domain/service/month-summary-service'
import type { BudgetStripDto } from '~/domain/models/dto/budget-strip-dto'
import type { MonthSummaryPageDto } from '~/domain/models/dto/month-summary-page-dto'

/** Application：本月預算與月報。純 TypeScript，全程只碰 DTO。 */
export class MonthSummaryApplication {
  constructor(private readonly monthSummaryService: MonthSummaryService) {}

  async fetchBudgetStrip(): Promise<BudgetStripDto | null> {
    return this.monthSummaryService.fetchBudgetStrip()
  }

  async fetchMonthSummary(month?: string): Promise<MonthSummaryPageDto> {
    return this.monthSummaryService.fetchMonthSummary(month)
  }
}
