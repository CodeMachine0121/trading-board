import type { ApiKey } from '~/domain/models/entities/api-key'
import type { CreatedApiKey } from '~/domain/models/entities/created-api-key'

/** 板子上的成員金鑰（只有登入的主人能管）。 */
export interface IApiKeyProxy {
  /** 新的在前。 */
  listApiKeys(): Promise<ApiKey[]>
  createApiKey(memberName: string, note: string): Promise<CreatedApiKey>
  revokeApiKey(apiKeyId: string): Promise<ApiKey>
}
