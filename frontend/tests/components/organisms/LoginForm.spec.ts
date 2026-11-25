import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import LoginForm from '~/components/organisms/LoginForm.vue'
import type { CredentialsDto } from '~/domain/models/dto/credentials-dto'

describe('LoginForm', () => {
  it('送出 → 交出帳號與密碼', async () => {
    const wrapper = mount(LoginForm, { props: { submitting: false, errorMessage: null } })
    await wrapper.find('[data-testid="login-username"]').setValue('owner')
    await wrapper.find('[data-testid="login-password"]').setValue('secret')

    await wrapper.find('form').trigger('submit')

    const [credentials] = wrapper.emitted('submit')?.[0] as [CredentialsDto]
    expect([credentials.username, credentials.password]).toEqual(['owner', 'secret'])
  })

  it('帳密錯誤 → 顯示「帳號或密碼錯誤」', () => {
    const wrapper = mount(LoginForm, { props: { submitting: false, errorMessage: '帳號或密碼錯誤' } })

    expect(wrapper.find('[data-testid="login-error"]').text()).toBe('帳號或密碼錯誤')
  })

  it('沒有錯誤 → 不顯示錯誤', () => {
    expect(mount(LoginForm, { props: { submitting: false, errorMessage: null } }).find('[data-testid="login-error"]').exists())
      .toBe(false)
  })
})
