import { describe, expect, it, vi } from 'vitest'
import { useAsyncView } from '~/composables/use-async-view'
import { BoardUnreachableError } from '~/domain/errors/board-unreachable-error'

describe('useAsyncView', () => {
  it('載入成功 → 有資料、沒有錯誤、不在載入中', async () => {
    const view = useAsyncView(async () => 42)

    await view.reload()

    expect([view.data.value, view.errorMessage.value, view.loading.value]).toEqual([42, null, false])
  })

  it('認得的錯誤 → 照它的話顯示', async () => {
    const view = useAsyncView(async () => {
      throw new BoardUnreachableError()
    })

    await view.reload()

    expect(view.errorMessage.value).toBe('無法連線，請稍後再試')
  })

  it('認不得的錯誤 → 通用訊息，並記在主控台（不默默吞掉）', async () => {
    const consoleError = vi.spyOn(console, 'error').mockImplementation(() => {})
    const view = useAsyncView(async () => {
      throw new Error('boom')
    })

    await view.reload()

    expect(view.errorMessage.value).toBe('發生未預期的錯誤，請重新整理')
    expect(consoleError).toHaveBeenCalled()
    consoleError.mockRestore()
  })

  it('重試成功 → 清掉上一次的錯誤', async () => {
    const load = vi.fn<() => Promise<number>>()
      .mockRejectedValueOnce(new BoardUnreachableError())
      .mockResolvedValueOnce(7)
    const view = useAsyncView(load)

    await view.reload()
    await view.reload()

    expect([view.data.value, view.errorMessage.value]).toEqual([7, null])
  })
})
