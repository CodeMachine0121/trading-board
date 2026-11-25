import type { Session } from '~/domain/models/entities/session'

/** 這個頁面手上的那一段登入。只放在記憶體：關掉或重新整理頁面就沒了，靠續期找回來。 */
export interface ISessionStorageProxy {
  readSession(): Session | null
  writeSession(session: Session): void
  clearSession(): void
}
