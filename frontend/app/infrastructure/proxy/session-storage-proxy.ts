import type { ISessionStorageProxy } from '~/domain/interface/i-session-storage-proxy'
import { Session } from '~/domain/models/entities/session'

const SESSION_STORAGE_KEY = 'trading-board:session'

/** 記在儲存裡的原始形狀，只存在於本檔內。 */
type StoredSessionWire = {
  accessToken?: unknown
  expiresAt?: unknown
}

/**
 * Proxy：唯一碰瀏覽器儲存的地方。
 *
 * 三個方法都**不拋**。瀏覽器把儲存關掉（無痕視窗、封鎖網站資料）時存取本身會拋，
 * 而那與「還沒登入過」對主人是同一件事；寫不進去也不該讓登入失敗——只是下次要重登。
 */
export class SessionStorageProxy implements ISessionStorageProxy {
  readSession(): Session | null {
    try {
      const storedValue = localStorage.getItem(SESSION_STORAGE_KEY)
      if (storedValue === null) {
        return null
      }

      const storedSession = JSON.parse(storedValue) as StoredSessionWire
      const expiresAt = typeof storedSession.expiresAt === 'string' ? new Date(storedSession.expiresAt) : null
      if (typeof storedSession.accessToken !== 'string' || expiresAt === null || Number.isNaN(expiresAt.getTime())) {
        return null
      }

      return new Session(storedSession.accessToken, expiresAt)
    }
    catch {
      return null
    }
  }

  writeSession(session: Session): void {
    try {
      localStorage.setItem(SESSION_STORAGE_KEY, JSON.stringify({
        accessToken: session.accessToken,
        expiresAt: session.expiresAt.toISOString(),
      }))
    }
    catch {
      // 記不住不影響這一次：畫面已經是登入狀態，只是下次打開要重登。
    }
  }

  clearSession(): void {
    try {
      localStorage.removeItem(SESSION_STORAGE_KEY)
    }
    catch {
      // 清不掉也不能讓登出失敗——它在畫面上一定要成功。
    }
  }
}
