import { afterEach, describe, expect, it, vi } from 'vitest'
import { BoardRejectedError } from '~/domain/errors/board-rejected-error'
import { BoardUnreachableError } from '~/domain/errors/board-unreachable-error'
import { SignedOutError } from '~/domain/errors/signed-out-error'
import type { ISessionStorageProxy } from '~/domain/interface/i-session-storage-proxy'
import { Session } from '~/domain/models/entities/session'
import { TradeCardProxy } from '~/infrastructure/proxy/trade-card-proxy'

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

describe('BoardApiProxy：每一發請求都帶著登入，打設定的那個板子', () => {
  it('打的是設定的 base URL，並帶 Bearer 憑證', async () => {
    const fetchMock = stubFetch().mockResolvedValue([])
    const proxy = new TradeCardProxy(BASE_URL, sessionStorageWith(new Session('abc', new Date())))

    await proxy.listCards('2026-11')

    expect(fetchMock).toHaveBeenCalledWith(`${BASE_URL}/cards`, {
      method: 'GET',
      query: { month: '2026-11' },
      body: undefined,
      headers: { Authorization: 'Bearer abc' },
    })
  })

  it('這台瀏覽器沒有登入 → 不帶憑證', async () => {
    const fetchMock = stubFetch().mockResolvedValue([])

    await new TradeCardProxy(BASE_URL, sessionStorageWith(null)).listCards('2026-11')

    expect(fetchMock.mock.calls[0]?.[1]?.headers).toEqual({})
  })
})

describe('BoardApiProxy：失敗翻成領域認得的錯誤', () => {
  it('401 → 登入已失效，並通知組裝根（清登入、回登入畫面）一次', async () => {
    stubFetch().mockRejectedValue(boardFailure(401, 'token expired'))
    const onSignedOut = vi.fn()
    const proxy = new TradeCardProxy(BASE_URL, sessionStorageWith(null), onSignedOut)

    await expect(proxy.listCards('2026-11')).rejects.toBeInstanceOf(SignedOutError)
    expect(onSignedOut).toHaveBeenCalledTimes(1)
  })

  it('401 而組裝根沒交代要做什麼 → 一樣是登入已失效', async () => {
    stubFetch().mockRejectedValue(boardFailure(401))

    await expect(new TradeCardProxy(BASE_URL, sessionStorageWith(null)).listCards('2026-11')).rejects.toBeInstanceOf(SignedOutError)
  })

  it.each([
    ['400 → 照板子的話拒絕', 400, '價格須大於 0', '價格須大於 0'],
    ['409 → 照板子的話拒絕', 409, '超出本月剩餘預算', '超出本月剩餘預算'],
    ['板子沒附說明 → 通用說法', 409, undefined, '板子拒絕了這次操作'],
  ])('%s', async (_, status, message, expected) => {
    stubFetch().mockRejectedValue(boardFailure(status, message))

    await expect(new TradeCardProxy(BASE_URL, sessionStorageWith(null)).listCards('2026-11'))
      .rejects.toThrow(new BoardRejectedError(expected))
  })

  it.each([
    ['500 → 連不上', boardFailure(500)],
    ['網路斷線（沒有回應）→ 連不上', new TypeError('Failed to fetch')],
    ['拋出的不是物件 → 連不上', 'boom'],
  ])('%s', async (_, failure) => {
    stubFetch().mockRejectedValue(failure)

    await expect(new TradeCardProxy(BASE_URL, sessionStorageWith(null)).listCards('2026-11'))
      .rejects.toBeInstanceOf(BoardUnreachableError)
  })

  it('找不到（404）在允許「沒有」的地方 → null', async () => {
    stubFetch().mockRejectedValue(boardFailure(404, 'no card'))

    expect(await new TradeCardProxy(BASE_URL, sessionStorageWith(null)).fetchLatestCard()).toBeNull()
  })

  it('允許「沒有」的地方遇到其他失敗 → 照樣翻譯', async () => {
    stubFetch().mockRejectedValue(boardFailure(503))

    await expect(new TradeCardProxy(BASE_URL, sessionStorageWith(null)).fetchLatestCard())
      .rejects.toBeInstanceOf(BoardUnreachableError)
  })
})
