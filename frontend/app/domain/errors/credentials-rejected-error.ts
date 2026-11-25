import { TradingBoardError } from '~/domain/errors/trading-board-error'

/** 帳號或密碼對不上。 */
export class CredentialsRejectedError extends TradingBoardError {
  constructor() {
    super('帳號或密碼錯誤')
    this.name = 'CredentialsRejectedError'
  }
}
