import type { DetailRowDto } from '~/domain/models/dto/detail-row-dto'
import type { Tone } from '~/domain/models/vo/tone'

/** 一個月的預算明細與已實現績效。 */
export class MonthSummaryDto {
  constructor(
    readonly modeLabel: string,
    readonly modeTone: Tone,
    readonly budgetRows: readonly DetailRowDto[],
    readonly realizedPnlText: string,
    readonly realizedPnlTone: Tone,
    readonly performanceRows: readonly DetailRowDto[],
  ) {}
}
