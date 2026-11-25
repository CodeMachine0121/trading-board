import type { Session } from '~/domain/models/entities/session'

/** 一段登入還能不能用。 */
export class SessionDomain {
  constructor(private readonly session: Session) {}

  /** 失效的那一刻起就不能用了——不是失效之後才不能用。 */
  isUsableAt(now: Date): boolean {
    return this.session.accessToken !== '' && this.session.expiresAt.getTime() > now.getTime()
  }
}
