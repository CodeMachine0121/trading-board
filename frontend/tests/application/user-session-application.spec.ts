import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { UserSessionApplication } from '~/application/user-session-application'
import { BoardRejectedError } from '~/domain/errors/board-rejected-error'
import { BoardUnreachableError } from '~/domain/errors/board-unreachable-error'
import { CredentialsRejectedError } from '~/domain/errors/credentials-rejected-error'
import { SignInFlowExpiredError } from '~/domain/errors/sign-in-flow-expired-error'
import { SignInLockedError } from '~/domain/errors/sign-in-locked-error'
import type { IAuthProxy } from '~/domain/interface/i-auth-proxy'
import type { IPendingSignInStorageProxy } from '~/domain/interface/i-pending-sign-in-storage-proxy'
import type { ISessionStorageProxy } from '~/domain/interface/i-session-storage-proxy'
import type { ISignInChallengeProxy } from '~/domain/interface/i-sign-in-challenge-proxy'
import { CredentialsDto } from '~/domain/models/dto/credentials-dto'
import { SignInReturnDto } from '~/domain/models/dto/sign-in-return-dto'
import { Owner } from '~/domain/models/entities/owner'
import { PendingSignIn } from '~/domain/models/entities/pending-sign-in'
import { Session } from '~/domain/models/entities/session'
import { SignInChallenge } from '~/domain/models/entities/sign-in-challenge'
import { UserSessionService } from '~/domain/service/user-session-service'

const NOW = new Date('2026-11-25T12:00:00Z')

function sessionExpiringIn(milliseconds: number, token = 'current-token'): Session {
  return new Session(token, new Date(NOW.getTime() + milliseconds))
}

/** 替身：記憶體裡那段登入（vi.fn 組成，狀態放在閉包裡）。 */
function sessionStorageHolding(initial: Session | null) {
  let held = initial

  return {
    readSession: vi.fn<ISessionStorageProxy['readSession']>(() => held),
    writeSession: vi.fn<ISessionStorageProxy['writeSession']>((session) => {
      held = session
    }),
    clearSession: vi.fn<ISessionStorageProxy['clearSession']>(() => {
      held = null
    }),
  } satisfies ISessionStorageProxy
}

function setUp(options: { session?: Session | null, pending?: PendingSignIn | null } = {}) {
  const authProxy = {
    authorize: vi.fn<IAuthProxy['authorize']>(),
    exchangeCode: vi.fn<IAuthProxy['exchangeCode']>(),
    refresh: vi.fn<IAuthProxy['refresh']>(),
    revoke: vi.fn<IAuthProxy['revoke']>(),
    fetchCurrentOwner: vi.fn<IAuthProxy['fetchCurrentOwner']>(),
  } satisfies IAuthProxy
  const sessionStorageProxy = sessionStorageHolding(options.session ?? null)
  const pendingSignInStorageProxy = {
    save: vi.fn<IPendingSignInStorageProxy['save']>(),
    take: vi.fn<IPendingSignInStorageProxy['take']>().mockReturnValue(options.pending ?? null),
  } satisfies IPendingSignInStorageProxy
  const signInChallengeProxy = {
    createChallenge: vi.fn<ISignInChallengeProxy['createChallenge']>()
      .mockResolvedValue(new SignInChallenge('verifier', 'challenge', 'state-1')),
  } satisfies ISignInChallengeProxy
  const application = new UserSessionApplication(new UserSessionService(
    authProxy, sessionStorageProxy, pendingSignInStorageProxy, signInChallengeProxy))

  return { authProxy, sessionStorageProxy, pendingSignInStorageProxy, application }
}

beforeEach(() => {
  vi.useFakeTimers({ toFake: ['Date'] })
  vi.setSystemTime(NOW)
})

afterEach(() => {
  vi.useRealTimers()
  vi.restoreAllMocks()
})

