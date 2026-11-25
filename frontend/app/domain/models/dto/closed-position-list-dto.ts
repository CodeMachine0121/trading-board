import type { ClosedPositionDto } from '~/domain/models/dto/closed-position-dto'
import type { MonthDto } from '~/domain/models/dto/month-dto'

/** 某個月出清的持股。 */
export class ClosedPositionListDto {
  constructor(
    readonly month: MonthDto,
    readonly positions: readonly ClosedPositionDto[],
  ) {}
}
