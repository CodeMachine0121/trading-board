import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { UserSessionApplication } from '~/application/user-session-application'
import { useSignInFlow } from '~/composables/use-sign-in-flow'
import { BoardRejectedError } from '~/domain/errors/board-rejected-error'
import { BoardUnreachableError } from '~/domain/errors/board-unreachable-error'
import { CredentialsRejectedError } from '~/domain/errors/credentials-rejected-error'
import { SignInLockedError } from '~/domain/errors/sign-in-locked-error'
import type { IAuthProxy } from '~/domain/interface/i-auth-proxy'
import type { IPendingSignInStorageProxy } from '~/domain/interface/i-pending-sign-in-storage-proxy'
import type { ISessionStorageProxy } from '~/domain/interface/i-session-storage-proxy'
import type { ISignInChallengeProxy } from '~/domain/interface/i-sign-in-challenge-proxy'
import { CredentialsDto } from '~/domain/models/dto/credentials-dto'
import { SignInReturnDto } from '~/domain/models/dto/sign-in-return-dto'
import { PendingSignIn } from '~/domain/models/entities/pending-sign-in'
import { Session } from '~/domain/models/entities/session'
import { SignInChallenge } from '~/domain/models/entities/sign-in-challenge'
import { UserSessionService } from '~/domain/service/user-session-service'

const NOW = new Date('2026-11-25T12:00:00Z')

function setUp(options: { session?: Session | null, pending?: PendingSignIn | null } = {}) {
  let held = options.session ?? null
  const authProxy = {
    authorize: vi.fn<IAuthProxy['authorize']>(),
    exchangeCode: vi.fn<IAuthProxy['exchangeCode']>(),
    refresh: vi.fn<IAuthProxy['refresh']>(),
    revoke: vi.fn<IAuthProxy['revoke']>(),
    fetchCurrentOwner: vi.fn<IAuthProxy['fetchCurrentOwner']>(),
  } satisfies IAuthProxy
  const sessionStorageProxy = {
    readSession: vi.fn<ISessionStorageProxy['readSession']>(() => held),
    writeSession: vi.fn<ISessionStorageProxy['writeSession']>((session) => {
      held = session
    }),
    clearSession: vi.fn<ISessionStorageProxy['clearSession']>(() => {
      held = null
    }),
  } satisfies ISessionStorageProxy
  const pendingSignInStorageProxy = {
    save: vi.fn<IPendingSignInStorageProxy['save']>(),
    take: vi.fn<IPendingSignInStorageProxy['take']>().mockReturnValue(options.pending ?? null),
  } satisfies IPendingSignInStorageProxy
  const signInChallengeProxy = {
    createChallenge: vi.fn<ISignInChallengeProxy['createChallenge']>()
      .mockResolvedValue(new SignInChallenge('verifier', 'challenge', 'state-1')),
  } satisfies ISignInChallengeProxy
  const navigate = vi.fn()
  const flow = useSignInFlow(new UserSessionApplication(new UserSessionService(
    authProxy, sessionStorageProxy, pendingSignInStorageProxy, signInChallengeProxy)), navigate)

  return { authProxy, navigate, flow, isSignedInHere: () => held !== null }
}

const SIGNED_IN = new Session('token', new Date(NOW.getTime() + 10 * 60_000))

beforeEach(() => {
  vi.useFakeTimers({ toFake: ['Date'] })
  vi.setSystemTime(NOW)
})

afterEach(() => {
  vi.useRealTimers()
  vi.restoreAllMocks()
})

describe('登入畫面', () => {
  it('帳密對了 → 整頁跳去板子回的網址', async () => {
    const { authProxy, navigate, flow } = setUp()
    authProxy.authorize.mockResolvedValue('https://web.example/auth/callback?code=c&state=state-1')

    await flow.signIn(new CredentialsDto('owner@example.com', 'pw'))

    expect(navigate).toHaveBeenCalledWith('https://web.example/auth/callback?code=c&state=state-1', { external: true })
    expect(flow.errorMessage.value).toBeNull()
  })

  it.each([
    [new CredentialsRejectedError(), '電子郵件或密碼錯誤'],
    [new SignInLockedError(), '嘗試次數過多，請 15 分鐘後再試'],
    [new BoardUnreachableError(), '無法連線，請稍後再試'],
  ])('被拒 → 留在登入畫面，顯示「%s」', async (failure, expectedMessage) => {
    const { authProxy, navigate, flow } = setUp()
    authProxy.authorize.mockRejectedValue(failure)

    await flow.signIn(new CredentialsDto('owner@example.com', 'x'))

    expect(navigate).not.toHaveBeenCalled()
    expect(flow.errorMessage.value).toBe(expectedMessage)
    expect(flow.submitting.value).toBe(false)
  })
})

