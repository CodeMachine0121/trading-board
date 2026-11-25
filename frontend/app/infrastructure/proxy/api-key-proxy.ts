import type { IApiKeyProxy } from '~/domain/interface/i-api-key-proxy'
import type { ApiKeyStatus } from '~/domain/models/vo/api-key-status'
import { ApiKey } from '~/domain/models/entities/api-key'
import { CreatedApiKey } from '~/domain/models/entities/created-api-key'
import { BoardApiProxy } from '~/infrastructure/proxy/board-api-proxy'

/** 後端回傳的原始 wire 形狀，只存在於本檔內。 */
type ApiKeyWire = {
  id: string
  memberName: string
  note?: string | null
  displayPrefix: string
  status: ApiKeyStatus
  createdAt: string
  lastUsedAt?: string | null
  revokedAt?: string | null
}

type CreatedApiKeyWire = ApiKeyWire & {
  key: string
}

/** Proxy：`GET/POST /api-keys`、`POST /api-keys/{id}/revoke`。 */
export class ApiKeyProxy extends BoardApiProxy implements IApiKeyProxy {
  async listApiKeys(): Promise<ApiKey[]> {
    return (await this.requestBoard<ApiKeyWire[]>('/api-keys')).map(apiKeyWire => this.toApiKey(apiKeyWire))
  }

  async createApiKey(memberName: string, note: string): Promise<CreatedApiKey> {
    const createdWire = await this.requestBoard<CreatedApiKeyWire>('/api-keys', {
      method: 'POST',
      body: { memberName, note },
    })

    return new CreatedApiKey(this.toApiKey(createdWire), createdWire.key)
  }

  async revokeApiKey(apiKeyId: string): Promise<ApiKey> {
    return this.toApiKey(await this.requestBoard<ApiKeyWire>(
      `/api-keys/${encodeURIComponent(apiKeyId)}/revoke`, { method: 'POST' }))
  }

  private toApiKey(apiKeyWire: ApiKeyWire): ApiKey {
    return new ApiKey(
      apiKeyWire.id,
      apiKeyWire.memberName,
      apiKeyWire.note ?? '',
      apiKeyWire.displayPrefix,
      apiKeyWire.status === 'revoked' ? 'revoked' : 'active',
      new Date(apiKeyWire.createdAt),
      apiKeyWire.lastUsedAt ? new Date(apiKeyWire.lastUsedAt) : null,
      apiKeyWire.revokedAt ? new Date(apiKeyWire.revokedAt) : null,
    )
  }
}
