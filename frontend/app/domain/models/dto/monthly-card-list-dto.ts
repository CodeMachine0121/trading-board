import type { MonthDto } from '~/domain/models/dto/month-dto'
import type { TradeCardSummaryDto } from '~/domain/models/dto/trade-card-summary-dto'

/** 一個月的交易卡清單，新的在前。 */
export class MonthlyCardListDto {
  constructor(
    readonly month: MonthDto,
    readonly cards: readonly TradeCardSummaryDto[],
  ) {}
}
