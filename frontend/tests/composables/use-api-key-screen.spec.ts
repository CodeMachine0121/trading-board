import { describe, expect, it, vi } from 'vitest'
import { ApiKeyApplication } from '~/application/api-key-application'
import { useApiKeyScreen } from '~/composables/use-api-key-screen'
import { BoardRejectedError } from '~/domain/errors/board-rejected-error'
import { ClipboardUnavailableError } from '~/domain/errors/clipboard-unavailable-error'
import type { IApiKeyProxy } from '~/domain/interface/i-api-key-proxy'
import type { IClipboardProxy } from '~/domain/interface/i-clipboard-proxy'
import { ApiKeyCreateDto } from '~/domain/models/dto/api-key-create-dto'
import { ApiKey } from '~/domain/models/entities/api-key'
import { CreatedApiKey } from '~/domain/models/entities/created-api-key'
import { ApiKeyService } from '~/domain/service/api-key-service'

const CREATED_AT = new Date('2026-11-25T12:00:00Z')

function anApiKey(id: string, status: 'active' | 'revoked' = 'active'): ApiKey {
  return new ApiKey(id, `member-${id}`, '', 'tbk_abcdefgh', status, CREATED_AT, null, status === 'revoked' ? CREATED_AT : null)
}

function setUp() {
  const apiKeyProxy = {
    listApiKeys: vi.fn<IApiKeyProxy['listApiKeys']>().mockResolvedValue([anApiKey('k1')]),
    createApiKey: vi.fn<IApiKeyProxy['createApiKey']>(),
    revokeApiKey: vi.fn<IApiKeyProxy['revokeApiKey']>(),
  } satisfies IApiKeyProxy
  const clipboardProxy = { copy: vi.fn<IClipboardProxy['copy']>() } satisfies IClipboardProxy
  const screen = useApiKeyScreen(new ApiKeyApplication(new ApiKeyService(apiKeyProxy), clipboardProxy))

  return { apiKeyProxy, clipboardProxy, screen }
}

describe('useApiKeyScreen：「API 金鑰」畫面的狀態', () => {
  it('建立成功 → 完整金鑰出現、清單重讀', async () => {
    const { apiKeyProxy, screen } = setUp()
    apiKeyProxy.createApiKey.mockResolvedValue(new CreatedApiKey(anApiKey('k2'), 'tbk_full-key'))

    await screen.createApiKey(new ApiKeyCreateDto('member-k2', ''))

    expect(screen.createdKey.value?.key).toBe('tbk_full-key')
    expect(apiKeyProxy.listApiKeys).toHaveBeenCalledTimes(1)
    expect(screen.creating.value).toBe(false)
  })

  it('再建立下一把 → 上一把的完整金鑰與「已複製」先消失；失敗時也不會再出現', async () => {
    const { apiKeyProxy, screen } = setUp()
    apiKeyProxy.createApiKey
      .mockResolvedValueOnce(new CreatedApiKey(anApiKey('k2'), 'tbk_first'))
      .mockRejectedValueOnce(new BoardRejectedError('這個成員名稱已有啟用中的金鑰'))
    await screen.createApiKey(new ApiKeyCreateDto('member-k2', ''))
    await screen.copyKey('tbk_first')

    await screen.createApiKey(new ApiKeyCreateDto('member-k2', ''))

    expect(screen.createdKey.value).toBeNull()
    expect(screen.copied.value).toBe(false)
    expect(screen.createErrorMessage.value).toBe('這個成員名稱已有啟用中的金鑰')
  })

  it('複製成功 → 已複製；剪貼簿不能用 → 說明', async () => {
    const { clipboardProxy, screen } = setUp()

    await screen.copyKey('tbk_key')
    expect(screen.copied.value).toBe(true)

    clipboardProxy.copy.mockRejectedValueOnce(new ClipboardUnavailableError())
    await screen.copyKey('tbk_key')
    expect(screen.copyErrorMessage.value).toBe('無法自動複製，請手動選取金鑰')
  })

  it('撤銷 → 那一列換成已撤銷，其他列不動', async () => {
    const { apiKeyProxy, screen } = setUp()
    apiKeyProxy.listApiKeys.mockResolvedValue([anApiKey('k1'), anApiKey('k2')])
    apiKeyProxy.revokeApiKey.mockResolvedValue(anApiKey('k1', 'revoked'))
    await screen.load()

    await screen.revokeApiKey('k1')

    expect(screen.apiKeys.value?.map(apiKey => apiKey.statusLabel)).toEqual(['已撤銷', '啟用中'])
    expect(screen.revokingId.value).toBeNull()
  })

  it('清單還沒讀到時撤銷成功 → 清單仍是「還沒讀到」', async () => {
    const { apiKeyProxy, screen } = setUp()
    apiKeyProxy.revokeApiKey.mockResolvedValue(anApiKey('k1', 'revoked'))

    await screen.revokeApiKey('k1')

    expect(screen.apiKeys.value).toBeNull()
    expect(screen.revokeErrorMessage.value).toBeNull()
  })

  it('撤銷失敗 → 照板子的話，清單不變', async () => {
    const { apiKeyProxy, screen } = setUp()
    apiKeyProxy.revokeApiKey.mockRejectedValue(new BoardRejectedError('找不到這把金鑰'))

    await screen.revokeApiKey('k9')

    expect(screen.revokeErrorMessage.value).toBe('找不到這把金鑰')
    expect(screen.apiKeys.value).toBeNull()
  })
})
