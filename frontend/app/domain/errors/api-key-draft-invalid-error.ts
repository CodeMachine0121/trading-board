import { TradingBoardError } from '~/domain/errors/trading-board-error'

/** 建立金鑰的表單填得不對；還沒送出去。 */
export class ApiKeyDraftInvalidError extends TradingBoardError {
  constructor(message: string) {
    super(message)
    this.name = 'ApiKeyDraftInvalidError'
  }
}
