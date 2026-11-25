import { TradingBoardError } from '~/domain/errors/trading-board-error'

/**
 * 板子依業務規則拒絕了這次操作（驗證失敗、衝突、超出預算…）。
 * 訊息是板子自己的說法，照實顯示給使用者。
 */
export class BoardRejectedError extends TradingBoardError {
  constructor(message: string) {
    super(message)
    this.name = 'BoardRejectedError'
  }
}
