import type { Session } from '~/domain/models/entities/session'

/** 剩不到這麼久就先換一張新的，免得帶著一張在路上過期的通行憑證去問板子。 */
const RENEWAL_MARGIN_MILLISECONDS = 60_000

/** 一段登入（頁面記憶體裡那張短效通行憑證）還能不能用、該不該先換新。 */
export class SessionDomain {
  constructor(private readonly session: Session) {}

  /** 剩下不到一分鐘（或已經失效）就該換新；剛好剩一分鐘還能用。 */
  needsRenewalAt(now: Date): boolean {
    return this.session.accessToken === ''
      || this.session.expiresAt.getTime() - now.getTime() < RENEWAL_MARGIN_MILLISECONDS
  }
}
