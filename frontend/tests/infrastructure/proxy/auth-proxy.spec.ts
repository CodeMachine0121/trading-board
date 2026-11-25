import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { BoardRejectedError } from '~/domain/errors/board-rejected-error'
import { CredentialsRejectedError } from '~/domain/errors/credentials-rejected-error'
import { SignInLockedError } from '~/domain/errors/sign-in-locked-error'
import { OAuthClientVo } from '~/domain/models/vo/oauth-client-vo'
import { AuthProxy } from '~/infrastructure/proxy/auth-proxy'

const BASE_URL = 'https://board.example/api/v1'
const CLIENT = new OAuthClientVo('trading-board-web', 'https://web.example/auth/callback')
const NOW = new Date('2026-11-25T12:00:00Z')

function stubFetch() {
  const fetchMock = vi.fn()
  vi.stubGlobal('$fetch', fetchMock)

  return fetchMock
}

function oauthFailure(status: number, error: string, description: string) {
  return Object.assign(new Error('fetch failed'), {
    response: { status },
    data: { error, error_description: description },
  })
}

function newProxy(accessToken: string | null = null) {
  return new AuthProxy(BASE_URL, async () => accessToken, vi.fn(), CLIENT)
}

beforeEach(() => {
  vi.useFakeTimers({ toFake: ['Date'] })
  vi.setSystemTime(NOW)
})

afterEach(() => {
  vi.useRealTimers()
  vi.unstubAllGlobals()
})

describe('AuthProxy：板子自己的 OAuth 登入', () => {
  it('authorize：帳密與 PKCE 參數送去 /oauth/authorize，不帶通行憑證，回板子要轉去的網址', async () => {
    const fetchMock = stubFetch().mockResolvedValue({ redirectTo: 'https://web.example/auth/callback?code=c&state=s' })

    const redirectTo = await newProxy('should-not-be-sent').authorize('owner@example.com', 'pw', 'challenge', 's')

    expect(redirectTo).toBe('https://web.example/auth/callback?code=c&state=s')
    expect(fetchMock).toHaveBeenCalledWith(`${BASE_URL}/oauth/authorize`, {
      method: 'POST',
      query: undefined,
      body: {
        email: 'owner@example.com',
        password: 'pw',
        response_type: 'code',
        client_id: 'trading-board-web',
        redirect_uri: 'https://web.example/auth/callback',
        code_challenge: 'challenge',
        code_challenge_method: 'S256',
        state: 's',
      },
      headers: {},
    })
  })

  it.each([
    ['401 → 電子郵件或密碼錯誤', oauthFailure(401, 'access_denied', '電子郵件或密碼錯誤'), CredentialsRejectedError],
    ['429 → 嘗試次數過多', oauthFailure(429, 'temporarily_unavailable', '嘗試次數過多，請 15 分鐘後再試'), SignInLockedError],
  ])('authorize 被拒：%s', async (_, failure, expectedError) => {
    stubFetch().mockRejectedValue(failure)

    await expect(newProxy().authorize('owner@example.com', 'x', 'c', 's')).rejects.toBeInstanceOf(expectedError)
  })

  it('換授權碼：表單編碼、讓瀏覽器帶上並收下 cookie；回一張 expires_in 秒後到期的通行憑證', async () => {
    const fetchMock = stubFetch().mockResolvedValue({ access_token: 'at', token_type: 'Bearer', expires_in: 900 })

    const session = await newProxy().exchangeCode('the-code', 'the-verifier')

    expect(session.accessToken).toBe('at')
    expect(session.expiresAt.toISOString()).toBe('2026-11-25T12:15:00.000Z')
    const [url, options] = fetchMock.mock.calls[0] ?? []
    expect(url).toBe(`${BASE_URL}/oauth/token`)
    expect(options.method).toBe('POST')
    expect(options.credentials).toBe('include')
    expect(options.headers).toEqual({})
    expect(Object.fromEntries(options.body as URLSearchParams)).toEqual({
      grant_type: 'authorization_code',
      code: 'the-code',
      code_verifier: 'the-verifier',
      client_id: 'trading-board-web',
      redirect_uri: 'https://web.example/auth/callback',
    })
  })

  it('換授權碼被拒 → 照板子的說法', async () => {
    stubFetch().mockRejectedValue(oauthFailure(400, 'invalid_grant', '授權碼無效或已過期'))

    await expect(newProxy().exchangeCode('c', 'v')).rejects.toThrow(new BoardRejectedError('授權碼無效或已過期'))
  })

  it('續期：只送 grant_type 與用戶端代號，續期憑證由瀏覽器的 cookie 帶', async () => {
    const fetchMock = stubFetch().mockResolvedValue({ access_token: 'renewed', token_type: 'Bearer', expires_in: 60 })

    const session = await newProxy().refresh()

    expect(session.accessToken).toBe('renewed')
    expect(session.expiresAt.toISOString()).toBe('2026-11-25T12:01:00.000Z')
    const [, options] = fetchMock.mock.calls[0] ?? []
    expect(options.credentials).toBe('include')
    expect(Object.fromEntries(options.body as URLSearchParams))
      .toEqual({ grant_type: 'refresh_token', client_id: 'trading-board-web' })
  })

  it('登出：POST /oauth/revoke，帶 cookie', async () => {
    const fetchMock = stubFetch().mockResolvedValue('')

    await newProxy().revoke()

    const [url, options] = fetchMock.mock.calls[0] ?? []
    expect(url).toBe(`${BASE_URL}/oauth/revoke`)
    expect(options.method).toBe('POST')
    expect(options.credentials).toBe('include')
  })

  it('我是誰：帶通行憑證問 /me', async () => {
    const fetchMock = stubFetch().mockResolvedValue({ email: 'owner@example.com' })

    const owner = await newProxy('at').fetchCurrentOwner()

    expect(owner.email).toBe('owner@example.com')
    expect(fetchMock.mock.calls[0]?.[1]?.headers).toEqual({ Authorization: 'Bearer at' })
    expect(fetchMock.mock.calls[0]?.[1]?.credentials).toBeUndefined()
  })
})
