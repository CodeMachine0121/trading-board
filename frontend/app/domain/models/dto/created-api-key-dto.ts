import type { ApiKeyDto } from '~/domain/models/dto/api-key-dto'

/** 剛建立的金鑰：完整金鑰與提醒只在這一次的回應裡。 */
export class CreatedApiKeyDto {
  constructor(
    readonly apiKey: ApiKeyDto,
    readonly key: string,
    readonly warning: string,
  ) {}
}
