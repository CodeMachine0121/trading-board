import type { IPositionProxy } from '~/domain/interface/i-position-proxy'
import type { OpenPositionDto } from '~/domain/models/dto/open-position-dto'
import { TaipeiCalendarDomain } from '~/domain/models/domains/taipei-calendar-domain'
import { ClosedPositionListDto } from '~/domain/models/dto/closed-position-list-dto'

/** 持股：還持有的，以及某個月出清的。 */
export class PositionService {
  constructor(private readonly positionProxy: IPositionProxy) {}

  async listOpenPositions(): Promise<OpenPositionDto[]> {
    const positions = await this.positionProxy.listOpenPositions()

    return positions.map(position => position.toDomain().toOpenDto())
  }

  async listClosedPositions(month?: string): Promise<ClosedPositionListDto> {
    const monthDto = new TaipeiCalendarDomain(new Date()).toMonthDto(month)
    const positions = await this.positionProxy.listClosedPositions(monthDto.value)

    return new ClosedPositionListDto(monthDto, positions.map(position => position.toDomain().toClosedDto()))
  }
}
