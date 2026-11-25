import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { TradeCardApplication } from '~/application/trade-card-application'
import { BoardRejectedError } from '~/domain/errors/board-rejected-error'
import { FillDraftInvalidError } from '~/domain/errors/fill-draft-invalid-error'
import type { IOrderProxy } from '~/domain/interface/i-order-proxy'
import type { ITradeCardProxy } from '~/domain/interface/i-trade-card-proxy'
import { FillReportDto } from '~/domain/models/dto/fill-report-dto'
import { OrderActionDto } from '~/domain/models/dto/order-action-dto'
import { TradeCardSummary } from '~/domain/models/entities/trade-card-summary'
import { TradeCardService } from '~/domain/service/trade-card-service'
import { aBuyOrder, aFill, aNoTradeCard, aSellOrder, aTradeCard } from '../fixtures/entities'

/** 台北 2026-11-25 20:30。 */
const NOW = new Date('2026-11-25T12:30:00Z')

function setUp() {
  const tradeCardProxy = {
    fetchLatestCard: vi.fn<ITradeCardProxy['fetchLatestCard']>(),
    fetchCard: vi.fn<ITradeCardProxy['fetchCard']>(),
    listCards: vi.fn<ITradeCardProxy['listCards']>(),
  } satisfies ITradeCardProxy
  const orderProxy = {
    reportFill: vi.fn<IOrderProxy['reportFill']>(),
    skipOrder: vi.fn<IOrderProxy['skipOrder']>(),
  } satisfies IOrderProxy
  const application = new TradeCardApplication(new TradeCardService(tradeCardProxy, orderProxy))

  return { tradeCardProxy, orderProxy, application }
}

function aFillReport(sharesText: string, priceText = '52', feeText = '', remainingShares = 30): FillReportDto {
  return new FillReportDto('buy-1', '2026-11-26', remainingShares, priceText, sharesText, feeText, '2026-11-26T09:05')
}

beforeEach(() => {
  vi.useFakeTimers({ toFake: ['Date'] })
  vi.setSystemTime(NOW)
})

afterEach(() => {
  vi.useRealTimers()
})

