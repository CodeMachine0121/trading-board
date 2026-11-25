import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import ClosedPositionList from '~/components/organisms/ClosedPositionList.vue'
import { ClosedPositionDto } from '~/domain/models/dto/closed-position-dto'

describe('ClosedPositionList', () => {
  it('這個月沒有出清 → 說明', () => {
    expect(mount(ClosedPositionList, { props: { positions: [] } }).find('[data-testid="closed-positions-empty"]').exists())
      .toBe(true)
  })

  it('列出結果與已實現損益', () => {
    const wrapper = mount(ClosedPositionList, {
      props: { positions: [new ClosedPositionDto('6770', '力積電 6770', '獲利', 'profit', '+NT$1,200', [])] },
    })

    expect(wrapper.find('[data-testid="closed-position"]').text()).toContain('力積電 6770')
    expect(wrapper.find('[data-testid="closed-position-outcome"]').text()).toBe('獲利')
    expect(wrapper.find('[data-testid="closed-position-pnl"]').text()).toBe('+NT$1,200')
  })
})
