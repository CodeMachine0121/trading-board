import type { PositionService } from '~/domain/service/position-service'
import type { ClosedPositionListDto } from '~/domain/models/dto/closed-position-list-dto'
import type { OpenPositionDto } from '~/domain/models/dto/open-position-dto'

/** Application：持股用例。純 TypeScript，全程只碰 DTO。 */
export class PositionApplication {
  constructor(private readonly positionService: PositionService) {}

  async listOpenPositions(): Promise<OpenPositionDto[]> {
    return this.positionService.listOpenPositions()
  }

  async listClosedPositions(month?: string): Promise<ClosedPositionListDto> {
    return this.positionService.listClosedPositions(month)
  }
}
