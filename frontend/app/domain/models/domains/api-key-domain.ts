import type { ApiKey } from '~/domain/models/entities/api-key'
import { ApiKeyDto } from '~/domain/models/dto/api-key-dto'
import { formatTaipeiDateTime } from '~/utilities/date-formatter'

/** 一把成員金鑰在清單上怎麼呈現、還能不能撤銷。 */
export class ApiKeyDomain {
  constructor(private readonly apiKey: ApiKey) {}

  toDto(): ApiKeyDto {
    const isActive = this.apiKey.status === 'active'

    return new ApiKeyDto(
      this.apiKey.id,
      this.apiKey.memberName,
      this.apiKey.note,
      `${this.apiKey.displayPrefix}…`,
      isActive ? '啟用中' : '已撤銷',
      isActive ? 'primary' : 'neutral',
      formatTaipeiDateTime(this.apiKey.createdAt),
      this.apiKey.lastUsedAt === null ? '從未使用' : formatTaipeiDateTime(this.apiKey.lastUsedAt),
      isActive,
    )
  }
}
