import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import MonthStepper from '~/components/molecules/MonthStepper.vue'
import { MonthDto } from '~/domain/models/dto/month-dto'

describe('MonthStepper', () => {
  it('顯示月份；本月沒有下一個月可去', () => {
    const wrapper = mount(MonthStepper, { props: { month: new MonthDto('2026-11', '2026 年 11 月', '2026-10', null) } })

    expect(wrapper.text()).toContain('2026 年 11 月')
    expect(wrapper.find('[data-testid="next-month"]').attributes('disabled')).toBeDefined()
  })

  it('上一個月 / 下一個月 → 交出那個月', async () => {
    const wrapper = mount(MonthStepper, { props: { month: new MonthDto('2026-10', '2026 年 10 月', '2026-09', '2026-11') } })

    await wrapper.find('[data-testid="previous-month"]').trigger('click')
    await wrapper.find('[data-testid="next-month"]').trigger('click')

    expect(wrapper.emitted('change')).toEqual([['2026-09'], ['2026-11']])
  })
})
