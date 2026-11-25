import type { Owner } from '~/domain/models/entities/owner'
import type { Session } from '~/domain/models/entities/session'

/**
 * 板子自己的登入（OAuth 2.0 授權碼 + PKCE）。
 *
 * 長效的續期憑證只在瀏覽器的 HttpOnly cookie 裡，這裡永遠看不到它：
 * 換授權碼與續期都只拿回短效通行憑證。
 */
export interface IAuthProxy {
  /** 帳密與授權參數送去板子；成功時回板子要我們轉去的網址（帶授權碼與標記）。 */
  authorize(
    email: string, password: string, codeChallenge: string, state: string,
  ): Promise<string>
  exchangeCode(code: string, codeVerifier: string): Promise<Session>
  /** 用瀏覽器帶著的續期憑證換一張新的短效通行憑證（續期憑證同時被換新）。 */
  refresh(): Promise<Session>
  /** 結束這一次登入（撤銷續期憑證）。 */
  revoke(): Promise<void>
  fetchCurrentOwner(): Promise<Owner>
}
