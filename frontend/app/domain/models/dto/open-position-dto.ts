import type { DetailRowDto } from '~/domain/models/dto/detail-row-dto'

/** 一檔還持有的股票。 */
export class OpenPositionDto {
  constructor(
    readonly symbol: string,
    readonly title: string,
    readonly detailRows: readonly DetailRowDto[],
  ) {}
}
