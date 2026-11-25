import type { OrderStatus } from '~/domain/models/vo/order-status'
import { OrderProgressVo } from '~/domain/models/vo/order-progress-vo'

/**
 * 一筆委託走到哪裡（需求 §3.3）。
 *
 * 狀態由「成交了幾股 vs 委託幾股」推出來，只有「略過」是主人明說的。
 * 「已過期」只是畫面上的說法：交易日已經過了、而且一股都還沒成交——
 * 已經成交一部分的委託不算過期，剩下的仍然可以補回報。
 */
export class OrderProgressDomain {
  private readonly filledShares: number

  constructor(
    private readonly status: OrderStatus,
    private readonly shares: number,
    filledShares: number,
    private readonly sessionIsPast: boolean,
  ) {
    this.filledShares = Math.min(Math.max(filledShares, 0), shares)
  }

  toVo(): OrderProgressVo {
    const remainingShares = this.shares - this.filledShares

    if (this.status === 'skipped') {
      return new OrderProgressVo('已略過', 'neutral', remainingShares, false)
    }
    if (remainingShares === 0) {
      return new OrderProgressVo('已成交', 'primary', 0, false)
    }
    if (this.filledShares > 0) {
      return new OrderProgressVo('部分成交', 'warning', remainingShares, true)
    }
    if (this.sessionIsPast) {
      return new OrderProgressVo('已過期', 'neutral', remainingShares, false)
    }

    return new OrderProgressVo('待成交', 'info', remainingShares, true)
  }
}
