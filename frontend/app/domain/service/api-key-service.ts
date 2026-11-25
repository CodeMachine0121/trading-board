import type { IApiKeyProxy } from '~/domain/interface/i-api-key-proxy'
import type { ApiKeyCreateDto } from '~/domain/models/dto/api-key-create-dto'
import type { ApiKeyDto } from '~/domain/models/dto/api-key-dto'
import type { CreatedApiKeyDto } from '~/domain/models/dto/created-api-key-dto'
import { ApiKeyDraftInvalidError } from '~/domain/errors/api-key-draft-invalid-error'

/** 主人管理顧問團隊成員的金鑰：列出、建立、撤銷。 */
export class ApiKeyService {
  constructor(private readonly apiKeyProxy: IApiKeyProxy) {}

  async listApiKeys(): Promise<ApiKeyDto[]> {
    return (await this.apiKeyProxy.listApiKeys()).map(apiKey => apiKey.toDomain().toDto())
  }

  /** 成員名稱空白就不送出去；其餘規則（長度、重名）由板子說了算，照它的話顯示。 */
  async createApiKey(apiKeyCreateDto: ApiKeyCreateDto): Promise<CreatedApiKeyDto> {
    const memberName = apiKeyCreateDto.memberName.trim()
    if (memberName === '') {
      throw new ApiKeyDraftInvalidError('請填寫成員名稱')
    }

    return (await this.apiKeyProxy.createApiKey(memberName, apiKeyCreateDto.note.trim())).toDto()
  }

  async revokeApiKey(apiKeyId: string): Promise<ApiKeyDto> {
    return (await this.apiKeyProxy.revokeApiKey(apiKeyId)).toDomain().toDto()
  }
}
