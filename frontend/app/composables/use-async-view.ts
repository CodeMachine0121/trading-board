import { ref, shallowRef } from 'vue'
import { TradingBoardError } from '~/domain/errors/trading-board-error'

const UNEXPECTED_FAILURE_MESSAGE = '發生未預期的錯誤，請重新整理'

/**
 * 一個畫面「載入 → 有資料／失敗 → 重試」的那一段狀態。每一頁都要，所以只寫一次。
 *
 * 認得的錯誤（哨兵錯誤）照它自己的話顯示；認不得的顯示通用訊息並記在主控台，
 * 不默默吞掉。
 */
export function useAsyncView<TData>(load: () => Promise<TData>) {
  const data = shallowRef<TData | null>(null)
  const loading = ref(false)
  const errorMessage = ref<string | null>(null)

  async function reload(): Promise<void> {
    loading.value = true
    errorMessage.value = null

    try {
      data.value = await load()
    }
    catch (error: unknown) {
      if (!(error instanceof TradingBoardError)) {
        console.error(error)
      }
      errorMessage.value = error instanceof TradingBoardError ? error.message : UNEXPECTED_FAILURE_MESSAGE
    }
    finally {
      loading.value = false
    }
  }

  return { data, loading, errorMessage, reload }
}
