import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import BudgetStrip from '~/components/molecules/BudgetStrip.vue'
import { BudgetStripDto } from '~/domain/models/dto/budget-strip-dto'

describe('BudgetStrip', () => {
  it('剩餘 / 月配額、進度、期間', () => {
    const wrapper = mount(BudgetStrip, { props: { budget: new BudgetStripDto('NT$5,275', 'NT$10,000', 52.75, '嚴格期', 'info') } })

    expect(wrapper.find('[data-testid="budget-remaining"]').text()).toBe('NT$5,275')
    expect(wrapper.text()).toContain('月配額 NT$10,000')
    expect(wrapper.find('[role="progressbar"]').attributes('aria-valuenow')).toBe('52.75')
    expect(wrapper.find('[data-testid="budget-mode"]').text()).toBe('嚴格期')
  })
})
