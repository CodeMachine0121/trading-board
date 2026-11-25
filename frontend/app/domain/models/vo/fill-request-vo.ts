import type Decimal from 'decimal.js'

/**
 * 一筆驗證過、可以送給板子的成交回報。
 * `fee` 是 `undefined` 代表沒填——由板子估算，不是零。
 */
export class FillRequestVo {
  constructor(
    readonly price: Decimal,
    readonly shares: number,
    readonly fee: Decimal | undefined,
    readonly filledAt: Date,
  ) {}
}