describe('TradeCardApplication：今日的交易卡', () => {
  it('有買有賣的交易卡 → 買進段是聯電、賣出段是晶豪科，每個欄位都在', async () => {
    const { tradeCardProxy, application } = setUp()
    tradeCardProxy.fetchLatestCard.mockResolvedValue(aTradeCard())

    const card = await application.fetchLatestCard()

    expect(card?.tradeDateLabel).toBe('2026/11/26（四）')
    expect(card?.isNoTrade).toBe(false)
    expect(card?.buyOrders.map(order => order.title)).toEqual(['聯電 2303'])
    expect(card?.buyOrders[0]?.detailRows.map(row => [row.label, row.value])).toEqual([
      ['限價', '52.5'],
      ['最高追價', '54'],
      ['股數', '90 股'],
      ['預估金額', 'NT$4,725'],
      ['停損', '48'],
      ['停利', '60'],
    ])
    expect(card?.buyOrders[0]?.reason).toBe('底部突破')
    expect(card?.buyOrders[0]?.statusLabel).toBe('待成交')
    expect(card?.sellOrders.map(order => order.title)).toEqual(['晶豪科 3006'])
    expect(card?.sellOrders[0]?.detailRows.map(row => [row.label, row.value])).toEqual([
      ['限價', '68'],
      ['股數', '50 股'],
      ['出場原因', '停利'],
    ])
    expect(card?.sellOrders[0]?.reason).toBe('到達停利')
  })

  it('沒有賣出委託 → 賣出段空白，說「無賣出委託」', async () => {
    const { tradeCardProxy, application } = setUp()
    tradeCardProxy.fetchLatestCard.mockResolvedValue(aTradeCard('2026-11-26', [aBuyOrder()], []))

    const card = await application.fetchLatestCard()

    expect(card?.sellOrders).toEqual([])
    expect(card?.emptySellText).toBe('無賣出委託')
  })

  it.each([
    ['月底配置的買進 → 標上「月底配置」', true, ['月底配置']],
    ['一般買進 → 沒有標籤', false, []],
  ])('%s', async (_, isMonthEndAllocation, tags) => {
    const { tradeCardProxy, application } = setUp()
    tradeCardProxy.fetchLatestCard.mockResolvedValue(aTradeCard('2026-11-26', [aBuyOrder('pending', 0, isMonthEndAllocation)]))

    const card = await application.fetchLatestCard()

    expect(card?.buyOrders[0]?.tags).toEqual(tags)
  })

  it('今日不買的交易卡 → 今日不買、附上說明、沒有委託', async () => {
    const { tradeCardProxy, application } = setUp()
    tradeCardProxy.fetchLatestCard.mockResolvedValue(aNoTradeCard('大盤轉弱'))

    const card = await application.fetchLatestCard()

    expect(card?.isNoTrade).toBe(true)
    expect(card?.note).toBe('大盤轉弱')
    expect(card?.buyOrders).toEqual([])
    expect(card?.sellOrders).toEqual([])
    expect(card?.publishedAtText).toBe('')
  })

  it('還沒有任何交易卡 → null', async () => {
    const { tradeCardProxy, application } = setUp()
    tradeCardProxy.fetchLatestCard.mockResolvedValue(null)

    expect(await application.fetchLatestCard()).toBeNull()
  })

  it('一股都還沒成交的 90 股委託 → 表單預設限價 52.5、股數 90', async () => {
    const { tradeCardProxy, application } = setUp()
    tradeCardProxy.fetchLatestCard.mockResolvedValue(aTradeCard('2026-11-26', [aBuyOrder('pending', 0)]))

    const order = (await application.fetchLatestCard())?.buyOrders[0]

    expect([order?.defaultFillPrice, order?.defaultFillShares]).toEqual(['52.5', '90'])
  })

  it('回報成交表單的初始值：限價、剩餘股數、現在（台北）', async () => {
    const { tradeCardProxy, application } = setUp()
    tradeCardProxy.fetchLatestCard.mockResolvedValue(aTradeCard('2026-11-26', [aBuyOrder('partiallyFilled', 60)]))

    const order = (await application.fetchLatestCard())?.buyOrders[0]

    expect([order?.defaultFillPrice, order?.defaultFillShares, order?.defaultFilledAt]).toEqual(['52.5', '30', '2026-11-25T20:30'])
  })

  it('賣出委託也列出已回報的成交', async () => {
    const { tradeCardProxy, application } = setUp()
    tradeCardProxy.fetchLatestCard.mockResolvedValue(aTradeCard('2026-11-26', [], [aSellOrder('partiallyFilled', 30, [aFill(30)])]))

    const order = (await application.fetchLatestCard())?.sellOrders[0]

    expect(order?.fills).toHaveLength(1)
    expect(order?.filledSharesText).toBe('已成交 30 / 50 股')
  })

  it('委託上列出已回報的成交與已成交股數', async () => {
    const { tradeCardProxy, application } = setUp()
    tradeCardProxy.fetchLatestCard.mockResolvedValue(
      aTradeCard('2026-11-26', [aBuyOrder('partiallyFilled', 30, false, [aFill(30, '2')])]))

    const order = (await application.fetchLatestCard())?.buyOrders[0]

    expect(order?.filledSharesText).toBe('已成交 30 / 90 股')
    expect(order?.fills.map(fill => [fill.priceText, fill.sharesText, fill.feeText, fill.amountText, fill.filledAtText]))
      .toEqual([['52.5', '30 股', 'NT$2', 'NT$1,575', '2026/11/26 09:05']])
  })
})

