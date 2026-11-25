import Decimal from 'decimal.js'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { TradeCardProxy } from '~/infrastructure/proxy/trade-card-proxy'

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

/** 契約 §4 的交易卡範例。 */
const CARD_WIRE = {
  tradeDate: '2026-11-26',
  status: 'trade',
  note: '偏多',
  buyOrders: [{
    id: 'b1', symbol: '2303', name: '聯電', limitPrice: '52.5', maxChasePrice: '54', shares: 90,
    estimatedAmount: '4725', stopLossPrice: '48', takeProfitPrice: '60', reason: '突破', isMonthEndAllocation: true,
    status: 'partiallyFilled', filledShares: 30,
    fills: [{ id: 'f1', price: '52.4', shares: 30, fee: '1', amount: '1572', filledAt: '2026-11-26T09:05:00+08:00' }],
  }],
  sellOrders: [{
    id: 's1', symbol: '3006', name: '晶豪科', limitPrice: '68', shares: 50, exitReason: 'stopLoss', reason: '跌破',
    status: 'pending', filledShares: 0, fills: [],
  }],
  publishedAt: '2026-11-25T20:30:00+08:00',
}

describe('TradeCardProxy：wire → entity', () => {
  it('金額轉成精確小數、日期轉成時刻、欄位一個不漏', async () => {
    const fetchMock = stubFetch().mockResolvedValue(CARD_WIRE)

    const card = await new TradeCardProxy(BASE_URL, NO_SESSION).fetchCard('2026-11-26')

    expect(fetchMock.mock.calls[0]?.[0]).toBe(`${BASE_URL}/cards/2026-11-26`)
    expect(card?.note).toBe('偏多')
    expect(card?.publishedAt?.toISOString()).toBe('2026-11-25T12:30:00.000Z')
    const buyOrder = card?.buyOrders[0]
    expect(buyOrder?.limitPrice).toEqual(new Decimal('52.5'))
    expect(buyOrder?.maxChasePrice.toFixed()).toBe('54')
    expect(buyOrder?.estimatedAmount.toFixed()).toBe('4725')
    expect([buyOrder?.stopLossPrice.toFixed(), buyOrder?.takeProfitPrice.toFixed()]).toEqual(['48', '60'])
    expect([buyOrder?.isMonthEndAllocation, buyOrder?.status, buyOrder?.filledShares, buyOrder?.reason]).toEqual([true, 'partiallyFilled', 30, '突破'])
    expect(buyOrder?.fills[0]?.fee.toFixed()).toBe('1')
    expect(buyOrder?.fills[0]?.filledAt.toISOString()).toBe('2026-11-26T01:05:00.000Z')
    expect([card?.sellOrders[0]?.exitReason, card?.sellOrders[0]?.limitPrice.toFixed()]).toEqual(['stopLoss', '68'])
  })

  it('null 與缺欄位收成明確的預設值', async () => {
    stubFetch().mockResolvedValue({
      tradeDate: '2026-11-26', status: 'noTrade', note: null, buyOrders: null, sellOrders: null, publishedAt: null,
    })

    const card = await new TradeCardProxy(BASE_URL, NO_SESSION).fetchLatestCard()

    expect([card?.note, card?.buyOrders, card?.sellOrders, card?.publishedAt]).toEqual(['', [], [], null])
  })

  it('委託上的選填欄位缺漏 → 空說明、非月底配置、0 股成交、沒有成交', async () => {
    stubFetch().mockResolvedValue({
      ...CARD_WIRE,
      buyOrders: [{ ...CARD_WIRE.buyOrders[0], reason: null, isMonthEndAllocation: undefined, filledShares: undefined, fills: null }],
      sellOrders: [{ ...CARD_WIRE.sellOrders[0], reason: undefined, filledShares: undefined, fills: undefined }],
    })

    const card = await new TradeCardProxy(BASE_URL, NO_SESSION).fetchLatestCard()

    expect([card?.buyOrders[0]?.reason, card?.buyOrders[0]?.isMonthEndAllocation, card?.buyOrders[0]?.filledShares, card?.buyOrders[0]?.fills])
      .toEqual(['', false, 0, []])
    expect([card?.sellOrders[0]?.reason, card?.sellOrders[0]?.filledShares, card?.sellOrders[0]?.fills]).toEqual(['', 0, []])
  })

  it('找不到那一天的交易卡 → null', async () => {
    stubFetch().mockRejectedValue(Object.assign(new Error('x'), { response: { status: 404 } }))

    expect(await new TradeCardProxy(BASE_URL, NO_SESSION).fetchCard('2026-01-01')).toBeNull()
  })

  it('一個月的清單 → 摘要', async () => {
    stubFetch().mockResolvedValue([{ tradeDate: '2026-11-26', status: 'trade', buyCount: 2, sellCount: 1, filledOrderCount: 1 }])

    const summaries = await new TradeCardProxy(BASE_URL, NO_SESSION).listCards('2026-11')

    expect(summaries.map(summary => [summary.tradeDate, summary.status, summary.buyCount, summary.sellCount, summary.filledOrderCount]))
      .toEqual([['2026-11-26', 'trade', 2, 1, 1]])
  })
})
