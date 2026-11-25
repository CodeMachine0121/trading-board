import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import CreatedApiKeyPanel from '~/components/molecules/CreatedApiKeyPanel.vue'
import { ApiKeyDto } from '~/domain/models/dto/api-key-dto'
import { CreatedApiKeyDto } from '~/domain/models/dto/created-api-key-dto'

const CREATED = new CreatedApiKeyDto(
  new ApiKeyDto('k1', 'stock-scanner', '', 'tbk_abcdefgh…', '啟用中', 'primary', '2026/11/25 20:00', '從未使用', true),
  'tbk_full-key',
  '這把金鑰只會顯示這一次，請現在複製並交給這位成員',
)

function mountPanel(copied: boolean, copyErrorMessage: string | null = null) {
  return mount(CreatedApiKeyPanel, { props: { createdKey: CREATED, copied, copyErrorMessage } })
}

describe('CreatedApiKeyPanel', () => {
  it('完整金鑰與只會顯示一次的提醒', () => {
    const wrapper = mountPanel(false)

    expect(wrapper.find('[data-testid="created-api-key-value"]').text()).toBe('tbk_full-key')
    expect(wrapper.text()).toContain('這把金鑰只會顯示這一次')
    expect(wrapper.text()).toContain('stock-scanner')
  })

  it('按複製 → 交出完整金鑰', async () => {
    const wrapper = mountPanel(false)

    await wrapper.find('[data-testid="created-api-key-copy"]').trigger('click')

    expect(wrapper.emitted('copy')?.[0]).toEqual(['tbk_full-key'])
  })

  it.each([
    [false, '複製金鑰'],
    [true, '已複製'],
  ])('複製過了嗎 %s → 按鈕寫「%s」', (copied, label) => {
    expect(mountPanel(copied).find('[data-testid="created-api-key-copy"]').text()).toBe(label)
  })

  it('剪貼簿不能用 → 說明', () => {
    expect(mountPanel(false, '無法自動複製，請手動選取金鑰').text()).toContain('無法自動複製，請手動選取金鑰')
  })
})
