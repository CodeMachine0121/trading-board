import type { PendingSignIn } from '~/domain/models/entities/pending-sign-in'

/** 轉址回來的標記是不是這個分頁發出去的那一個。不是的話，那張授權碼不能拿來換。 */
export class PendingSignInDomain {
  constructor(private readonly pendingSignIn: PendingSignIn) {}

  isReturnOf(state: string): boolean {
    return this.pendingSignIn.state !== '' && this.pendingSignIn.state === state
  }
}
