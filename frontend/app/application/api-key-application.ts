import type { IClipboardProxy } from '~/domain/interface/i-clipboard-proxy'
import type { ApiKeyService } from '~/domain/service/api-key-service'
import type { ApiKeyCreateDto } from '~/domain/models/dto/api-key-create-dto'
import type { ApiKeyDto } from '~/domain/models/dto/api-key-dto'
import type { CreatedApiKeyDto } from '~/domain/models/dto/created-api-key-dto'

/** Application：「API 金鑰」畫面的用例。純 TypeScript，全程只碰 DTO。 */
export class ApiKeyApplication {
  constructor(
    private readonly apiKeyService: ApiKeyService,
    private readonly clipboardProxy: IClipboardProxy,
  ) {}

  async listApiKeys(): Promise<ApiKeyDto[]> {
    return this.apiKeyService.listApiKeys()
  }

  async createApiKey(apiKeyCreateDto: ApiKeyCreateDto): Promise<CreatedApiKeyDto> {
    return this.apiKeyService.createApiKey(apiKeyCreateDto)
  }

  async revokeApiKey(apiKeyId: string): Promise<ApiKeyDto> {
    return this.apiKeyService.revokeApiKey(apiKeyId)
  }

  /** 剪貼簿是瀏覽器的資源，所以經過 proxy；沒有業務規則，不必經過 domain service。 */
  async copyKey(key: string): Promise<void> {
    return this.clipboardProxy.copy(key)
  }
}
