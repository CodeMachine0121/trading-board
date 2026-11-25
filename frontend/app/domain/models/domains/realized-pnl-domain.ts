import type Decimal from 'decimal.js'
import type { Tone } from '~/domain/models/vo/tone'
import { formatSignedTwd } from '~/utilities/decimal-formatter'

/**
 * 一筆已實現損益（需求 §3.5）。
 *
 * 台股慣例：獲利紅、虧損綠。大於零才算獲利——損益兩平算「虧損」，
 * 但顏色用中性色，不把一筆沒賠錢的交易畫成綠色。
 */
export class RealizedPnlDomain {
  constructor(private readonly amount: Decimal) {}

  tone(): Tone {
    if (this.amount.isZero()) {
      return 'neutral'
    }

    return this.amount.isPositive() ? 'profit' : 'loss'
  }

  outcomeLabel(): string {
    return this.amount.greaterThan(0) ? '獲利' : '虧損'
  }

  signedText(): string {
    return formatSignedTwd(this.amount)
  }
}
