import type { ApiKey } from '~/domain/models/entities/api-key'
import { CreatedApiKeyDto } from '~/domain/models/dto/created-api-key-dto'

/** 剛建立的金鑰：清單上的那一列，加上只會出現這一次的完整金鑰。 */
export class CreatedApiKey {
  constructor(
    readonly apiKey: ApiKey,
    readonly key: string,
  ) {}

  toDto(): CreatedApiKeyDto {
    return new CreatedApiKeyDto(this.apiKey.toDomain().toDto(), this.key, '這把金鑰只會顯示這一次，請現在複製並交給這位成員')
  }
}
