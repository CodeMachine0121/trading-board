import type { Tone } from '~/domain/models/vo/tone'

/** 月份清單裡的一列交易卡。 */
export class TradeCardSummaryDto {
  constructor(
    readonly tradeDate: string,
    readonly tradeDateLabel: string,
    readonly statusLabel: string,
    readonly statusTone: Tone,
    readonly orderCountText: string,
    readonly filledOrderCountText: string,
  ) {}
}
