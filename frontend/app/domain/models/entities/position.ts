import type Decimal from 'decimal.js'
import { PositionDomain } from '~/domain/models/domains/position-domain'

/**
 * 一檔持股（需求 §3.5）。成本採平均成本法、含買進手續費。
 * 出清後 `shares` 為 0、`closedDate` 有值。停損停利來自該檔最近一筆買進委託，可能沒有。
 */
export class Position {
  constructor(
    readonly symbol: string,
    readonly name: string,
    readonly shares: number,
    readonly averageCost: Decimal,
    readonly totalCost: Decimal,
    readonly stopLossPrice: Decimal | null,
    readonly takeProfitPrice: Decimal | null,
    readonly openedDate: string,
    readonly closedDate: string | null,
    readonly daysHeld: number,
    readonly realizedPnl: Decimal,
  ) {}

  toDomain(): PositionDomain {
    return new PositionDomain(this)
  }
}
