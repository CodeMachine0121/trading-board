/** 登入表單交給 Application 的電子郵件與密碼。 */
export class CredentialsDto {
  constructor(
    readonly email: string,
    readonly password: string,
  ) {}
}
