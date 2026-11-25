import Decimal from 'decimal.js'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { OrderProxy } from '~/infrastructure/proxy/order-proxy'
import type { ISessionStorageProxy } from '~/domain/interface/i-session-storage-proxy'
import { FillRequestVo } from '~/domain/models/vo/fill-request-vo'

const BASE_URL = '/mock/api/v1'
const NO_SESSION: ISessionStorageProxy = {
  readSession: () => null,
  writeSession: vi.fn<ISessionStorageProxy['writeSession']>(),
  clearSession: vi.fn<ISessionStorageProxy['clearSession']>(),
}

function stubFetch() {
  const fetchMock = vi.fn()
  vi.stubGlobal('$fetch', fetchMock)

  return fetchMock
}

afterEach(() => {
  vi.unstubAllGlobals()
})

describe('OrderProxy', () => {
  it('回報成交：價格以字串送出、時間是 RFC 3339；沒填手續費就整個不送', async () => {
    const fetchMock = stubFetch().mockResolvedValue({})

    await new OrderProxy(BASE_URL, NO_SESSION).reportFill('b1',
      new FillRequestVo(new Decimal('52.5'), 30, undefined, new Date('2026-11-26T01:05:00Z')))

    expect(fetchMock.mock.calls[0]?.[0]).toBe(`${BASE_URL}/orders/b1/fills`)
    expect(fetchMock.mock.calls[0]?.[1]?.method).toBe('POST')
    expect(fetchMock.mock.calls[0]?.[1]?.body).toEqual({ price: '52.5', shares: 30, filledAt: '2026-11-26T01:05:00.000Z' })
  })

  it('回報成交：填了手續費就帶上', async () => {
    const fetchMock = stubFetch().mockResolvedValue({})

    await new OrderProxy(BASE_URL, NO_SESSION).reportFill('b1',
      new FillRequestVo(new Decimal('52.5'), 30, new Decimal('20'), new Date('2026-11-26T01:05:00Z')))

    expect(fetchMock.mock.calls[0]?.[1]?.body).toMatchObject({ fee: '20' })
  })

  it('略過', async () => {
    const fetchMock = stubFetch().mockResolvedValue({})

    await new OrderProxy(BASE_URL, NO_SESSION).skipOrder('b1')

    expect(fetchMock.mock.calls[0]?.[0]).toBe(`${BASE_URL}/orders/b1/skip`)
    expect(fetchMock.mock.calls[0]?.[1]?.method).toBe('POST')
  })
})
