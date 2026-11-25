import { OwnerDto } from '~/domain/models/dto/owner-dto'

/** 登入中的主人。 */
export class Owner {
  constructor(readonly email: string) {}

  toDto(): OwnerDto {
    return new OwnerDto(this.email)
  }
}
