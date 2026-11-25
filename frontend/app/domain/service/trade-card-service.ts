import type { IOrderProxy } from '~/domain/interface/i-order-proxy'
import type { ITradeCardProxy } from '~/domain/interface/i-trade-card-proxy'
import type { FillReportDto } from '~/domain/models/dto/fill-report-dto'
import type { OrderActionDto } from '~/domain/models/dto/order-action-dto'
import type { TradeCardDto } from '~/domain/models/dto/trade-card-dto'
import type { TradeCard } from '~/domain/models/entities/trade-card'
import { FillDraftDomain } from '~/domain/models/domains/fill-draft-domain'
import { TaipeiCalendarDomain } from '~/domain/models/domains/taipei-calendar-domain'
import { MonthlyCardListDto } from '~/domain/models/dto/monthly-card-list-dto'

/**
 * 交易卡與它上面的委託：讀卡、列一個月的卡、回報成交、略過。
 *
 * 回報與略過之後**重新讀一次整張卡**，而不是只換掉那一筆委託：板子回的「更新後的委託」
 * 沒說是買進還是賣出，而重讀一次保證畫面上的每一筆都是板子現在的樣子。
 */
export class TradeCardService {
  constructor(
    private readonly tradeCardProxy: ITradeCardProxy,
    private readonly orderProxy: IOrderProxy,
  ) {}

  async fetchLatestCard(): Promise<TradeCardDto | null> {
    return this.toCardDto(await this.tradeCardProxy.fetchLatestCard())
  }

  async fetchCard(tradeDate: string): Promise<TradeCardDto | null> {
    return this.toCardDto(await this.tradeCardProxy.fetchCard(tradeDate))
  }

  async listCards(month?: string): Promise<MonthlyCardListDto> {
    const monthDto = new TaipeiCalendarDomain(new Date()).toMonthDto(month)
    const summaries = await this.tradeCardProxy.listCards(monthDto.value)

    return new MonthlyCardListDto(
      monthDto,
      [...summaries]
        .sort((left, right) => right.tradeDate.localeCompare(left.tradeDate))
        .map(summary => summary.toDto()),
    )
  }

  /** 先擋掉一定會被拒絕的表單內容（不送出），再請板子記下這筆成交。 */
  async reportFill(fillReportDto: FillReportDto): Promise<TradeCardDto | null> {
    const fillRequest = new FillDraftDomain(
      fillReportDto.priceText,
      fillReportDto.sharesText,
      fillReportDto.feeText,
      fillReportDto.filledAtText,
      fillReportDto.remainingShares,
    ).toRequestVo()

    await this.orderProxy.reportFill(fillReportDto.orderId, fillRequest)

    return this.toCardDto(await this.tradeCardProxy.fetchCard(fillReportDto.tradeDate))
  }

  async skipOrder(orderActionDto: OrderActionDto): Promise<TradeCardDto | null> {
    await this.orderProxy.skipOrder(orderActionDto.orderId)

    return this.toCardDto(await this.tradeCardProxy.fetchCard(orderActionDto.tradeDate))
  }

  private toCardDto(card: TradeCard | null): TradeCardDto | null {
    return card === null ? null : card.toDomain(new TaipeiCalendarDomain(new Date())).toDto()
  }
}
