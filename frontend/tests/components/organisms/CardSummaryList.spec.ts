import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import CardSummaryList from '~/components/organisms/CardSummaryList.vue'
import { TradeCardSummaryDto } from '~/domain/models/dto/trade-card-summary-dto'

const SUMMARY = new TradeCardSummaryDto('2026-11-20', '2026/11/20（五）', '交易', 'info', '買 2・賣 1', '已成交 1 筆')

describe('CardSummaryList', () => {
  it('這個月沒有交易卡 → 「這個月沒有交易卡」', () => {
    const wrapper = mount(CardSummaryList, { props: { cards: [] } })

    expect(wrapper.find('[data-testid="card-list-empty"]').text()).toBe('這個月沒有交易卡')
  })

  it('每一列：日期、交易／今日不買、買賣筆數、已成交筆數', () => {
    const row = mount(CardSummaryList, { props: { cards: [SUMMARY] } }).find('[data-testid="card-summary-row"]')

    expect(row.text()).toContain('2026/11/20（五）')
    expect(row.text()).toContain('交易')
    expect(row.text()).toContain('買 2・賣 1')
    expect(row.text()).toContain('已成交 1 筆')
  })

  it('點一列 → 交出那一天', async () => {
    const wrapper = mount(CardSummaryList, { props: { cards: [SUMMARY] } })

    await wrapper.find('[data-testid="card-summary-row"]').trigger('click')

    expect(wrapper.emitted('select')?.[0]).toEqual(['2026-11-20'])
  })
})
