import { describe, expect, it, vi } from 'vitest'
import { ApiKeyApplication } from '~/application/api-key-application'
import { ApiKeyDraftInvalidError } from '~/domain/errors/api-key-draft-invalid-error'
import { BoardRejectedError } from '~/domain/errors/board-rejected-error'
import type { IApiKeyProxy } from '~/domain/interface/i-api-key-proxy'
import type { IClipboardProxy } from '~/domain/interface/i-clipboard-proxy'
import { ApiKeyCreateDto } from '~/domain/models/dto/api-key-create-dto'
import { ApiKey } from '~/domain/models/entities/api-key'
import { CreatedApiKey } from '~/domain/models/entities/created-api-key'
import { ApiKeyService } from '~/domain/service/api-key-service'

/** 台北 2026-11-25 20:00 / 21:30。 */
const CREATED_AT = new Date('2026-11-25T12:00:00Z')
const USED_AT = new Date('2026-11-25T13:30:00Z')

function anApiKey(status: 'active' | 'revoked', lastUsedAt: Date | null = null): ApiKey {
  return new ApiKey('k1', 'stock-scanner', 'daily scan', 'tbk_abcdefgh', status, CREATED_AT, lastUsedAt,
    status === 'revoked' ? USED_AT : null)
}

function setUp() {
  const apiKeyProxy = {
    listApiKeys: vi.fn<IApiKeyProxy['listApiKeys']>(),
    createApiKey: vi.fn<IApiKeyProxy['createApiKey']>(),
    revokeApiKey: vi.fn<IApiKeyProxy['revokeApiKey']>(),
  } satisfies IApiKeyProxy
  const clipboardProxy = { copy: vi.fn<IClipboardProxy['copy']>() } satisfies IClipboardProxy

  return { apiKeyProxy, clipboardProxy, application: new ApiKeyApplication(new ApiKeyService(apiKeyProxy), clipboardProxy) }
}

describe('「API 金鑰」：建立', () => {
  it('建立 stock-scanner → 完整金鑰與「只會顯示這一次」的提醒，清單上是啟用中', async () => {
    const { apiKeyProxy, application } = setUp()
    apiKeyProxy.createApiKey.mockResolvedValue(new CreatedApiKey(anApiKey('active'), 'tbk_full-key'))

    const created = await application.createApiKey(new ApiKeyCreateDto(' stock-scanner ', ' daily scan '))

    expect(apiKeyProxy.createApiKey).toHaveBeenCalledWith('stock-scanner', 'daily scan')
    expect(created.key).toBe('tbk_full-key')
    expect(created.warning).toContain('這把金鑰只會顯示這一次')
    expect(created.apiKey.memberName).toBe('stock-scanner')
    expect(created.apiKey.statusLabel).toBe('啟用中')
  })

  it.each(['', '   '])('成員名稱空白（%j）→ 不送出去', async (memberName) => {
    const { apiKeyProxy, application } = setUp()

    await expect(application.createApiKey(new ApiKeyCreateDto(memberName, '')))
      .rejects.toThrow(new ApiKeyDraftInvalidError('請填寫成員名稱'))
    expect(apiKeyProxy.createApiKey).not.toHaveBeenCalled()
  })

  it('板子說名稱已被使用 → 照板子的話', async () => {
    const { apiKeyProxy, application } = setUp()
    apiKeyProxy.createApiKey.mockRejectedValue(new BoardRejectedError('這個成員名稱已有啟用中的金鑰'))

    await expect(application.createApiKey(new ApiKeyCreateDto('stock-scanner', '')))
      .rejects.toThrow('這個成員名稱已有啟用中的金鑰')
  })
})

describe('「API 金鑰」：清單與撤銷', () => {
  it.each([
    ['啟用中、從未使用', anApiKey('active'), ['啟用中', 'primary', '從未使用', true]],
    ['啟用中、用過', anApiKey('active', USED_AT), ['啟用中', 'primary', '2026/11/25 21:30', true]],
    ['已撤銷', anApiKey('revoked', USED_AT), ['已撤銷', 'neutral', '2026/11/25 21:30', false]],
  ])('%s', async (_, apiKey, [statusLabel, statusTone, lastUsedAtText, canRevoke]) => {
    const { apiKeyProxy, application } = setUp()
    apiKeyProxy.listApiKeys.mockResolvedValue([apiKey])

    const [row] = await application.listApiKeys()

    expect([row?.statusLabel, row?.statusTone, row?.lastUsedAtText, row?.canRevoke])
      .toEqual([statusLabel, statusTone, lastUsedAtText, canRevoke])
    expect(row?.createdAtText).toBe('2026/11/25 20:00')
    expect(row?.displayPrefixText).toBe('tbk_abcdefgh…')
    expect(row?.memberName).toBe('stock-scanner')
    expect(row?.note).toBe('daily scan')
  })

  it('撤銷 → 回已撤銷、沒有動作', async () => {
    const { apiKeyProxy, application } = setUp()
    apiKeyProxy.revokeApiKey.mockResolvedValue(anApiKey('revoked'))

    const revoked = await application.revokeApiKey('k1')

    expect(apiKeyProxy.revokeApiKey).toHaveBeenCalledWith('k1')
    expect([revoked.statusLabel, revoked.canRevoke]).toEqual(['已撤銷', false])
  })

  it('複製金鑰 → 交給剪貼簿', async () => {
    const { clipboardProxy, application } = setUp()

    await application.copyKey('tbk_full-key')

    expect(clipboardProxy.copy).toHaveBeenCalledWith('tbk_full-key')
  })
})
