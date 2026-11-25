import type { TradeCardService } from '~/domain/service/trade-card-service'
import type { FillReportDto } from '~/domain/models/dto/fill-report-dto'
import type { MonthlyCardListDto } from '~/domain/models/dto/monthly-card-list-dto'
import type { OrderActionDto } from '~/domain/models/dto/order-action-dto'
import type { TradeCardDto } from '~/domain/models/dto/trade-card-dto'

/** Application：讀交易卡、回報成交、略過委託。純 TypeScript，全程只碰 DTO。 */
export class TradeCardApplication {
  constructor(private readonly tradeCardService: TradeCardService) {}

  async fetchLatestCard(): Promise<TradeCardDto | null> {
    return this.tradeCardService.fetchLatestCard()
  }

  async fetchCard(tradeDate: string): Promise<TradeCardDto | null> {
    return this.tradeCardService.fetchCard(tradeDate)
  }

  async listCards(month?: string): Promise<MonthlyCardListDto> {
    return this.tradeCardService.listCards(month)
  }

  async reportFill(fillReportDto: FillReportDto): Promise<TradeCardDto | null> {
    return this.tradeCardService.reportFill(fillReportDto)
  }

  async skipOrder(orderActionDto: OrderActionDto): Promise<TradeCardDto | null> {
    return this.tradeCardService.skipOrder(orderActionDto)
  }
}
