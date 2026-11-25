import type { IAuthProxy } from '~/domain/interface/i-auth-proxy'
import type { IPendingSignInStorageProxy } from '~/domain/interface/i-pending-sign-in-storage-proxy'
import type { ISessionStorageProxy } from '~/domain/interface/i-session-storage-proxy'
import type { ISignInChallengeProxy } from '~/domain/interface/i-sign-in-challenge-proxy'
import type { CredentialsDto } from '~/domain/models/dto/credentials-dto'
import type { OwnerDto } from '~/domain/models/dto/owner-dto'
import type { SignInReturnDto } from '~/domain/models/dto/sign-in-return-dto'
import { SignInFlowExpiredError } from '~/domain/errors/sign-in-flow-expired-error'
import { TradingBoardError } from '~/domain/errors/trading-board-error'

/**
 * 主人的登入，從頭到尾：送出帳密、轉址回來換通行憑證、續期、登出。
 *
 * 每一發請求要的通行憑證都從 `accessToken()` 拿。快過期時它先續期，而且同一時間**只續一次**：
 * 兩發同時續期時，板子會把第二發看成「已被換掉的續期憑證又被拿來用」而撤銷整段登入。
 */
export class UserSessionService {
  /** 正在進行的那一次續期；大家都等它，不另開一次。 */
  private renewal: Promise<string | null> | null = null

  constructor(
    private readonly authProxy: IAuthProxy,
    private readonly sessionStorageProxy: ISessionStorageProxy,
    private readonly pendingSignInStorageProxy: IPendingSignInStorageProxy,
    private readonly signInChallengeProxy: ISignInChallengeProxy,
  ) {}

  /** 回板子要我們轉去的網址；帳密不對、被鎖住或連不上時，照板子的意思拋出哨兵錯誤。 */
  async beginSignIn(credentialsDto: CredentialsDto): Promise<string> {
    const challenge = await this.signInChallengeProxy.createChallenge()
    this.pendingSignInStorageProxy.save(challenge.toPendingSignIn())

    return this.authProxy.authorize(
      credentialsDto.email.trim(), credentialsDto.password, challenge.codeChallenge, challenge.state)
  }

  /** 標記不是這個分頁發出的、根本沒發過，或板子不收這張授權碼——一律 SignInFlowExpiredError。 */
  async completeSignIn(signInReturnDto: SignInReturnDto): Promise<void> {
    const pendingSignIn = this.pendingSignInStorageProxy.take()
    if (pendingSignIn === null || signInReturnDto.code === ''
      || !pendingSignIn.toDomain().isReturnOf(signInReturnDto.state)) {
      throw new SignInFlowExpiredError()
    }

    try {
      this.sessionStorageProxy.writeSession(
        await this.authProxy.exchangeCode(signInReturnDto.code, pendingSignIn.codeVerifier))
    }
    catch (error: unknown) {
      if (error instanceof TradingBoardError) {
        throw new SignInFlowExpiredError()
      }
      throw error
    }
  }

  /** 打開頁面時：手上沒有通行憑證就試著續期一次。續不到就是沒登入。 */
  async restoreSession(): Promise<boolean> {
    return (await this.usableAccessToken()) !== null
  }

  /** 給每一發請求用：剩不到一分鐘就先續期；續不到回 `null`（那一發就不帶憑證，板子會說沒登入）。 */
  async accessToken(): Promise<string | null> {
    return this.usableAccessToken()
  }

  /** 登出在畫面上一定成功：板子連不上也照樣忘掉這一頁的登入。 */
  async signOut(): Promise<void> {
    this.sessionStorageProxy.clearSession()
    try {
      await this.authProxy.revoke()
    }
    catch (error: unknown) {
      // 板子沒收到也只是那段續期憑證留到它自己過期；畫面上已經登出了。
      console.warn('sign-out could not reach the board', error)
    }
  }

  /** 被板子告知登入不算數時：只忘掉這一頁的登入，不必再去煩板子。 */
  forgetSession(): void {
    this.sessionStorageProxy.clearSession()
  }

  async fetchCurrentOwner(): Promise<OwnerDto> {
    return (await this.authProxy.fetchCurrentOwner()).toDto()
  }

  private async usableAccessToken(): Promise<string | null> {
    const session = this.sessionStorageProxy.readSession()
    if (session !== null && !session.toDomain().needsRenewalAt(new Date())) {
      return session.accessToken
    }

    this.renewal ??= this.authProxy.refresh()
      .then((renewed) => {
        this.sessionStorageProxy.writeSession(renewed)

        return renewed.accessToken
      })
      .catch(() => {
        // 續期被拒或連不上，對畫面都是「需要重新登入」；把這一頁的登入一併忘掉。
        this.sessionStorageProxy.clearSession()

        return null
      })
      .finally(() => {
        this.renewal = null
      })

    return this.renewal
  }
}
