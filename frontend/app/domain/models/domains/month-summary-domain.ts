import Decimal from 'decimal.js'
import type { MonthSummary } from '~/domain/models/entities/month-summary'
import type { BudgetMode } from '~/domain/models/vo/budget-mode'
import type { Tone } from '~/domain/models/vo/tone'
import { RealizedPnlDomain } from '~/domain/models/domains/realized-pnl-domain'
import { BudgetStripDto } from '~/domain/models/dto/budget-strip-dto'
import { DetailRowDto } from '~/domain/models/dto/detail-row-dto'
import { MonthSummaryDto } from '~/domain/models/dto/month-summary-dto'
import { formatPercent, formatTwd } from '~/utilities/decimal-formatter'

const MODE_LABELS: Record<BudgetMode, string> = {
  strict: '嚴格期',
  monthEnd: '月底配置期',
}

const MODE_TONES: Record<BudgetMode, Tone> = {
  strict: 'info',
  monthEnd: 'warning',
}

/** 當月沒有出清時，勝率與報酬率沒有意義，以這個字代替。 */
const NOT_APPLICABLE = '—'
const HUNDRED = 100

/** 一個月的預算與績效要怎麼給主人看（需求 §3.6、§3.7）。 */
export class MonthSummaryDomain {
  constructor(private readonly summary: MonthSummary) {}

  /**
   * 今日畫面頂端那一條。進度條是「本月可用的錢還剩幾成」：
   * 剩餘 ÷（月配額＋上月結轉＋賣出回收），夾在 0–100 之間；分母是零時是空的。
   */
  toBudgetStripDto(): BudgetStripDto {
    const available = this.summary.allocation.plus(this.summary.carriedOver).plus(this.summary.recycledProceeds)
    const remainingPercent = available.lessThanOrEqualTo(0)
      ? 0
      : Decimal.min(Decimal.max(this.summary.remaining.dividedBy(available).times(HUNDRED), 0), HUNDRED)
          .toDecimalPlaces(2, Decimal.ROUND_HALF_UP)
          .toNumber()

    return new BudgetStripDto(
      formatTwd(this.summary.remaining),
      formatTwd(this.summary.allocation),
      remainingPercent,
      MODE_LABELS[this.summary.mode],
      MODE_TONES[this.summary.mode],
    )
  }

  toDto(): MonthSummaryDto {
    const realizedPnl = new RealizedPnlDomain(this.summary.realizedPnl)
    const { closedPositionCount, winCount, lossCount, realizedReturnPct } = this.summary
    const hasClosedPositions = closedPositionCount > 0

    return new MonthSummaryDto(
      MODE_LABELS[this.summary.mode],
      MODE_TONES[this.summary.mode],
      [
        new DetailRowDto('月配額', formatTwd(this.summary.allocation)),
        new DetailRowDto('上月結轉', formatTwd(this.summary.carriedOver)),
        new DetailRowDto('賣出回收', formatTwd(this.summary.recycledProceeds)),
        new DetailRowDto('已投入', formatTwd(this.summary.spent)),
        new DetailRowDto('剩餘', formatTwd(this.summary.remaining)),
      ],
      realizedPnl.signedText(),
      realizedPnl.tone(),
      [
        new DetailRowDto('已出清', `${closedPositionCount} 檔`),
        new DetailRowDto('獲利', `${winCount} 檔`, 'profit'),
        new DetailRowDto('虧損', `${lossCount} 檔`, 'loss'),
        new DetailRowDto('勝率', hasClosedPositions
          ? formatPercent(new Decimal(winCount).dividedBy(closedPositionCount).times(HUNDRED))
          : NOT_APPLICABLE),
        new DetailRowDto('已實現報酬率', hasClosedPositions && realizedReturnPct !== null
          ? formatPercent(realizedReturnPct)
          : NOT_APPLICABLE),
      ],
    )
  }
}
