/** 這個網頁在板子那邊登記的 OAuth 公開用戶端：用戶端代號，與登入後轉回來的那一頁。 */
export class OAuthClientVo {
  constructor(
    readonly clientId: string,
    readonly redirectUri: string,
  ) {}
}
