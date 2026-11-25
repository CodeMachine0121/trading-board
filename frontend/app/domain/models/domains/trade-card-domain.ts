import type { TradeCard } from '~/domain/models/entities/trade-card'
import type { TaipeiCalendarDomain } from '~/domain/models/domains/taipei-calendar-domain'
import { TradeCardDto } from '~/domain/models/dto/trade-card-dto'
import { formatCalendarDate, formatTaipeiDateTime } from '~/utilities/date-formatter'

/**
 * 一張交易卡在某一天被看到時的樣子。
 *
 * 同一張卡今天看與明天看不一樣：交易日過了之後，還沒成交的委託就成了「已過期」。
 * 所以它要知道「今天」是哪一天，而那一天以台北為準。
 */
export class TradeCardDomain {
  constructor(
    private readonly card: TradeCard,
    private readonly calendar: TaipeiCalendarDomain,
  ) {}

  toDto(): TradeCardDto {
    const sessionIsPast = this.calendar.isPastSession(this.card.tradeDate)
    const defaultFilledAt = this.calendar.nowInputText()
    const isNoTrade = this.card.status === 'noTrade'

    return new TradeCardDto(
      this.card.tradeDate,
      formatCalendarDate(this.card.tradeDate),
      isNoTrade,
      this.card.note,
      isNoTrade
        ? []
        : this.card.buyOrders.map(order =>
            order.toDto(this.card.tradeDate, order.toProgressDomain(sessionIsPast).toVo(), defaultFilledAt)),
      isNoTrade
        ? []
        : this.card.sellOrders.map(order =>
            order.toDto(this.card.tradeDate, order.toProgressDomain(sessionIsPast).toVo(), defaultFilledAt)),
      '無買進委託',
      '無賣出委託',
      this.card.publishedAt === null ? '' : `發布於 ${formatTaipeiDateTime(this.card.publishedAt)}`,
      this.card.publishedBy === '' ? '' : `發布者：${this.card.publishedBy}`,
    )
  }
}
