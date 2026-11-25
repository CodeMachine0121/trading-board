/** 登入表單交給 Application 的帳號密碼。 */
export class CredentialsDto {
  constructor(
    readonly username: string,
    readonly password: string,
  ) {}
}
