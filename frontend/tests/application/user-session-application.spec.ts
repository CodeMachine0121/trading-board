import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { UserSessionApplication } from '~/application/user-session-application'
import { CredentialsRejectedError } from '~/domain/errors/credentials-rejected-error'
import type { IAuthProxy } from '~/domain/interface/i-auth-proxy'
import type { ISessionStorageProxy } from '~/domain/interface/i-session-storage-proxy'
import { CredentialsDto } from '~/domain/models/dto/credentials-dto'
import { Session } from '~/domain/models/entities/session'
import { UserSessionService } from '~/domain/service/user-session-service'

const NOW = new Date('2026-11-25T12:00:00Z')

function setUp(storedSession: Session | null = null) {
  const authProxy = { signIn: vi.fn<IAuthProxy['signIn']>() } satisfies IAuthProxy
  const sessionStorageProxy = {
    readSession: vi.fn<ISessionStorageProxy['readSession']>().mockReturnValue(storedSession),
    writeSession: vi.fn<ISessionStorageProxy['writeSession']>(),
    clearSession: vi.fn<ISessionStorageProxy['clearSession']>(),
  } satisfies ISessionStorageProxy
  const application = new UserSessionApplication(new UserSessionService(authProxy, sessionStorageProxy))

  return { authProxy, sessionStorageProxy, application }
}

beforeEach(() => {
  vi.useFakeTimers({ toFake: ['Date'] })
  vi.setSystemTime(NOW)
})

afterEach(() => {
  vi.useRealTimers()
})

describe('UserSessionApplication', () => {
  it('帳密正確 → 換到的登入記在這台瀏覽器', async () => {
    const { authProxy, sessionStorageProxy, application } = setUp()
    const session = new Session('token', new Date('2026-11-25T13:00:00Z'))
    authProxy.signIn.mockResolvedValue(session)

    await application.signIn(new CredentialsDto('owner', 'secret'))

    expect(authProxy.signIn).toHaveBeenCalledWith('owner', 'secret')
    expect(sessionStorageProxy.writeSession).toHaveBeenCalledWith(session)
  })

  it('帳密錯誤 → 帳號或密碼錯誤，什麼都不記', async () => {
    const { authProxy, sessionStorageProxy, application } = setUp()
    authProxy.signIn.mockRejectedValue(new CredentialsRejectedError())

    await expect(application.signIn(new CredentialsDto('owner', 'wrong'))).rejects.toThrow('帳號或密碼錯誤')
    expect(sessionStorageProxy.writeSession).not.toHaveBeenCalled()
  })

  it.each([
    ['沒有記著任何登入 → 未登入', null, false],
    ['記著的登入已過期 → 未登入', new Session('token', new Date('2026-11-25T11:00:00Z')), false],
    ['記著的登入還沒到期 → 已登入', new Session('token', new Date('2026-11-25T13:00:00Z')), true],
  ])('%s', (_, storedSession, signedIn) => {
    expect(setUp(storedSession).application.isSignedIn()).toBe(signedIn)
  })

  it('登出 → 清掉這台瀏覽器記著的登入', () => {
    const { sessionStorageProxy, application } = setUp()

    application.signOut()

    expect(sessionStorageProxy.clearSession).toHaveBeenCalledTimes(1)
  })
})
