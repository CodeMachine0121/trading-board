import type { Position } from '~/domain/models/entities/position'
import { RealizedPnlDomain } from '~/domain/models/domains/realized-pnl-domain'
import { ClosedPositionDto } from '~/domain/models/dto/closed-position-dto'
import { DetailRowDto } from '~/domain/models/dto/detail-row-dto'
import { OpenPositionDto } from '~/domain/models/dto/open-position-dto'
import { formatDecimal, formatTwd } from '~/utilities/decimal-formatter'
import { formatCalendarDate } from '~/utilities/date-formatter'

/** 沒有停損停利時顯示的字。 */
const NOT_SET = '—'

/** 一檔持股要怎麼給主人看：還持有的看成本與防守價，出清的看賺賠。 */
export class PositionDomain {
  constructor(private readonly position: Position) {}

  toOpenDto(): OpenPositionDto {
    const { stopLossPrice, takeProfitPrice } = this.position

    return new OpenPositionDto(
      this.position.symbol,
      `${this.position.name} ${this.position.symbol}`,
      [
        new DetailRowDto('股數', `${this.position.shares} 股`),
        new DetailRowDto('平均成本', formatDecimal(this.position.averageCost)),
        new DetailRowDto('總成本', formatTwd(this.position.totalCost)),
        new DetailRowDto('停損', stopLossPrice === null ? NOT_SET : formatDecimal(stopLossPrice), 'loss'),
        new DetailRowDto('停利', takeProfitPrice === null ? NOT_SET : formatDecimal(takeProfitPrice), 'profit'),
        new DetailRowDto('進場日', formatCalendarDate(this.position.openedDate)),
        new DetailRowDto('持有天數', `${this.position.daysHeld} 天`),
      ],
    )
  }

  toClosedDto(): ClosedPositionDto {
    const realizedPnl = new RealizedPnlDomain(this.position.realizedPnl)
    const { closedDate } = this.position

    return new ClosedPositionDto(
      this.position.symbol,
      `${this.position.name} ${this.position.symbol}`,
      realizedPnl.outcomeLabel(),
      realizedPnl.tone(),
      realizedPnl.signedText(),
      [
        new DetailRowDto('進場日', formatCalendarDate(this.position.openedDate)),
        new DetailRowDto('出場日', closedDate === null ? NOT_SET : formatCalendarDate(closedDate)),
        new DetailRowDto('持有天數', `${this.position.daysHeld} 天`),
      ],
    )
  }
}
