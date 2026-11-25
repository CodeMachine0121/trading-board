import type Decimal from 'decimal.js'
import { FillDto } from '~/domain/models/dto/fill-dto'
import { formatDecimal, formatTwd } from '~/utilities/decimal-formatter'
import { formatTaipeiDateTime } from '~/utilities/date-formatter'

/** 一筆成交：板子記下來的價格、股數、手續費（含稅）、金額與時間。只增不改。 */
export class Fill {
  constructor(
    readonly id: string,
    readonly price: Decimal,
    readonly shares: number,
    readonly fee: Decimal,
    readonly amount: Decimal,
    readonly filledAt: Date,
  ) {}

  toDto(): FillDto {
    return new FillDto(
      this.id,
      formatDecimal(this.price),
      `${this.shares} 股`,
      formatTwd(this.fee),
      formatTwd(this.amount),
      formatTaipeiDateTime(this.filledAt),
    )
  }
}
