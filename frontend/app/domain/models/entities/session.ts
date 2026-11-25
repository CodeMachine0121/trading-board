import { SessionDomain } from '~/domain/models/domains/session-domain'

/** 一段登入：板子發給這個頁面的短效通行憑證，以及它到什麼時候失效。只放在頁面記憶體裡。 */
export class Session {
  constructor(
    readonly accessToken: string,
    readonly expiresAt: Date,
  ) {}

  toDomain(): SessionDomain {
    return new SessionDomain(this)
  }
}
