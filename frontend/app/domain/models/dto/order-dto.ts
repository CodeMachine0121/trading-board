import type { DetailRowDto } from '~/domain/models/dto/detail-row-dto'
import type { FillDto } from '~/domain/models/dto/fill-dto'
import type { Tone } from '~/domain/models/vo/tone'

/**
 * 一筆委託（買進或賣出）在畫面上的全部樣子。
 *
 * 買進與賣出的欄位不同，但差異全部收在 `detailRows` 裡，所以同一個元件畫得出兩種。
 * `defaultFillPrice` / `defaultFillShares` / `defaultFilledAt` 是回報成交表單的初始值。
 */
export class OrderDto {
  constructor(
    readonly id: string,
    readonly tradeDate: string,
    readonly side: 'buy' | 'sell',
    readonly title: string,
    readonly tags: readonly string[],
    readonly statusLabel: string,
    readonly statusTone: Tone,
    readonly filledSharesText: string,
    readonly detailRows: readonly DetailRowDto[],
    readonly reason: string,
    readonly fills: readonly FillDto[],
    readonly actionable: boolean,
    readonly remainingShares: number,
    readonly defaultFillPrice: string,
    readonly defaultFillShares: string,
    readonly defaultFilledAt: string,
  ) {}
}
