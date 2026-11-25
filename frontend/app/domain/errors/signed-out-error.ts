import { TradingBoardError } from '~/domain/errors/trading-board-error'

/** 這一段登入已經不算數了（過期或被板子拒絕），得重新登入。 */
export class SignedOutError extends TradingBoardError {
  constructor() {
    super('登入已失效，請重新登入')
    this.name = 'SignedOutError'
  }
}