describe('TradeCardApplication：回報成交', () => {
  it('剩 30 股、回報 30 股 → 送給板子，換回板子現在的樣子（已成交）', async () => {
    const { tradeCardProxy, orderProxy, application } = setUp()
    tradeCardProxy.fetchCard.mockResolvedValue(aTradeCard('2026-11-26', [aBuyOrder('filled', 90)]))

    const card = await application.reportFill(aFillReport('30'))

    expect(orderProxy.reportFill).toHaveBeenCalledTimes(1)
    const [orderId, request] = orderProxy.reportFill.mock.calls[0] ?? []
    expect(orderId).toBe('buy-1')
    expect([request?.price.toFixed(), request?.shares, request?.fee]).toEqual(['52', 30, undefined])
    expect(tradeCardProxy.fetchCard).toHaveBeenCalledWith('2026-11-26')
    expect(card?.buyOrders[0]?.statusLabel).toBe('已成交')
  })

  it.each([
    ['剩 30 股卻回報 31 股', '31', '52', '股數不可超過剩餘 30 股'],
    ['股數 0', '0', '52', '股數須為 1 以上的整數'],
    ['價格 0', '10', '0', '價格須大於 0'],
  ])('%s → 表單拒絕，不送給板子', async (_, shares, price, message) => {
    const { orderProxy, application } = setUp()

    await expect(application.reportFill(aFillReport(shares, price))).rejects.toThrow(new FillDraftInvalidError(message))
    expect(orderProxy.reportFill).not.toHaveBeenCalled()
  })

  it('板子拒絕（超出本月預算）→ 照板子的話往上拋，不重讀交易卡', async () => {
    const { tradeCardProxy, orderProxy, application } = setUp()
    orderProxy.reportFill.mockRejectedValue(new BoardRejectedError('超出本月剩餘預算'))

    await expect(application.reportFill(aFillReport('30'))).rejects.toThrow(new BoardRejectedError('超出本月剩餘預算'))
    expect(tradeCardProxy.fetchCard).not.toHaveBeenCalled()
  })
})

describe('TradeCardApplication：略過', () => {
  it('略過 → 請板子略過那一筆，換回已略過的交易卡', async () => {
    const { tradeCardProxy, orderProxy, application } = setUp()
    tradeCardProxy.fetchCard.mockResolvedValue(aTradeCard('2026-11-26', [aBuyOrder('skipped')]))

    const card = await application.skipOrder(new OrderActionDto('buy-1', '2026-11-26'))

    expect(orderProxy.skipOrder).toHaveBeenCalledWith('buy-1')
    expect(card?.buyOrders[0]?.statusLabel).toBe('已略過')
    expect(card?.buyOrders[0]?.actionable).toBe(false)
  })
})

describe('TradeCardApplication：歷史', () => {
  it('沒指定月份 → 本月（台北 2026-11），清單新的在前', async () => {
    const { tradeCardProxy, application } = setUp()
    tradeCardProxy.listCards.mockResolvedValue([
      new TradeCardSummary('2026-11-20', 'trade', 2, 1, 1),
      new TradeCardSummary('2026-11-24', 'noTrade', 0, 0, 0),
    ])

    const list = await application.listCards()

    expect(tradeCardProxy.listCards).toHaveBeenCalledWith('2026-11')
    expect(list.month.label).toBe('2026 年 11 月')
    expect(list.month.next).toBeNull()
    expect(list.cards.map(card => [card.tradeDate, card.statusLabel, card.orderCountText, card.filledOrderCountText])).toEqual([
      ['2026-11-24', '今日不買', '買 0・賣 0', '已成交 0 筆'],
      ['2026-11-20', '交易', '買 2・賣 1', '已成交 1 筆'],
    ])
  })

  it('指定過去的月份 → 查那個月', async () => {
    const { tradeCardProxy, application } = setUp()
    tradeCardProxy.listCards.mockResolvedValue([])

    const list = await application.listCards('2026-10')

    expect(tradeCardProxy.listCards).toHaveBeenCalledWith('2026-10')
    expect(list.cards).toEqual([])
  })

  it('打開 11/20 的交易卡（今天 11/25）→ 還沒成交的委託已過期', async () => {
    const { tradeCardProxy, application } = setUp()
    tradeCardProxy.fetchCard.mockResolvedValue(aTradeCard('2026-11-20', [aBuyOrder('pending')], [aSellOrder('partiallyFilled', 20)]))

    const card = await application.fetchCard('2026-11-20')

    expect(tradeCardProxy.fetchCard).toHaveBeenCalledWith('2026-11-20')
    expect(card?.buyOrders[0]?.statusLabel).toBe('已過期')
    expect(card?.buyOrders[0]?.actionable).toBe(false)
    expect(card?.sellOrders[0]?.statusLabel).toBe('部分成交')
    expect(card?.sellOrders[0]?.actionable).toBe(true)
  })
})
