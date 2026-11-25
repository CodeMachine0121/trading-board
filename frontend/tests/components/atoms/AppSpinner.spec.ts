import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import AppSpinner from '~/components/atoms/AppSpinner.vue'

describe('AppSpinner', () => {
  it('給螢幕閱讀器的說明', () => {
    expect(mount(AppSpinner).text()).toBe('載入中')
  })
})
