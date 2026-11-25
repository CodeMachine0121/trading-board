import type { Tone } from '~/domain/models/vo/tone'

/**
 * 一列「名稱：值」。委託、持股與月報都由 Domain 決定要列哪幾列，
 * 元件只負責照順序畫出來——新增一個欄位不必改任何元件。
 */
export class DetailRowDto {
  constructor(
    readonly label: string,
    readonly value: string,
    readonly tone: Tone | null = null,
  ) {}
}
