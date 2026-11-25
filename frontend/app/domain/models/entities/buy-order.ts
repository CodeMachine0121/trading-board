import type Decimal from 'decimal.js'
import type { Fill } from '~/domain/models/entities/fill'
import type { OrderStatus } from '~/domain/models/vo/order-status'
import type { OrderProgressVo } from '~/domain/models/vo/order-progress-vo'
import { OrderProgressDomain } from '~/domain/models/domains/order-progress-domain'
import { DetailRowDto } from '~/domain/models/dto/detail-row-dto'
import { OrderDto } from '~/domain/models/dto/order-dto'
import { formatDecimal, formatTwd } from '~/utilities/decimal-formatter'

/** 一筆買進委託（需求 §3.2）。 */
export class BuyOrder {
  constructor(
    readonly id: string,
    readonly symbol: string,
    readonly name: string,
    readonly limitPrice: Decimal,
    readonly maxChasePrice: Decimal,
    readonly shares: number,
    readonly estimatedAmount: Decimal,
    readonly stopLossPrice: Decimal,
    readonly takeProfitPrice: Decimal,
    readonly reason: string,
    readonly isMonthEndAllocation: boolean,
    readonly status: OrderStatus,
    readonly filledShares: number,
    readonly fills: readonly Fill[],
  ) {}

  toProgressDomain(sessionIsPast: boolean): OrderProgressDomain {
    return new OrderProgressDomain(this.status, this.shares, this.filledShares, sessionIsPast)
  }

  toDto(tradeDate: string, progress: OrderProgressVo, defaultFilledAt: string): OrderDto {
    return new OrderDto(
      this.id,
      tradeDate,
      'buy',
      `${this.name} ${this.symbol}`,
      this.isMonthEndAllocation ? ['月底配置'] : [],
      progress.statusLabel,
      progress.statusTone,
      `已成交 ${this.filledShares} / ${this.shares} 股`,
      [
        new DetailRowDto('限價', formatDecimal(this.limitPrice)),
        new DetailRowDto('最高追價', formatDecimal(this.maxChasePrice)),
        new DetailRowDto('股數', `${this.shares} 股`),
        new DetailRowDto('預估金額', formatTwd(this.estimatedAmount)),
        new DetailRowDto('停損', formatDecimal(this.stopLossPrice), 'loss'),
        new DetailRowDto('停利', formatDecimal(this.takeProfitPrice), 'profit'),
      ],
      this.reason,
      this.fills.map(fill => fill.toDto()),
      progress.actionable,
      progress.remainingShares,
      this.limitPrice.toFixed(),
      String(progress.remainingShares),
      defaultFilledAt,
    )
  }
}
