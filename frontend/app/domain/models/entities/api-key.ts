import type { ApiKeyStatus } from '~/domain/models/vo/api-key-status'
import { ApiKeyDomain } from '~/domain/models/domains/api-key-domain'

/** 一位顧問團隊成員的金鑰，從板子讀回來的樣子——完整的金鑰本身從不在裡面。 */
export class ApiKey {
  constructor(
    readonly id: string,
    readonly memberName: string,
    readonly note: string,
    readonly displayPrefix: string,
    readonly status: ApiKeyStatus,
    readonly createdAt: Date,
    readonly lastUsedAt: Date | null,
    readonly revokedAt: Date | null,
  ) {}

  toDomain(): ApiKeyDomain {
    return new ApiKeyDomain(this)
  }
}
