import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import MonthSummaryPanel from '~/components/organisms/MonthSummaryPanel.vue'
import { DetailRowDto } from '~/domain/models/dto/detail-row-dto'
import { MonthSummaryDto } from '~/domain/models/dto/month-summary-dto'

describe('MonthSummaryPanel', () => {
  it('那個月沒有資料 → 「這個月沒有資料」', () => {
    expect(mount(MonthSummaryPanel, { props: { summary: null } }).find('[data-testid="month-summary-empty"]').text())
      .toBe('這個月沒有資料')
  })

  it('列出期間、預算明細、已實現損益與績效', () => {
    const wrapper = mount(MonthSummaryPanel, {
      props: {
        summary: new MonthSummaryDto('嚴格期', 'info', [new DetailRowDto('月配額', 'NT$10,000')], '+NT$1,200', 'profit',
          [new DetailRowDto('勝率', '66.67%')]),
      },
    })

    expect(wrapper.text()).toContain('嚴格期')
    expect(wrapper.text()).toContain('NT$10,000')
    expect(wrapper.find('[data-testid="month-realized-pnl"]').text()).toBe('+NT$1,200')
    expect(wrapper.text()).toContain('66.67%')
  })
})
