import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import AppAlert from '~/components/atoms/AppAlert.vue'

describe('AppAlert', () => {
  it.each([
    ['錯誤 → 立刻念出來', 'error', 'alert'],
    ['一般提示 → 有空再念', 'info', 'status'],
  ] as const)('%s', (_, variant, role) => {
    expect(mount(AppAlert, { props: { variant }, slots: { default: '訊息' } }).attributes('role')).toBe(role)
  })
})
