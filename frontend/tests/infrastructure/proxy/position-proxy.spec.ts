import { afterEach, describe, expect, it, vi } from 'vitest'
import { PositionProxy } from '~/infrastructure/proxy/position-proxy'

const BASE_URL = '/mock/api/v1'
const NO_SESSION = async (): Promise<string | null> => null

function stubFetch() {
  const fetchMock = vi.fn()
  vi.stubGlobal('$fetch', fetchMock)

  return fetchMock
}

afterEach(() => {
  vi.unstubAllGlobals()
})

describe('PositionProxy', () => {
  const OPEN_WIRE = {
    symbol: '2303', name: '聯電', shares: 90, averageCost: '52.58', totalCost: '4732', stopLossPrice: '48',
    takeProfitPrice: '60', openedDate: '2026-11-20', closedDate: null, daysHeld: 5, realizedPnl: '0',
  }

  it('持有中 → status=open，金額轉成精確小數', async () => {
    const fetchMock = stubFetch().mockResolvedValue([OPEN_WIRE])

    const [position] = await new PositionProxy(BASE_URL, NO_SESSION).listOpenPositions()

    expect(fetchMock.mock.calls[0]?.[1]?.query).toEqual({ status: 'open' })
    expect([position?.averageCost.toFixed(), position?.stopLossPrice?.toFixed(), position?.closedDate, position?.daysHeld])
      .toEqual(['52.58', '48', null, 5])
  })

  it('出清 → status=closed&month=，缺少的防守價與損益收成 null / 0', async () => {
    const fetchMock = stubFetch().mockResolvedValue([{
      ...OPEN_WIRE, shares: 0, stopLossPrice: null, takeProfitPrice: undefined, closedDate: '2026-11-24', realizedPnl: undefined,
    }])

    const [position] = await new PositionProxy(BASE_URL, NO_SESSION).listClosedPositions('2026-11')

    expect(fetchMock.mock.calls[0]?.[1]?.query).toEqual({ status: 'closed', month: '2026-11' })
    expect([position?.stopLossPrice, position?.takeProfitPrice, position?.closedDate, position?.realizedPnl.toFixed()])
      .toEqual([null, null, '2026-11-24', '0'])
  })
})
