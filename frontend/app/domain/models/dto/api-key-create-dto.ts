/** 建立金鑰的表單內容：成員名稱與選填的備註。 */
export class ApiKeyCreateDto {
  constructor(
    readonly memberName: string,
    readonly note: string,
  ) {}
}
