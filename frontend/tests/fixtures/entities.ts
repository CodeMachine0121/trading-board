import Decimal from 'decimal.js'
import { BuyOrder } from '~/domain/models/entities/buy-order'
import { Fill } from '~/domain/models/entities/fill'
import { MonthSummary } from '~/domain/models/entities/month-summary'
import { Position } from '~/domain/models/entities/position'
import { SellOrder } from '~/domain/models/entities/sell-order'
import { TradeCard } from '~/domain/models/entities/trade-card'
import type { BudgetMode } from '~/domain/models/vo/budget-mode'
import type { OrderStatus } from '~/domain/models/vo/order-status'

/** 測試用的固定資料：只把會影響結果的欄位開成參數，其餘用需求 §4 的範例值。 */

export function aFill(shares = 30, fee = '2'): Fill {
  return new Fill('fill-1', new Decimal('52.5'), shares, new Decimal(fee), new Decimal('52.5').times(shares),
    new Date('2026-11-26T01:05:00Z'))
}

export function aBuyOrder(
  status: OrderStatus = 'pending',
  filledShares = 0,
  isMonthEndAllocation = false,
  fills: Fill[] = [],
): BuyOrder {
  return new BuyOrder('buy-1', '2303', '聯電', new Decimal('52.5'), new Decimal('54'), 90, new Decimal('4725'),
    new Decimal('48'), new Decimal('60'), '底部突破', isMonthEndAllocation, status, filledShares, fills)
}

export function aSellOrder(status: OrderStatus = 'pending', filledShares = 0, fills: Fill[] = []): SellOrder {
  return new SellOrder('sell-1', '3006', '晶豪科', new Decimal('68'), 50, 'takeProfit', '到達停利', status, filledShares, fills)
}

export function aTradeCard(
  tradeDate = '2026-11-26',
  buyOrders: BuyOrder[] = [aBuyOrder()],
  sellOrders: SellOrder[] = [aSellOrder()],
): TradeCard {
  return new TradeCard(tradeDate, 'trade', '', buyOrders, sellOrders, new Date('2026-11-25T12:30:00Z'))
}

export function aNoTradeCard(note = '大盤轉弱'): TradeCard {
  return new TradeCard('2026-11-26', 'noTrade', note, [], [], null)
}

export function aMonthSummary(
  amounts: { allocation?: string, carriedOver?: string, recycledProceeds?: string, spent?: string, remaining?: string } = {},
  mode: BudgetMode = 'strict',
  performance: { realizedPnl?: string, closed?: number, wins?: number, losses?: number, returnPct?: string | null } = {},
): MonthSummary {
  return new MonthSummary(
    '2026-11',
    new Decimal(amounts.allocation ?? '10000'),
    new Decimal(amounts.carriedOver ?? '0'),
    new Decimal(amounts.recycledProceeds ?? '0'),
    new Decimal(amounts.spent ?? '4725'),
    new Decimal(amounts.remaining ?? '5275'),
    mode,
    new Decimal(performance.realizedPnl ?? '0'),
    performance.closed ?? 0,
    performance.wins ?? 0,
    performance.losses ?? 0,
    performance.returnPct === undefined ? null : performance.returnPct === null ? null : new Decimal(performance.returnPct),
  )
}

export function anOpenPosition(stopLossPrice: string | null = '48', takeProfitPrice: string | null = '60'): Position {
  return new Position('2303', '聯電', 90, new Decimal('52.58'), new Decimal('4732'),
    stopLossPrice === null ? null : new Decimal(stopLossPrice),
    takeProfitPrice === null ? null : new Decimal(takeProfitPrice), '2026-11-20', null, 5, new Decimal('0'))
}

export function aClosedPosition(realizedPnl: string, closedDate: string | null = '2026-11-24'): Position {
  return new Position('6770', '力積電', 0, new Decimal('0'), new Decimal('0'), null, null, '2026-11-18', closedDate, 6,
    new Decimal(realizedPnl))
}
