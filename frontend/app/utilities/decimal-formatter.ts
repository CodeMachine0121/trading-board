import Decimal from 'decimal.js'

/**
 * 把精確小數排成畫面上的字。純技術性的格式轉換：不碰任何領域規則，
 * 誰呼叫都一樣（見 code-style.md 的 helper 門檻）。
 *
 * 一律從 Decimal 的字串表示出發，中途不經過 JavaScript number——經過就已經失真了。
 */

/** 千分位：`4725.5` → `4,725.5`。保留原本的小數位數，不四捨五入。 */
export function formatDecimal(value: Decimal): string {
  const integerText = value.abs().trunc().toFixed()
  // 小數部分連同小數點原樣接回去（沒有小數時是空字串）。
  const fractionText = value.abs().toFixed().slice(integerText.length)
  const groupedInteger = integerText.replace(/\B(?=(\d{3})+(?!\d))/g, ',')
  const sign = value.isNegative() && !value.isZero() ? '-' : ''

  return `${sign}${groupedInteger}${fractionText}`
}

/** 新台幣金額：`-800` → `-NT$800`。 */
export function formatTwd(value: Decimal): string {
  const sign = value.isNegative() && !value.isZero() ? '-' : ''

  return `${sign}NT$${formatDecimal(value.abs())}`
}

/** 帶正負號的新台幣金額：`1200` → `+NT$1,200`，零不帶號。 */
export function formatSignedTwd(value: Decimal): string {
  return value.isPositive() && !value.isZero() ? `+${formatTwd(value)}` : formatTwd(value)
}

/** 百分比，固定兩位小數（四捨五入）：`66.666…` → `66.67%`。 */
export function formatPercent(value: Decimal): string {
  return `${value.toDecimalPlaces(2, Decimal.ROUND_HALF_UP).toFixed(2)}%`
}
