import type { Session } from '~/domain/models/entities/session'

/** 這台瀏覽器記著的那一段登入。三個方法都不拋：記不住與沒登入過，對主人是同一件事。 */
export interface ISessionStorageProxy {
  readSession(): Session | null
  writeSession(session: Session): void
  clearSession(): void
}
