import { TradingBoardError } from '~/domain/errors/trading-board-error'

/** 電子郵件或密碼對不上。板子刻意不說是哪一個錯，畫面也不說。 */
export class CredentialsRejectedError extends TradingBoardError {
  constructor() {
    super('電子郵件或密碼錯誤')
    this.name = 'CredentialsRejectedError'
  }
}
