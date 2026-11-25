import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import ApiKeyList from '~/components/organisms/ApiKeyList.vue'
import { ApiKeyDto } from '~/domain/models/dto/api-key-dto'

function aKeyDto(id: string, canRevoke: boolean, note = ''): ApiKeyDto {
  return new ApiKeyDto(id, `member-${id}`, note, 'tbk_abcdefgh…', canRevoke ? '啟用中' : '已撤銷',
    canRevoke ? 'primary' : 'neutral', '2026/11/25 20:00', '從未使用', canRevoke)
}

function mountList(apiKeys: ApiKeyDto[]) {
  return mount(ApiKeyList, { props: { apiKeys, revokingId: null } })
}

describe('ApiKeyList', () => {
  it('沒有金鑰 → 「還沒有任何金鑰」', () => {
    expect(mountList([]).find('[data-testid="api-key-list-empty"]').text()).toBe('還沒有任何金鑰')
  })

  it('每一列：成員名稱、狀態、最後使用、備註', () => {
    const wrapper = mountList([aKeyDto('k1', true, 'daily scan')])

    const row = wrapper.find('[data-testid="api-key-row"]')
    expect(row.text()).toContain('member-k1')
    expect(row.text()).toContain('daily scan')
    expect(row.text()).toContain('tbk_abcdefgh…')
    expect(row.find('[data-testid="api-key-status"]').text()).toBe('啟用中')
    expect(row.find('[data-testid="api-key-last-used"]').text()).toBe('從未使用')
  })

  it('按撤銷再確定 → 撤銷那一把', async () => {
    const wrapper = mountList([aKeyDto('k1', true)])

    await wrapper.find('[data-testid="api-key-revoke"]').trigger('click')
    await wrapper.find('[data-testid="api-key-revoke-confirm"]').trigger('click')

    expect(wrapper.emitted('revoke')?.[0]).toEqual(['k1'])
    expect(wrapper.find('[data-testid="api-key-revoke-confirm"]').exists()).toBe(false)
  })

  it('按撤銷再取消 → 什麼都沒發生，回到原本的按鈕', async () => {
    const wrapper = mountList([aKeyDto('k1', true)])

    await wrapper.find('[data-testid="api-key-revoke"]').trigger('click')
    await wrapper.find('[data-testid="api-key-revoke-cancel"]').trigger('click')

    expect(wrapper.emitted('revoke')).toBeUndefined()
    expect(wrapper.find('[data-testid="api-key-revoke"]').exists()).toBe(true)
  })

  it('已撤銷的金鑰沒有任何動作', () => {
    const wrapper = mountList([aKeyDto('k2', false)])

    expect(wrapper.find('[data-testid="api-key-revoke"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="api-key-status"]').text()).toBe('已撤銷')
  })
})
