import { TradingBoardError } from '~/domain/errors/trading-board-error'

/** 連續登入失敗太多次，板子暫時不讓人登入。 */
export class SignInLockedError extends TradingBoardError {
  constructor() {
    super('嘗試次數過多，請 15 分鐘後再試')
    this.name = 'SignInLockedError'
  }
}