describe('登入：電子郵件與密碼', () => {
  it('對了 → 記下密語與標記，回板子要轉去的網址', async () => {
    const { authProxy, pendingSignInStorageProxy, application } = setUp()
    authProxy.authorize.mockResolvedValue('https://web.example/auth/callback?code=c&state=state-1')

    const redirectTo = await application.beginSignIn(new CredentialsDto(' owner@example.com ', 'pw'))

    expect(redirectTo).toBe('https://web.example/auth/callback?code=c&state=state-1')
    expect(authProxy.authorize).toHaveBeenCalledWith('owner@example.com', 'pw', 'challenge', 'state-1')
    expect(pendingSignInStorageProxy.save).toHaveBeenCalledWith(new PendingSignIn('verifier', 'state-1'))
  })

  it.each([
    ['錯的密碼或不認得的電子郵件', new CredentialsRejectedError(), '電子郵件或密碼錯誤'],
    ['板子鎖住了登入', new SignInLockedError(), '嘗試次數過多，請 15 分鐘後再試'],
    ['連不上板子', new BoardUnreachableError(), '無法連線，請稍後再試'],
  ])('%s → 留在登入畫面並說明', async (_, failure, expectedMessage) => {
    const { authProxy, application } = setUp()
    authProxy.authorize.mockRejectedValue(failure)

    await expect(application.beginSignIn(new CredentialsDto('owner@example.com', 'x'))).rejects.toThrow(expectedMessage)
  })
})

describe('轉址回來：比對標記、換通行憑證', () => {
  it('標記相符、板子收下授權碼 → 這一頁登入了', async () => {
    const { authProxy, application } = setUp({ pending: new PendingSignIn('verifier', 'state-1') })
    authProxy.exchangeCode.mockResolvedValue(sessionExpiringIn(15 * 60_000, 'fresh-token'))

    await application.completeSignIn(new SignInReturnDto('the-code', 'state-1'))

    expect(authProxy.exchangeCode).toHaveBeenCalledWith('the-code', 'verifier')
    expect(await application.accessToken()).toBe('fresh-token')
    expect(authProxy.refresh).not.toHaveBeenCalled()
  })

  it.each([
    ['標記不同', new PendingSignIn('verifier', 'state-1'), new SignInReturnDto('the-code', 'someone-else')],
    ['這個分頁沒有發出過登入', null, new SignInReturnDto('the-code', 'state-1')],
    ['回來的網址沒有授權碼', new PendingSignIn('verifier', 'state-1'), new SignInReturnDto('', 'state-1')],
    ['記下的標記是空的', new PendingSignIn('verifier', ''), new SignInReturnDto('the-code', '')],
  ])('%s → 登入流程已失效，授權碼不拿去換', async (_, pending, signInReturn) => {
    const { authProxy, application } = setUp({ pending })

    await expect(application.completeSignIn(signInReturn)).rejects.toThrow(new SignInFlowExpiredError())
    expect(authProxy.exchangeCode).not.toHaveBeenCalled()
  })

  it('板子不收這張授權碼 → 登入流程已失效', async () => {
    const { authProxy, application } = setUp({ pending: new PendingSignIn('verifier', 'state-1') })
    authProxy.exchangeCode.mockRejectedValue(new BoardRejectedError('授權碼無效或已過期'))

    await expect(application.completeSignIn(new SignInReturnDto('the-code', 'state-1')))
      .rejects.toThrow('登入流程已失效，請重新登入')
  })

  it('認不得的失敗照原樣往上拋', async () => {
    const { authProxy, application } = setUp({ pending: new PendingSignIn('verifier', 'state-1') })
    authProxy.exchangeCode.mockRejectedValue(new TypeError('bug'))

    await expect(application.completeSignIn(new SignInReturnDto('the-code', 'state-1'))).rejects.toThrow('bug')
  })
})

