import type { BuyOrder } from '~/domain/models/entities/buy-order'
import type { SellOrder } from '~/domain/models/entities/sell-order'
import type { TradeCardStatus } from '~/domain/models/vo/trade-card-status'
import type { TaipeiCalendarDomain } from '~/domain/models/domains/taipei-calendar-domain'
import { TradeCardDomain } from '~/domain/models/domains/trade-card-domain'

/** 一張交易卡：某個交易日要買什麼、要賣什麼（需求 §3.2）。 */
export class TradeCard {
  constructor(
    readonly tradeDate: string,
    readonly status: TradeCardStatus,
    readonly note: string,
    readonly buyOrders: readonly BuyOrder[],
    readonly sellOrders: readonly SellOrder[],
    readonly publishedAt: Date | null,
  ) {}

  toDomain(calendar: TaipeiCalendarDomain): TradeCardDomain {
    return new TradeCardDomain(this, calendar)
  }
}