describe('轉址回來的那一頁', () => {
  it('標記相符、授權碼被收下 → 進首頁', async () => {
    const { authProxy, navigate, flow, isSignedInHere } = setUp({ pending: new PendingSignIn('verifier', 'state-1') })
    authProxy.exchangeCode.mockResolvedValue(SIGNED_IN)

    await flow.completeSignIn(new SignInReturnDto('code', 'state-1'))

    expect(navigate).toHaveBeenCalledWith('/', { replace: true })
    expect(isSignedInHere()).toBe(true)
  })

  it.each([
    ['標記不同', new PendingSignIn('verifier', 'state-1'), new SignInReturnDto('code', 'other')],
    ['沒發出過登入', null, new SignInReturnDto('code', 'state-1')],
  ])('%s → 留在這頁說「登入流程已失效，請重新登入」', async (_, pending, signInReturn) => {
    const { navigate, flow } = setUp({ pending })

    await flow.completeSignIn(signInReturn)

    expect(navigate).not.toHaveBeenCalled()
    expect(flow.errorMessage.value).toBe('登入流程已失效，請重新登入')
  })

  it('板子不收授權碼 → 同樣說明', async () => {
    const { authProxy, navigate, flow } = setUp({ pending: new PendingSignIn('verifier', 'state-1') })
    authProxy.exchangeCode.mockRejectedValue(new BoardRejectedError('授權碼無效或已過期'))

    await flow.completeSignIn(new SignInReturnDto('code', 'state-1'))

    expect(navigate).not.toHaveBeenCalled()
    expect(flow.errorMessage.value).toBe('登入流程已失效，請重新登入')
  })
})

describe('每次換頁前的把關', () => {
  it.each([
    ['續期成功、要看首頁 → 照走', true, '/', undefined],
    ['續期成功、要看登入畫面 → 改去首頁', true, '/login', '/'],
    ['沒有或被拒絕的續期憑證、要看持股 → 改去登入畫面', false, '/positions', '/login'],
    ['沒登入、要看登入畫面 → 照走', false, '/login', undefined],
  ])('%s', async (_, renewalSucceeds, path, expected) => {
    const { authProxy, flow } = setUp()
    if (renewalSucceeds) {
      authProxy.refresh.mockResolvedValue(SIGNED_IN)
    }
    else {
      authProxy.refresh.mockRejectedValue(new BoardRejectedError('登入已失效，請重新登入'))
    }

    expect(await flow.guard(path)).toBe(expected)
  })

  it('轉址回來的那一頁不必登入，也不去續期', async () => {
    const { authProxy, flow } = setUp()

    expect(await flow.guard('/auth/callback')).toBeUndefined()
    expect(authProxy.refresh).not.toHaveBeenCalled()
  })
})

describe('登出', () => {
  it.each([
    ['板子收到', () => Promise.resolve()],
    ['板子連不上', () => Promise.reject(new BoardUnreachableError())],
  ])('%s → 這一頁忘掉登入，回登入畫面', async (_, revoke) => {
    const { authProxy, navigate, flow, isSignedInHere } = setUp({ session: SIGNED_IN })
    authProxy.revoke.mockImplementation(revoke)
    vi.spyOn(console, 'warn').mockImplementation(() => {})

    await flow.signOut()

    expect(isSignedInHere()).toBe(false)
    expect(navigate).toHaveBeenCalledWith('/login')
  })

  it('已經在別的分頁登出過 → 照樣回登入畫面', async () => {
    const { authProxy, navigate, flow } = setUp()
    authProxy.revoke.mockResolvedValue()

    await expect(flow.signOut()).resolves.toBeUndefined()
    expect(navigate).toHaveBeenCalledWith('/login')
  })
})
