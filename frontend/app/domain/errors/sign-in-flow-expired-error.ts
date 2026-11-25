import { TradingBoardError } from '~/domain/errors/trading-board-error'

/**
 * 登入的來回沒能完成：回來的標記不是這個分頁發出去的、根本沒發出去過，或板子不收這張一次性授權碼。
 * 只能重新登入——重試同一張授權碼沒有意義（它只能用一次）。
 */
export class SignInFlowExpiredError extends TradingBoardError {
  constructor() {
    super('登入流程已失效，請重新登入')
    this.name = 'SignInFlowExpiredError'
  }
}
