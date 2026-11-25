import type { MonthDto } from '~/domain/models/dto/month-dto'
import type { MonthSummaryDto } from '~/domain/models/dto/month-summary-dto'

/** 月報頁：選的是哪個月，以及那個月的月報（`null` 代表那個月板子沒有資料）。 */
export class MonthSummaryPageDto {
  constructor(
    readonly month: MonthDto,
    readonly summary: MonthSummaryDto | null,
  ) {}
}
