import type { IClipboardProxy } from '~/domain/interface/i-clipboard-proxy'
import { ClipboardUnavailableError } from '~/domain/errors/clipboard-unavailable-error'

/** Proxy：瀏覽器的剪貼簿。 */
export class BrowserClipboardProxy implements IClipboardProxy {
  async copy(text: string): Promise<void> {
    try {
      await navigator.clipboard.writeText(text)
    }
    catch {
      throw new ClipboardUnavailableError()
    }
  }
}
