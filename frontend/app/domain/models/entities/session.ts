import { SessionDomain } from '~/domain/models/domains/session-domain'

/** 一段登入：板子簽給這台瀏覽器的通行憑證，以及它到什麼時候失效。 */
export class Session {
  constructor(
    readonly accessToken: string,
    readonly expiresAt: Date,
  ) {}

  toDomain(): SessionDomain {
    return new SessionDomain(this)
  }
}
