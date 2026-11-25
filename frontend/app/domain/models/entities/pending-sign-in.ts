import { PendingSignInDomain } from '~/domain/models/domains/pending-sign-in-domain'

/** 一次還沒走完的登入：轉址回來時要用的密語與標記。 */
export class PendingSignIn {
  constructor(
    readonly codeVerifier: string,
    readonly state: string,
  ) {}

  toDomain(): PendingSignInDomain {
    return new PendingSignInDomain(this)
  }
}