describe('保持登入：續期', () => {
  it('打開頁面時瀏覽器帶著有效的續期憑證 → 不必看登入畫面', async () => {
    const { authProxy, application } = setUp()
    authProxy.refresh.mockResolvedValue(sessionExpiringIn(15 * 60_000, 'renewed'))

    expect(await application.restoreSession()).toBe(true)
    expect(await application.accessToken()).toBe('renewed')
  })

  it('沒有或被拒絕的續期憑證 → 要看登入畫面', async () => {
    const { authProxy, sessionStorageProxy, application } = setUp()
    authProxy.refresh.mockRejectedValue(new BoardRejectedError('登入已失效，請重新登入'))

    expect(await application.restoreSession()).toBe(false)
    expect(sessionStorageProxy.clearSession).toHaveBeenCalled()
  })

  it('通行憑證 30 秒後就到期 → 先換新再問板子', async () => {
    const { authProxy, application } = setUp({ session: sessionExpiringIn(30_000) })
    authProxy.refresh.mockResolvedValue(sessionExpiringIn(15 * 60_000, 'renewed'))

    expect(await application.accessToken()).toBe('renewed')
    expect(authProxy.refresh).toHaveBeenCalledTimes(1)
  })

  it('通行憑證還有 10 分鐘 → 直接用，不續期', async () => {
    const { authProxy, application } = setUp({ session: sessionExpiringIn(10 * 60_000) })

    expect(await application.accessToken()).toBe('current-token')
    expect(await application.restoreSession()).toBe(true)
    expect(authProxy.refresh).not.toHaveBeenCalled()
  })

  it('兩個畫面同時需要續期 → 只續一次，兩邊拿到同一張新的', async () => {
    const { authProxy, application } = setUp({ session: sessionExpiringIn(10_000) })
    authProxy.refresh.mockResolvedValue(sessionExpiringIn(15 * 60_000, 'renewed'))

    const tokens = await Promise.all([application.accessToken(), application.accessToken()])

    expect(tokens).toEqual(['renewed', 'renewed'])
    expect(authProxy.refresh).toHaveBeenCalledTimes(1)
  })

  it('一次續期結束後，下一次需要時會再續', async () => {
    const { authProxy, application } = setUp()
    authProxy.refresh
      .mockResolvedValueOnce(sessionExpiringIn(10_000, 'short-lived'))
      .mockResolvedValueOnce(sessionExpiringIn(15 * 60_000, 'renewed-again'))

    expect(await application.accessToken()).toBe('short-lived')
    expect(await application.accessToken()).toBe('renewed-again')
    expect(authProxy.refresh).toHaveBeenCalledTimes(2)
  })
})

describe('登出', () => {
  it('結束板子上的登入，這一頁也忘掉', async () => {
    const { authProxy, application } = setUp({ session: sessionExpiringIn(10 * 60_000) })
    authProxy.revoke.mockResolvedValue()
    authProxy.refresh.mockRejectedValue(new BoardRejectedError('登入已失效，請重新登入'))

    await application.signOut()

    expect(authProxy.revoke).toHaveBeenCalledTimes(1)
    expect(await application.restoreSession()).toBe(false)
  })

  it('板子連不上 → 畫面上照樣登出，不拋', async () => {
    const { authProxy, sessionStorageProxy, application } = setUp({ session: sessionExpiringIn(10 * 60_000) })
    authProxy.revoke.mockRejectedValue(new BoardUnreachableError())
    vi.spyOn(console, 'warn').mockImplementation(() => {})

    await expect(application.signOut()).resolves.toBeUndefined()
    expect(sessionStorageProxy.readSession()).toBeNull()
  })

  it('已經在別的分頁登出過 → 照樣回登入畫面，不拋', async () => {
    const { authProxy, application } = setUp()
    authProxy.revoke.mockResolvedValue()

    await expect(application.signOut()).resolves.toBeUndefined()
  })

  it('板子說登入不算數 → 只忘掉這一頁，不再去問板子', () => {
    const { authProxy, sessionStorageProxy, application } = setUp({ session: sessionExpiringIn(10 * 60_000) })

    application.forgetSession()

    expect(sessionStorageProxy.readSession()).toBeNull()
    expect(authProxy.revoke).not.toHaveBeenCalled()
  })
})

describe('我是誰', () => {
  it('回登入中的電子郵件', async () => {
    const { authProxy, application } = setUp()
    authProxy.fetchCurrentOwner.mockResolvedValue(new Owner('owner@example.com'))

    expect((await application.fetchCurrentOwner()).email).toBe('owner@example.com')
  })
})
