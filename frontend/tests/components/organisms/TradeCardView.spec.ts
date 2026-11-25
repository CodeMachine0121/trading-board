import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import TradeCardView from '~/components/organisms/TradeCardView.vue'
import type { OrderActionDto } from '~/domain/models/dto/order-action-dto'
import { TradeCardDto } from '~/domain/models/dto/trade-card-dto'
import { aNoTradeCardDto, anOrderDto, aTradeCardDto } from '../../fixtures/dtos'

function mountView(card: TradeCardDto, failedOrderId: string | null = null, failureMessage: string | null = null) {
  return mount(TradeCardView, { props: { card, busyOrderId: null, failedOrderId, failureMessage } })
}

describe('TradeCardView', () => {
  it('交易卡 → 日期、發布時間、買進段與賣出段', () => {
    const wrapper = mountView(aTradeCardDto([anOrderDto('buy-1')], [anOrderDto('sell-1', true, 50, 'sell')], '偏多操作'))

    expect(wrapper.text()).toContain('2026/11/26（四）')
    expect(wrapper.text()).toContain('發布於 2026/11/25 20:30')
    expect(wrapper.text()).toContain('偏多操作')
    expect(wrapper.find('[data-testid="buy-section"]').text()).toContain('聯電 2303')
    expect(wrapper.find('[data-testid="sell-section"]').text()).toContain('晶豪科 3006')
    expect(wrapper.find('[data-testid="no-trade-banner"]').exists()).toBe(false)
  })

  it.each([
    ['有發布者 → 顯示那一行', '發布者：desk-chief', true],
    ['沒有發布者 → 不顯示', '', false],
  ])('%s', (_, publisherText, shown) => {
    const card = aTradeCardDto()
    const wrapper = mountView(new TradeCardDto(card.tradeDate, card.tradeDateLabel, card.isNoTrade, card.note,
      card.buyOrders, card.sellOrders, card.emptyBuyText, card.emptySellText, card.publishedAtText, publisherText))

    expect(wrapper.find('[data-testid="card-publisher"]').exists()).toBe(shown)
    if (shown) {
      expect(wrapper.find('[data-testid="card-publisher"]').text()).toBe('發布者：desk-chief')
    }
  })

  it('沒有賣出委託 → 賣出段說「無賣出委託」', () => {
    const wrapper = mountView(aTradeCardDto([anOrderDto()], []))

    expect(wrapper.find('[data-testid="sell-section"]').text()).toContain('無賣出委託')
  })

  it('沒有買進委託 → 買進段說「無買進委託」', () => {
    const wrapper = mountView(aTradeCardDto([], [anOrderDto('sell-1', true, 50, 'sell')]))

    expect(wrapper.find('[data-testid="buy-section"]').text()).toContain('無買進委託')
  })

  it('今日不買 → 橫幅與說明，沒有委託段', () => {
    const wrapper = mountView(aNoTradeCardDto('大盤轉弱'))

    expect(wrapper.find('[data-testid="no-trade-banner"]').text()).toContain('今日不買')
    expect(wrapper.find('[data-testid="no-trade-banner"]').text()).toContain('大盤轉弱')
    expect(wrapper.find('[data-testid="buy-section"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="sell-section"]').exists()).toBe(false)
  })

  it('今日不買但沒附說明 → 只有橫幅', () => {
    expect(mountView(aNoTradeCardDto('')).find('[data-testid="no-trade-banner"]').text()).toBe('今日不買')
  })

  it('失敗理由只掛在失敗的那一筆上', () => {
    const wrapper = mountView(aTradeCardDto([anOrderDto('buy-1')], [anOrderDto('sell-1', true, 50, 'sell')]), 'sell-1', '持股不足')

    expect(wrapper.find('[data-testid="order-buy-1"]').text()).not.toContain('持股不足')
    expect(wrapper.find('[data-testid="order-sell-1"]').text()).toContain('持股不足')
  })

  it('失敗在買進那一筆 → 只掛在買進那一筆', () => {
    const wrapper = mountView(aTradeCardDto(), 'buy-1', '超出本月剩餘預算')

    expect(wrapper.find('[data-testid="order-buy-1"]').text()).toContain('超出本月剩餘預算')
    expect(wrapper.find('[data-testid="order-sell-1"]').text()).not.toContain('超出本月剩餘預算')
  })

  it('委託的動作往外傳（買進與賣出都是）', async () => {
    const wrapper = mountView(aTradeCardDto())

    await wrapper.find('[data-testid="order-buy-1"] [data-testid="order-skip"]').trigger('click')
    await wrapper.find('[data-testid="order-sell-1"] [data-testid="order-skip"]').trigger('click')
    await wrapper.find('[data-testid="order-buy-1"] [data-testid="order-report"]').trigger('click')
    await wrapper.find('[data-testid="order-buy-1"] form').trigger('submit')
    await wrapper.find('[data-testid="order-sell-1"] [data-testid="order-report"]').trigger('click')
    await wrapper.find('[data-testid="order-sell-1"] form').trigger('submit')

    expect((wrapper.emitted('skip') as [OrderActionDto][]).map(([action]) => action.orderId)).toEqual(['buy-1', 'sell-1'])
    expect(wrapper.emitted('report')).toHaveLength(2)
  })
})
