import type { TradeCard } from '~/domain/models/entities/trade-card'
import type { TradeCardSummary } from '~/domain/models/entities/trade-card-summary'

/** 板子上的交易卡。找不到時回 `null`，不拋。 */
export interface ITradeCardProxy {
  fetchLatestCard(): Promise<TradeCard | null>
  fetchCard(tradeDate: string): Promise<TradeCard | null>
  listCards(month: string): Promise<TradeCardSummary[]>
}
