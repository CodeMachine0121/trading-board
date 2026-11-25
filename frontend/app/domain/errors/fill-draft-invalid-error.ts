import { TradingBoardError } from '~/domain/errors/trading-board-error'

/** 回報成交的表單填得不對；還沒送出去。 */
export class FillDraftInvalidError extends TradingBoardError {
  constructor(message: string) {
    super(message)
    this.name = 'FillDraftInvalidError'
  }
}
