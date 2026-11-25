import { PendingSignIn } from '~/domain/models/entities/pending-sign-in'

/**
 * 一次登入前先準備好的三樣東西（OAuth PKCE）：只有這個分頁知道的密語（code verifier）、
 * 由它推出、可以公開送出去的證明（S256 code challenge），以及來回比對用的標記（state）。
 */
export class SignInChallenge {
  constructor(
    readonly codeVerifier: string,
    readonly codeChallenge: string,
    readonly state: string,
  ) {}

  /** 跨過轉址要留下來的只有密語與標記；證明已經送出去了。 */
  toPendingSignIn(): PendingSignIn {
    return new PendingSignIn(this.codeVerifier, this.state)
  }
}
