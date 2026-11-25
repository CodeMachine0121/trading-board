import { ref, shallowRef } from 'vue'
import { failureMessageOf } from '~/composables/failure-message'

/**
 * 一個畫面「載入 → 有資料／失敗 → 重試」的那一段狀態。每一頁都要，所以只寫一次。
 *
 * 失敗要說什麼由 failureMessageOf 決定。
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
      errorMessage.value = failureMessageOf(error)
    }
    finally {
      loading.value = false
    }
  }

  return { data, loading, errorMessage, reload }
}
