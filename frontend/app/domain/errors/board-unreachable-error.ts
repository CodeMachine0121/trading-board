import { TradingBoardError } from '~/domain/errors/trading-board-error'

/** 連不上板子，或板子自己出了問題——不是使用者輸入的錯。 */
export class BoardUnreachableError extends TradingBoardError {
  constructor() {
    super('無法連線，請稍後再試')
    this.name = 'BoardUnreachableError'
  }
}
