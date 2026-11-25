import type { IPendingSignInStorageProxy } from '~/domain/interface/i-pending-sign-in-storage-proxy'
import { PendingSignIn } from '~/domain/models/entities/pending-sign-in'

const PENDING_SIGN_IN_KEY = 'trading-board:pending-sign-in'

/** 記在儲存裡的原始形狀，只存在於本檔內。 */
type StoredPendingSignInWire = {
  codeVerifier?: unknown
  state?: unknown
}

/**
 * Proxy：跨過「送出帳密 → 板子轉回 /auth/callback」那次整頁跳轉，記住密語與標記。
 *
 * 用 sessionStorage：只屬於這個分頁、關掉就沒了，別的分頁拿不到。取出時立刻刪掉——一次登入只用一次。
 * 兩個方法都不拋：儲存被瀏覽器關掉時，回來就當作來回失效，請主人重新登入。
 */
export class BrowserPendingSignInStorageProxy implements IPendingSignInStorageProxy {
  save(pendingSignIn: PendingSignIn): void {
    try {
      sessionStorage.setItem(PENDING_SIGN_IN_KEY, JSON.stringify({
        codeVerifier: pendingSignIn.codeVerifier,
        state: pendingSignIn.state,
      }))
    }
    catch (error: unknown) {
      console.warn('could not remember the pending sign-in', error)
    }
  }

  take(): PendingSignIn | null {
    try {
      const storedValue = sessionStorage.getItem(PENDING_SIGN_IN_KEY)
      sessionStorage.removeItem(PENDING_SIGN_IN_KEY)
      if (storedValue === null) {
        return null
      }

      const stored = JSON.parse(storedValue) as StoredPendingSignInWire
      if (typeof stored.codeVerifier !== 'string' || typeof stored.state !== 'string') {
        return null
      }

      return new PendingSignIn(stored.codeVerifier, stored.state)
    }
    catch {
      return null
    }
  }
}
