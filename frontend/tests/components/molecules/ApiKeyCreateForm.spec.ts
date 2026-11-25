import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import ApiKeyCreateForm from '~/components/molecules/ApiKeyCreateForm.vue'
import type { ApiKeyCreateDto } from '~/domain/models/dto/api-key-create-dto'

describe('ApiKeyCreateForm', () => {
  it('送出 → 交出成員名稱與備註', async () => {
    const wrapper = mount(ApiKeyCreateForm, { props: { submitting: false, errorMessage: null } })
    await wrapper.find('[data-testid="api-key-member-name"]').setValue('stock-scanner')
    await wrapper.find('[data-testid="api-key-note"]').setValue('daily scan')

    await wrapper.find('form').trigger('submit')

    const [dto] = wrapper.emitted('submit')?.[0] as [ApiKeyCreateDto]
    expect([dto.memberName, dto.note]).toEqual(['stock-scanner', 'daily scan'])
  })

  it('成員名稱是必填、最多 40 字', () => {
    const memberName = mount(ApiKeyCreateForm, { props: { submitting: false, errorMessage: null } })
      .find('[data-testid="api-key-member-name"]')

    expect(memberName.attributes('required')).toBeDefined()
    expect(memberName.attributes('maxlength')).toBe('40')
  })

  it.each([
    ['板子的拒絕照原話顯示', '這個成員名稱已有啟用中的金鑰', true],
    ['沒有錯誤就不顯示', null, false],
  ])('%s', (_, errorMessage, shown) => {
    const error = mount(ApiKeyCreateForm, { props: { submitting: false, errorMessage } })
      .find('[data-testid="api-key-create-error"]')

    expect(error.exists()).toBe(shown)
    if (shown) {
      expect(error.text()).toBe(errorMessage)
    }
  })
})
