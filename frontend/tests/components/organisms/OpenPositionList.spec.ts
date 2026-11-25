import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import OpenPositionList from '~/components/organisms/OpenPositionList.vue'
import { DetailRowDto } from '~/domain/models/dto/detail-row-dto'
import { OpenPositionDto } from '~/domain/models/dto/open-position-dto'

describe('OpenPositionList', () => {
  it('沒有持股 → 「目前沒有持股」', () => {
    expect(mount(OpenPositionList, { props: { positions: [] } }).find('[data-testid="open-positions-empty"]').text())
      .toBe('目前沒有持股')
  })

  it('列出每一檔與它的明細', () => {
    const wrapper = mount(OpenPositionList, {
      props: { positions: [new OpenPositionDto('2303', '聯電 2303', [new DetailRowDto('股數', '90 股')])] },
    })

    expect(wrapper.find('[data-testid="open-position"]').text()).toContain('聯電 2303')
    expect(wrapper.find('[data-testid="open-position"]').text()).toContain('90 股')
  })
})
