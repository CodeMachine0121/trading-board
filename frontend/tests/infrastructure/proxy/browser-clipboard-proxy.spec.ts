import { afterEach, describe, expect, it, vi } from 'vitest'
import { ClipboardUnavailableError } from '~/domain/errors/clipboard-unavailable-error'
import { BrowserClipboardProxy } from '~/infrastructure/proxy/browser-clipboard-proxy'

function stubClipboard(writeText: (text: string) => Promise<void>) {
  vi.stubGlobal('navigator', { clipboard: { writeText } })
}

afterEach(() => {
  vi.unstubAllGlobals()
})

describe('BrowserClipboardProxy', () => {
  it('把文字寫進剪貼簿', async () => {
    const writeText = vi.fn(async () => {})
    stubClipboard(writeText)

    await new BrowserClipboardProxy().copy('tbk_key')

    expect(writeText).toHaveBeenCalledWith('tbk_key')
  })

  it('瀏覽器不讓寫 → 無法自動複製', async () => {
    stubClipboard(async () => {
      throw new Error('denied')
    })

    await expect(new BrowserClipboardProxy().copy('tbk_key')).rejects.toBeInstanceOf(ClipboardUnavailableError)
  })
})
