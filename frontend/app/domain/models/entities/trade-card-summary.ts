import type { TradeCardStatus } from '~/domain/models/vo/trade-card-status'
import { TradeCardSummaryDto } from '~/domain/models/dto/trade-card-summary-dto'
import { formatCalendarDate } from '~/utilities/date-formatter'

/** 月份清單裡的一張交易卡摘要。 */
export class TradeCardSummary {
  constructor(
    readonly tradeDate: string,
    readonly status: TradeCardStatus,
    readonly buyCount: number,
    readonly sellCount: number,
    readonly filledOrderCount: number,
  ) {}

  toDto(): TradeCardSummaryDto {
    const isNoTrade = this.status === 'noTrade'

    return new TradeCardSummaryDto(
      this.tradeDate,
      formatCalendarDate(this.tradeDate),
      isNoTrade ? '今日不買' : '交易',
      isNoTrade ? 'neutral' : 'info',
      `買 ${this.buyCount}・賣 ${this.sellCount}`,
      `已成交 ${this.filledOrderCount} 筆`,
    )
  }
}
