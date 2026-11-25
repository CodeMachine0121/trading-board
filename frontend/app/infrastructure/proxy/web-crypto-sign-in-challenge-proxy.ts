import type { ISignInChallengeProxy } from '~/domain/interface/i-sign-in-challenge-proxy'
import { SignInChallenge } from '~/domain/models/entities/sign-in-challenge'

/** 32 bytes 的亂數：base64url 之後是 43 個字，正好落在 PKCE 規定的 43–128 之間。 */
const RANDOM_BYTES = 32

/** Proxy：用瀏覽器的 Web Crypto 抽密語與標記，並算出 S256 證明（RFC 7636）。 */
export class WebCryptoSignInChallengeProxy implements ISignInChallengeProxy {
  async createChallenge(): Promise<SignInChallenge> {
    const codeVerifier = this.randomText()
    const digest = await crypto.subtle.digest('SHA-256', new TextEncoder().encode(codeVerifier))

    return new SignInChallenge(codeVerifier, this.base64Url(new Uint8Array(digest)), this.randomText())
  }

  private randomText(): string {
    return this.base64Url(crypto.getRandomValues(new Uint8Array(RANDOM_BYTES)))
  }

  private base64Url(bytes: Uint8Array): string {
    return btoa(String.fromCharCode(...bytes)).replaceAll('+', '-').replaceAll('/', '_').replace(/=+$/, '')
  }
}
