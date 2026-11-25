import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import AppShell from '~/components/templates/AppShell.vue'

describe('AppShell', () => {
  it('把名稱、動作、導覽與內容放進各自的位置', () => {
    const wrapper = mount(AppShell, {
      slots: { brand: '名稱', actions: '登出', nav: '今日', default: '內容' },
    })

    expect(wrapper.find('.app-shell__brand').text()).toBe('名稱')
    expect(wrapper.find('.app-shell__actions').text()).toBe('登出')
    expect(wrapper.find('nav').text()).toBe('今日')
    expect(wrapper.find('main').text()).toBe('內容')
  })
})
