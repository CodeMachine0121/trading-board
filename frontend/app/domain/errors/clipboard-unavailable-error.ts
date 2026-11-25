import { TradingBoardError } from '~/domain/errors/trading-board-error'

/** 瀏覽器不讓我們寫剪貼簿（權限、非安全連線）。金鑰還在畫面上，可以手動選取複製。 */
export class ClipboardUnavailableError extends TradingBoardError {
  constructor() {
    super('無法自動複製，請手動選取金鑰')
    this.name = 'ClipboardUnavailableError'
  }
}
