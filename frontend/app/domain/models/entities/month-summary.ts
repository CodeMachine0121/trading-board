import type Decimal from 'decimal.js'
import type { BudgetMode } from '~/domain/models/vo/budget-mode'
import { MonthSummaryDomain } from '~/domain/models/domains/month-summary-domain'

/**
 * 一個月的預算（需求 §3.6）與已實現績效（需求 §3.7），金額全是新台幣。
 * `realizedReturnPct` 是百分比數字（`12.5` 代表 12.5%）；當月沒有出清時可能沒有。
 */
export class MonthSummary {
  constructor(
    readonly month: string,
    readonly allocation: Decimal,
    readonly carriedOver: Decimal,
    readonly recycledProceeds: Decimal,
    readonly spent: Decimal,
    readonly remaining: Decimal,
    readonly mode: BudgetMode,
    readonly realizedPnl: Decimal,
    readonly closedPositionCount: number,
    readonly winCount: number,
    readonly lossCount: number,
    readonly realizedReturnPct: Decimal | null,
  ) {}

  toDomain(): MonthSummaryDomain {
    return new MonthSummaryDomain(this)
  }
}
