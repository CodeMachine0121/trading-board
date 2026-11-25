import { afterEach, describe, expect, it, vi } from 'vitest'
import { ApiKeyProxy } from '~/infrastructure/proxy/api-key-proxy'

const BASE_URL = '/mock/api/v1'

function stubFetch() {
  const fetchMock = vi.fn()
  vi.stubGlobal('$fetch', fetchMock)

  return fetchMock
}

function newProxy() {
  return new ApiKeyProxy(BASE_URL, async () => 'at')
}

const ACTIVE_WIRE = {
  id: 'k1', memberName: 'stock-scanner', note: null, displayPrefix: 'tbk_abcdefgh', status: 'active',
  createdAt: '2026-11-25T12:00:00Z', lastUsedAt: null, revokedAt: null,
}

afterEach(() => {
  vi.unstubAllGlobals()
})

describe('ApiKeyProxy', () => {
  it('列出：日期字串成了時刻、沒有的成了 null、沒有備註成了空字串', async () => {
    stubFetch().mockResolvedValue([
      { ...ACTIVE_WIRE, id: 'k2', status: 'revoked', note: 'n', lastUsedAt: '2026-11-25T13:00:00Z', revokedAt: '2026-11-25T14:00:00Z' },
      ACTIVE_WIRE,
    ])

    const [revoked, active] = await newProxy().listApiKeys()

    expect(revoked?.status).toBe('revoked')
    expect(revoked?.note).toBe('n')
    expect(revoked?.lastUsedAt?.toISOString()).toBe('2026-11-25T13:00:00.000Z')
    expect(revoked?.revokedAt?.toISOString()).toBe('2026-11-25T14:00:00.000Z')
    expect(active?.status).toBe('active')
    expect(active?.note).toBe('')
    expect(active?.lastUsedAt).toBeNull()
    expect(active?.createdAt.toISOString()).toBe('2026-11-25T12:00:00.000Z')
  })

  it('建立：送成員名稱與備註，帶回只出現一次的完整金鑰', async () => {
    const fetchMock = stubFetch().mockResolvedValue({ ...ACTIVE_WIRE, key: 'tbk_full-key' })

    const created = await newProxy().createApiKey('stock-scanner', 'daily scan')

    expect(created.key).toBe('tbk_full-key')
    expect(created.apiKey.memberName).toBe('stock-scanner')
    expect(fetchMock).toHaveBeenCalledWith(`${BASE_URL}/api-keys`, expect.objectContaining({
      method: 'POST', body: { memberName: 'stock-scanner', note: 'daily scan' },
    }))
  })

  it('撤銷：POST /api-keys/{id}/revoke', async () => {
    const fetchMock = stubFetch().mockResolvedValue({ ...ACTIVE_WIRE, status: 'revoked', revokedAt: '2026-11-25T14:00:00Z' })

    const revoked = await newProxy().revokeApiKey('k 1')

    expect(revoked.status).toBe('revoked')
    expect(fetchMock.mock.calls[0]?.[0]).toBe(`${BASE_URL}/api-keys/k%201/revoke`)
  })
})
