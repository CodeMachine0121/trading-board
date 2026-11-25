import type { ISessionStorageProxy } from '~/domain/interface/i-session-storage-proxy'
import type { Session } from '~/domain/models/entities/session'

/**
 * Proxy：這個頁面手上那張短效通行憑證，**只放在記憶體**。
 *
 * 不寫進 localStorage / sessionStorage：任何能在這個網域跑腳本的人都讀得到那裡。
 * 重新整理頁面就沒了——由瀏覽器帶著的 HttpOnly 續期 cookie 換一張新的回來。
 */
export class MemorySessionStorageProxy implements ISessionStorageProxy {
  private session: Session | null = null

  readSession(): Session | null {
    return this.session
  }

  writeSession(session: Session): void {
    this.session = session
  }

  clearSession(): void {
    this.session = null
  }
}
