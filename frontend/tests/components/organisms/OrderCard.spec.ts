import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import OrderCard from '~/components/organisms/OrderCard.vue'
import type { FillReportDto } from '~/domain/models/dto/fill-report-dto'
import type { OrderActionDto } from '~/domain/models/dto/order-action-dto'
import type { OrderDto } from '~/domain/models/dto/order-dto'
import { aFillDto, anOrderDto } from '../../fixtures/dtos'

function mountCard(order: OrderDto = anOrderDto(), busy = false, failureMessage: string | null = null) {
  return mount(OrderCard, { props: { order, busy, failureMessage } })
}

describe('OrderCard：一筆委託', () => {
  it('列出標題、標籤、狀態、明細、理由與已成交股數', () => {
    const wrapper = mountCard(anOrderDto('buy-1', true, 60, 'buy', [aFillDto()]))

    expect(wrapper.text()).toContain('聯電 2303')
    expect(wrapper.text()).toContain('月底配置')
    expect(wrapper.find('[data-testid="order-status"]').text()).toBe('待成交')
    expect(wrapper.text()).toContain('限價')
    expect(wrapper.text()).toContain('底部突破')
    expect(wrapper.text()).toContain('已成交 30 / 90 股')
    expect(wrapper.findAll('[data-testid="order-fill"]')).toHaveLength(1)
    expect(wrapper.find('[data-testid="order-fill"]').text()).toContain('52.4 × 30 股')
  })

  it('不能操作的委託 → 沒有回報與略過', () => {
    const wrapper = mountCard(anOrderDto('buy-1', false, 0))

    expect(wrapper.find('[data-testid="order-report"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="order-skip"]').exists()).toBe(false)
  })

  it('按「回報成交」→ 打開表單，價格是限價、股數是剩餘股數、手續費空白', async () => {
    const wrapper = mountCard(anOrderDto('buy-1', true, 30))

    await wrapper.find('[data-testid="order-report"]').trigger('click')

    expect((wrapper.find('[data-testid="fill-price"]').element as HTMLInputElement).value).toBe('52.5')
    expect((wrapper.find('[data-testid="fill-shares"]').element as HTMLInputElement).value).toBe('30')
    expect((wrapper.find('[data-testid="fill-fee"]').element as HTMLInputElement).value).toBe('')
  })

  it('送出表單 → 把打的字交出去', async () => {
    const wrapper = mountCard(anOrderDto('buy-1', true, 30))
    await wrapper.find('[data-testid="order-report"]').trigger('click')

    await wrapper.find('[data-testid="fill-shares"]').setValue('20')
    await wrapper.find('form').trigger('submit')

    const [fillReport] = wrapper.emitted('report')?.[0] as [FillReportDto]
    expect([fillReport.orderId, fillReport.tradeDate, fillReport.remainingShares, fillReport.priceText, fillReport.sharesText])
      .toEqual(['buy-1', '2026-11-26', 30, '52.5', '20'])
  })

  it('按「略過」→ 交出是哪一筆、在哪一張卡上', async () => {
    const wrapper = mountCard()

    await wrapper.find('[data-testid="order-skip"]').trigger('click')

    const [orderAction] = wrapper.emitted('skip')?.[0] as [OrderActionDto]
    expect([orderAction.orderId, orderAction.tradeDate]).toEqual(['buy-1', '2026-11-26'])
  })

  it('表單開著時被拒絕 → 理由顯示在表單裡', async () => {
    const wrapper = mountCard()
    await wrapper.find('[data-testid="order-report"]').trigger('click')

    await wrapper.setProps({ failureMessage: '超出本月剩餘預算' })

    expect(wrapper.find('[data-testid="fill-error"]').text()).toBe('超出本月剩餘預算')
  })

  it('略過被拒絕（表單沒開）→ 理由顯示在卡片上', () => {
    const wrapper = mountCard(anOrderDto(), false, '已全部成交的委託不能略過')

    expect(wrapper.find('[data-testid="order-error"]').text()).toBe('已全部成交的委託不能略過')
  })

  it('取消 → 收起表單', async () => {
    const wrapper = mountCard()
    await wrapper.find('[data-testid="order-report"]').trigger('click')

    await wrapper.findAll('button').find(button => button.text() === '取消')?.trigger('click')

    expect(wrapper.find('form').exists()).toBe(false)
  })

  it('板子收下後換來新的委託 → 表單收起來', async () => {
    const wrapper = mountCard()
    await wrapper.find('[data-testid="order-report"]').trigger('click')

    await wrapper.setProps({ order: anOrderDto('buy-1', false, 0) })

    expect(wrapper.find('form').exists()).toBe(false)
  })

  it('沒有理由 → 不畫理由那一塊', () => {
    const wrapper = mountCard(anOrderDto('buy-1', true, 90, 'buy', [], ''))

    expect(wrapper.find('.order-card__reason').exists()).toBe(false)
  })
})
