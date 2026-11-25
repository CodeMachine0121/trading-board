import { afterEach, describe, expect, it, vi } from 'vitest'
import { CredentialsRejectedError } from '~/domain/errors/credentials-rejected-error'
import type { ISessionStorageProxy } from '~/domain/interface/i-session-storage-proxy'
import { Session } from '~/domain/models/entities/session'
import { AuthProxy } from '~/infrastructure/proxy/auth-proxy'

const BASE_URL = 'https://board.example/api/v1'

function sessionStorageWith(session: Session | null): ISessionStorageProxy {
  return {
    readSession: vi.fn<ISessionStorageProxy['readSession']>().mockReturnValue(session),
    writeSession: vi.fn<ISessionStorageProxy['writeSession']>(),
    clearSession: vi.fn<ISessionStorageProxy['clearSession']>(),
  }
}

/** 照 ofetch 的 FetchError 長相做一個失敗：狀態碼與契約的錯誤內容。 */
function boardFailure(status: number, message?: string) {
  return Object.assign(new Error('fetch failed'), {
    response: { status },
    data: message === undefined ? undefined : { error: { code: 'any', message } },
  })
}

function stubFetch() {
  const fetchMock = vi.fn()
  vi.stubGlobal('$fetch', fetchMock)

  return fetchMock
}

afterEach(() => {
  vi.unstubAllGlobals()
})

describe('AuthProxy', () => {
  it('帳密正確 → 換到一段登入（憑證與到期時刻），請求不帶舊憑證', async () => {
    const fetchMock = stubFetch().mockResolvedValue({ accessToken: 'new-token', expiresAt: '2026-11-26T00:00:00+08:00' })

    const session = await new AuthProxy(BASE_URL, sessionStorageWith(new Session('old', new Date()))).signIn('owner', 'pw')

    expect(fetchMock).toHaveBeenCalledWith(`${BASE_URL}/auth/login`, {
      method: 'POST',
      query: undefined,
      body: { username: 'owner', password: 'pw' },
      headers: {},
    })
    expect(session.accessToken).toBe('new-token')
    expect(session.expiresAt.toISOString()).toBe('2026-11-25T16:00:00.000Z')
  })

  it('帳密錯誤（401）→ 帳號或密碼錯誤，而不是「被登出」', async () => {
    stubFetch().mockRejectedValue(boardFailure(401, 'invalid'))
    const onSignedOut = vi.fn()

    await expect(new AuthProxy(BASE_URL, sessionStorageWith(null), onSignedOut).signIn('owner', 'wrong'))
      .rejects.toBeInstanceOf(CredentialsRejectedError)
    expect(onSignedOut).not.toHaveBeenCalled()
  })
})
