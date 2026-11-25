import Decimal from 'decimal.js'
import type { ITradeCardProxy } from '~/domain/interface/i-trade-card-proxy'
import type { ExitReason } from '~/domain/models/vo/exit-reason'
import type { OrderStatus } from '~/domain/models/vo/order-status'
import type { TradeCardStatus } from '~/domain/models/vo/trade-card-status'
import { BuyOrder } from '~/domain/models/entities/buy-order'
import { Fill } from '~/domain/models/entities/fill'
import { SellOrder } from '~/domain/models/entities/sell-order'
import { TradeCard } from '~/domain/models/entities/trade-card'
import { TradeCardSummary } from '~/domain/models/entities/trade-card-summary'
import { BoardApiProxy } from '~/infrastructure/proxy/board-api-proxy'

/** 後端回傳的原始 wire 形狀（契約 §4），只存在於本檔內。金額是字串。 */
type FillWire = {
  id: string
  price: string
  shares: number
  fee: string
  amount: string
  filledAt: string
}

type BuyOrderWire = {
  id: string
  symbol: string
  name: string
  limitPrice: string
  maxChasePrice: string
  shares: number
  estimatedAmount: string
  stopLossPrice: string
  takeProfitPrice: string
  reason?: string | null
  isMonthEndAllocation?: boolean
  status: OrderStatus
  filledShares?: number
  fills?: FillWire[] | null
}

type SellOrderWire = {
  id: string
  symbol: string
  name: string
  limitPrice: string
  shares: number
  exitReason: ExitReason
  reason?: string | null
  status: OrderStatus
  filledShares?: number
  fills?: FillWire[] | null
}

type TradeCardWire = {
  tradeDate: string
  status: TradeCardStatus
  note?: string | null
  buyOrders?: BuyOrderWire[] | null
  sellOrders?: SellOrderWire[] | null
  publishedAt?: string | null
  publishedBy?: string | null
}

type TradeCardSummaryWire = {
  tradeDate: string
  status: TradeCardStatus
  buyCount: number
  sellCount: number
  filledOrderCount: number
}

/** Proxy：`GET /cards/latest`、`GET /cards/{tradeDate}`、`GET /cards?month=`。 */
export class TradeCardProxy extends BoardApiProxy implements ITradeCardProxy {
  async fetchLatestCard(): Promise<TradeCard | null> {
    const tradeCardWire = await this.requestBoardOrNull<TradeCardWire>('/cards/latest')

    return tradeCardWire === null ? null : this.toTradeCard(tradeCardWire)
  }

  async fetchCard(tradeDate: string): Promise<TradeCard | null> {
    const tradeCardWire = await this.requestBoardOrNull<TradeCardWire>(`/cards/${encodeURIComponent(tradeDate)}`)

    return tradeCardWire === null ? null : this.toTradeCard(tradeCardWire)
  }

  async listCards(month: string): Promise<TradeCardSummary[]> {
    const summaryWires = await this.requestBoard<TradeCardSummaryWire[]>('/cards', { query: { month } })

    return summaryWires.map(summaryWire => new TradeCardSummary(
      summaryWire.tradeDate,
      summaryWire.status,
      summaryWire.buyCount,
      summaryWire.sellCount,
      summaryWire.filledOrderCount,
    ))
  }

  /** 被兩個讀卡的方法共用：wire → entity，null 收成明確的預設值。 */
  private toTradeCard(tradeCardWire: TradeCardWire): TradeCard {
    const toFill = (fillWire: FillWire): Fill => new Fill(
      fillWire.id,
      new Decimal(fillWire.price),
      fillWire.shares,
      new Decimal(fillWire.fee),
      new Decimal(fillWire.amount),
      new Date(fillWire.filledAt),
    )

    return new TradeCard(
      tradeCardWire.tradeDate,
      tradeCardWire.status,
      tradeCardWire.note ?? '',
      (tradeCardWire.buyOrders ?? []).map(orderWire => new BuyOrder(
        orderWire.id,
        orderWire.symbol,
        orderWire.name,
        new Decimal(orderWire.limitPrice),
        new Decimal(orderWire.maxChasePrice),
        orderWire.shares,
        new Decimal(orderWire.estimatedAmount),
        new Decimal(orderWire.stopLossPrice),
        new Decimal(orderWire.takeProfitPrice),
        orderWire.reason ?? '',
        orderWire.isMonthEndAllocation ?? false,
        orderWire.status,
        orderWire.filledShares ?? 0,
        (orderWire.fills ?? []).map(toFill),
      )),
      (tradeCardWire.sellOrders ?? []).map(orderWire => new SellOrder(
        orderWire.id,
        orderWire.symbol,
        orderWire.name,
        new Decimal(orderWire.limitPrice),
        orderWire.shares,
        orderWire.exitReason,
        orderWire.reason ?? '',
        orderWire.status,
        orderWire.filledShares ?? 0,
        (orderWire.fills ?? []).map(toFill),
      )),
      tradeCardWire.publishedAt ? new Date(tradeCardWire.publishedAt) : null,
      tradeCardWire.publishedBy ?? '',
    )
  }
}
