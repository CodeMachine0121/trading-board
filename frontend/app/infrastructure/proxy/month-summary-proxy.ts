import Decimal from 'decimal.js'
import type { IMonthSummaryProxy } from '~/domain/interface/i-month-summary-proxy'
import type { BudgetMode } from '~/domain/models/vo/budget-mode'
import { MonthSummary } from '~/domain/models/entities/month-summary'
import { BoardApiProxy } from '~/infrastructure/proxy/board-api-proxy'

/**
 * 後端回傳的原始 wire 形狀，只存在於本檔內。
 *
 * 契約 §4 只說「§3.6、§3.7 的欄位」，這是依那些欄位名推定的扁平形狀，
 * 已提交契約變更申請（docs/contract-change-requests/frontend.md）。後端改形狀時只改這裡。
 * `winRate` 不讀：勝率由獲利檔數 ÷ 出清檔數算出，不必猜它是 0–1 還是百分比。
 */
type MonthSummaryWire = {
  month: string
  allocation: string
  carriedOver: string
  recycledProceeds: string
  spent: string
  remaining: string
  mode: BudgetMode
  realizedPnl: string
  closedPositionCount: number
  winCount: number
  lossCount: number
  realizedReturnPct?: string | null
}

/** Proxy：`GET /months/{month}/summary`。那個月沒有資料（404）時回 `null`。 */
export class MonthSummaryProxy extends BoardApiProxy implements IMonthSummaryProxy {
  async fetchMonthSummary(month: string): Promise<MonthSummary | null> {
    const summaryWire = await this.requestBoardOrNull<MonthSummaryWire>(`/months/${encodeURIComponent(month)}/summary`)
    if (summaryWire === null) {
      return null
    }

    return new MonthSummary(
      summaryWire.month,
      new Decimal(summaryWire.allocation),
      new Decimal(summaryWire.carriedOver),
      new Decimal(summaryWire.recycledProceeds),
      new Decimal(summaryWire.spent),
      new Decimal(summaryWire.remaining),
      summaryWire.mode === 'monthEnd' ? 'monthEnd' : 'strict',
      new Decimal(summaryWire.realizedPnl),
      summaryWire.closedPositionCount,
      summaryWire.winCount,
      summaryWire.lossCount,
      summaryWire.realizedReturnPct ? new Decimal(summaryWire.realizedReturnPct) : null,
    )
  }
}
