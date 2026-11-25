import type { DetailRowDto } from '~/domain/models/dto/detail-row-dto'
import type { Tone } from '~/domain/models/vo/tone'

/** 一檔已經出清的股票，以及它最後是獲利還是虧損。 */
export class ClosedPositionDto {
  constructor(
    readonly symbol: string,
    readonly title: string,
    readonly outcomeLabel: string,
    readonly tone: Tone,
    readonly realizedPnlText: string,
    readonly detailRows: readonly DetailRowDto[],
  ) {}
}
