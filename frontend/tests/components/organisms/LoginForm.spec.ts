import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import LoginForm from '~/components/organisms/LoginForm.vue'
import type { CredentialsDto } from '~/domain/models/dto/credentials-dto'

describe('LoginForm', () => {
  it('送出 → 交出電子郵件與密碼', async () => {
    const wrapper = mount(LoginForm, { props: { submitting: false, errorMessage: null } })
    await wrapper.find('[data-testid="login-email"]').setValue('owner@example.com')
    await wrapper.find('[data-testid="login-password"]').setValue('secret')

    await wrapper.find('form').trigger('submit')

    const [credentials] = wrapper.emitted('submit')?.[0] as [CredentialsDto]
    expect([credentials.email, credentials.password]).toEqual(['owner@example.com', 'secret'])
  })

  it.each(['電子郵件或密碼錯誤', '嘗試次數過多，請 15 分鐘後再試'])('板子拒絕 → 顯示「%s」', (message) => {
    const wrapper = mount(LoginForm, { props: { submitting: false, errorMessage: message } })

    expect(wrapper.find('[data-testid="login-error"]').text()).toBe(message)
  })

  it('電子郵件欄位是 email 型別（手機跳出對的鍵盤）', () => {
    const wrapper = mount(LoginForm, { props: { submitting: false, errorMessage: null } })

    expect(wrapper.find('[data-testid="login-email"]').attributes('type')).toBe('email')
  })

  it('沒有錯誤 → 不顯示錯誤', () => {
    expect(mount(LoginForm, { props: { submitting: false, errorMessage: null } }).find('[data-testid="login-error"]').exists())
      .toBe(false)
  })
})
