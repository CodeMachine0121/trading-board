/** 轉址回到 `/auth/callback` 時網址帶回來的兩樣東西。缺的就是空字串。 */
export class SignInReturnDto {
  constructor(
    readonly code: string,
    readonly state: string,
  ) {}
}
