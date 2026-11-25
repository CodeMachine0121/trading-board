import Decimal from 'decimal.js'
import type { IPositionProxy } from '~/domain/interface/i-position-proxy'
import { Position } from '~/domain/models/entities/position'
import { BoardApiProxy } from '~/infrastructure/proxy/board-api-proxy'

/**
 * 後端回傳的原始 wire 形狀，只存在於本檔內。
 *
 * 契約 §4 沒有寫出持股的欄位，這是依需求 §3.5 的用語推定的形狀，
 * 已提交契約變更申請（docs/contract-change-requests/frontend.md）。後端改名時只改這裡。
 */
type PositionWire = {
  symbol: string
  name: string
  shares: number
  averageCost: string
  totalCost: string
  stopLossPrice?: string | null
  takeProfitPrice?: string | null
  openedDate: string
  closedDate?: string | null
  daysHeld: number
  realizedPnl?: string | null
}

/** Proxy：`GET /positions?status=open|closed&month=`。 */
export class PositionProxy extends BoardApiProxy implements IPositionProxy {
  async listOpenPositions(): Promise<Position[]> {
    return this.toPositions(await this.requestBoard<PositionWire[]>('/positions', { query: { status: 'open' } }))
  }

  async listClosedPositions(month: string): Promise<Position[]> {
    return this.toPositions(await this.requestBoard<PositionWire[]>('/positions', { query: { status: 'closed', month } }))
  }

  private toPositions(positionWires: PositionWire[]): Position[] {
    return positionWires.map(positionWire => new Position(
      positionWire.symbol,
      positionWire.name,
      positionWire.shares,
      new Decimal(positionWire.averageCost),
      new Decimal(positionWire.totalCost),
      positionWire.stopLossPrice ? new Decimal(positionWire.stopLossPrice) : null,
      positionWire.takeProfitPrice ? new Decimal(positionWire.takeProfitPrice) : null,
      positionWire.openedDate,
      positionWire.closedDate ?? null,
      positionWire.daysHeld,
      new Decimal(positionWire.realizedPnl ?? '0'),
    ))
  }
}
