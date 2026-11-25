import { TradingBoardError } from '~/domain/errors/trading-board-error'

const UNEXPECTED_FAILURE_MESSAGE = '發生未預期的錯誤，請重新整理'

/**
 * 任何一個失敗要對主人說什麼——全站只有這一條規則。
 *
 * 認得的錯誤（哨兵錯誤）照它自己的話說；認不得的說通用訊息，並記在主控台，
 * 不默默吞掉。每個畫面各寫一份的話，總有一份會忘了記主控台、或把技術訊息直接秀出來。
 */
export function failureMessageOf(error: unknown): string {
  if (error instanceof TradingBoardError) {
    return error.message
  }

  console.error(error)

  return UNEXPECTED_FAILURE_MESSAGE
}
