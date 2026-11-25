/**
 * 選月份時畫面要的一切：這個月是哪個月、怎麼念、上一個月與下一個月是哪個月。
 * `next` 是 `null` 代表這已經是本月——未來的月份沒有東西可看。
 */
export class MonthDto {
  constructor(
    readonly value: string,
    readonly label: string,
    readonly previous: string,
    readonly next: string | null,
  ) {}
}
