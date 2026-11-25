import { afterEach, describe, expect, it, vi } from 'vitest'
import { MonthSummaryProxy } from '~/infrastructure/proxy/month-summary-proxy'
import type { ISessionStorageProxy } from '~/domain/interface/i-session-storage-proxy'

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

describe('MonthSummaryProxy', () => {
  const SUMMARY_WIRE = {
    month: '2026-11', allocation: '10000', carriedOver: '0', recycledProceeds: '3300', spent: '9800', remaining: '3500',
    mode: 'monthEnd', realizedPnl: '1200', closedPositionCount: 3, winCount: 2, lossCount: 1, winRate: '0.6667',
    realizedReturnPct: '8.5',
  }

  it('月報 → 金額轉成精確小數', async () => {
    const fetchMock = stubFetch().mockResolvedValue(SUMMARY_WIRE)

    const summary = await new MonthSummaryProxy(BASE_URL, NO_SESSION).fetchMonthSummary('2026-11')

    expect(fetchMock.mock.calls[0]?.[0]).toBe(`${BASE_URL}/months/2026-11/summary`)
    expect([summary?.recycledProceeds.toFixed(), summary?.mode, summary?.winCount, summary?.realizedReturnPct?.toFixed()])
      .toEqual(['3300', 'monthEnd', 2, '8.5'])
  })

  it('不認得的期間 → 當作嚴格期；沒有報酬率 → null', async () => {
    stubFetch().mockResolvedValue({ ...SUMMARY_WIRE, mode: 'weird', realizedReturnPct: null })

    const summary = await new MonthSummaryProxy(BASE_URL, NO_SESSION).fetchMonthSummary('2026-11')

    expect([summary?.mode, summary?.realizedReturnPct]).toEqual(['strict', null])
  })

  it('那個月沒有資料（404）→ null', async () => {
    stubFetch().mockRejectedValue(Object.assign(new Error('x'), { response: { status: 404 } }))

    expect(await new MonthSummaryProxy(BASE_URL, NO_SESSION).fetchMonthSummary('2026-01')).toBeNull()
  })
})
