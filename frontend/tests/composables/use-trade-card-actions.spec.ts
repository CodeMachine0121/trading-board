import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { ref } from 'vue'
import { TradeCardApplication } from '~/application/trade-card-application'
import { useTradeCardActions } from '~/composables/use-trade-card-actions'
import { BoardRejectedError } from '~/domain/errors/board-rejected-error'
import type { IOrderProxy } from '~/domain/interface/i-order-proxy'
import type { ITradeCardProxy } from '~/domain/interface/i-trade-card-proxy'
import { FillReportDto } from '~/domain/models/dto/fill-report-dto'
import { OrderActionDto } from '~/domain/models/dto/order-action-dto'
import type { TradeCardDto } from '~/domain/models/dto/trade-card-dto'
import { TradeCardService } from '~/domain/service/trade-card-service'
import { aBuyOrder, aTradeCard } from '../fixtures/entities'
import { aTradeCardDto } from '../fixtures/dtos'

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
  const card = ref<TradeCardDto | null>(aTradeCardDto())
  const onCardChanged = vi.fn()
  const actions = useTradeCardActions(new TradeCardApplication(new TradeCardService(tradeCardProxy, orderProxy)), card, onCardChanged)

  return { tradeCardProxy, orderProxy, card, onCardChanged, actions }
}

const FILL_REPORT = new FillReportDto('buy-1', '2026-11-26', 90, '52.5', '90', '', '2026-11-26T09:05')

beforeEach(() => {
  vi.useFakeTimers({ toFake: ['Date'] })
  vi.setSystemTime(new Date('2026-11-25T12:30:00Z'))
})

afterEach(() => {
  vi.useRealTimers()
})

describe('useTradeCardActions', () => {
  it('回報成功 → 卡換成板子現在的樣子，並通知外面（今日要重讀預算）', async () => {
    const { tradeCardProxy, card, onCardChanged, actions } = setUp()
    tradeCardProxy.fetchCard.mockResolvedValue(aTradeCard('2026-11-26', [aBuyOrder('filled', 90)]))

    await actions.reportFill(FILL_REPORT)

    expect(card.value?.buyOrders[0]?.statusLabel).toBe('已成交')
    expect(onCardChanged).toHaveBeenCalledTimes(1)
    expect([actions.busyOrderId.value, actions.failedOrderId.value]).toEqual([null, null])
  })

  it('略過成功 → 卡換成已略過', async () => {
    const { tradeCardProxy, card, actions } = setUp()
    tradeCardProxy.fetchCard.mockResolvedValue(aTradeCard('2026-11-26', [aBuyOrder('skipped')]))

    await actions.skipOrder(new OrderActionDto('buy-1', '2026-11-26'))

    expect(card.value?.buyOrders[0]?.statusLabel).toBe('已略過')
  })

  it('板子拒絕 → 卡維持原樣，理由掛在那一筆上', async () => {
    const { orderProxy, card, onCardChanged, actions } = setUp()
    const before = card.value
    orderProxy.reportFill.mockRejectedValue(new BoardRejectedError('超出本月剩餘預算'))

    await actions.reportFill(FILL_REPORT)

    expect(card.value).toBe(before)
    expect(onCardChanged).not.toHaveBeenCalled()
    expect([actions.failedOrderId.value, actions.failureMessage.value]).toEqual(['buy-1', '超出本月剩餘預算'])
    expect(actions.busyOrderId.value).toBeNull()
  })

  it('送出中 → 那一筆是忙碌的', async () => {
    const { orderProxy, actions } = setUp()
    orderProxy.skipOrder.mockReturnValue(new Promise(() => {}))

    void actions.skipOrder(new OrderActionDto('buy-1', '2026-11-26'))

    expect(actions.busyOrderId.value).toBe('buy-1')
  })

  it('板子回來時那張卡不見了 → 留著原本的卡', async () => {
    const { tradeCardProxy, card, actions } = setUp()
    const before = card.value
    tradeCardProxy.fetchCard.mockResolvedValue(null)

    await actions.skipOrder(new OrderActionDto('buy-1', '2026-11-26'))

    expect(card.value).toBe(before)
  })

  it('認不得的錯誤 → 通用訊息並記在主控台', async () => {
    const consoleError = vi.spyOn(console, 'error').mockImplementation(() => {})
    const { orderProxy, actions } = setUp()
    orderProxy.skipOrder.mockRejectedValue(new Error('boom'))

    await actions.skipOrder(new OrderActionDto('buy-1', '2026-11-26'))

    expect(actions.failureMessage.value).toBe('發生未預期的錯誤，請重新整理')
    expect(consoleError).toHaveBeenCalled()
    consoleError.mockRestore()
  })

  it('沒給「卡換了」的通知也照樣運作（歷史明細不必重讀預算）', async () => {
    const { tradeCardProxy, orderProxy } = setUp()
    tradeCardProxy.fetchCard.mockResolvedValue(aTradeCard('2026-11-26', [aBuyOrder('skipped')]))
    const card = ref<TradeCardDto | null>(null)
    const actions = useTradeCardActions(new TradeCardApplication(new TradeCardService(tradeCardProxy, orderProxy)), card)

    await actions.skipOrder(new OrderActionDto('buy-1', '2026-11-26'))

    expect(card.value?.buyOrders[0]?.statusLabel).toBe('已略過')
  })
})
