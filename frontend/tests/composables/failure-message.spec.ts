import { describe, expect, it, vi } from 'vitest'
import { failureMessageOf } from '~/composables/failure-message'
import { BoardRejectedError } from '~/domain/errors/board-rejected-error'
import { CredentialsRejectedError } from '~/domain/errors/credentials-rejected-error'

describe('failureMessageOf：一個失敗要對主人說什麼', () => {
  it.each([
    ['帳密錯誤 → 照它的話', new CredentialsRejectedError(), '電子郵件或密碼錯誤'],
    ['板子拒絕 → 照板子的話', new BoardRejectedError('超出本月剩餘預算'), '超出本月剩餘預算'],
  ])('%s', (_, error, message) => {
    expect(failureMessageOf(error)).toBe(message)
  })

  it('認不得的錯誤 → 通用訊息，並記在主控台', () => {
    const consoleError = vi.spyOn(console, 'error').mockImplementation(() => {})
    const error = new Error('boom')

    expect(failureMessageOf(error)).toBe('發生未預期的錯誤，請重新整理')
    expect(consoleError).toHaveBeenCalledWith(error)
    consoleError.mockRestore()
  })
})
