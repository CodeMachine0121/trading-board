import type Decimal from 'decimal.js'
import type { Fill } from '~/domain/models/entities/fill'
import type { ExitReason } from '~/domain/models/vo/exit-reason'
import type { OrderStatus } from '~/domain/models/vo/order-status'
import type { OrderProgressVo } from '~/domain/models/vo/order-progress-vo'
import { OrderProgressDomain } from '~/domain/models/domains/order-progress-domain'
import { DetailRowDto } from '~/domain/models/dto/detail-row-dto'
import { OrderDto } from '~/domain/models/dto/order-dto'
import { formatDecimal } from '~/utilities/decimal-formatter'

const EXIT_REASON_LABELS: Record<ExitReason, string> = {
  takeProfit: '停利',
  stopLoss: '停損',
  trailingStop: '移動停利',
  timeStop: '時間停損',
}

/** 一筆賣出委託（需求 §3.2）。 */
export class SellOrder {
  constructor(
    readonly id: string,
    readonly symbol: string,
    readonly name: string,
    readonly limitPrice: Decimal,
    readonly shares: number,
    readonly exitReason: ExitReason,
    readonly reason: string,
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
      'sell',
      `${this.name} ${this.symbol}`,
      [],
      progress.statusLabel,
      progress.statusTone,
      `已成交 ${this.filledShares} / ${this.shares} 股`,
      [
        new DetailRowDto('限價', formatDecimal(this.limitPrice)),
        new DetailRowDto('股數', `${this.shares} 股`),
        new DetailRowDto('出場原因', EXIT_REASON_LABELS[this.exitReason]),
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
