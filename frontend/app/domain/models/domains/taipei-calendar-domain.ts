import { MonthDto } from '~/domain/models/dto/month-dto'

const TAIPEI_TIME_ZONE = 'Asia/Taipei'
const MONTH_PATTERN = /^(\d{4})-(0[1-9]|1[0-2])$/
const MONTHS_PER_YEAR = 12

/** 把一個時刻拆成台北時間的年月日時分。en-CA 讓日期部分固定是數字。 */
const TAIPEI_PARTS_FORMAT = new Intl.DateTimeFormat('en-CA', {
  timeZone: TAIPEI_TIME_ZONE,
  year: 'numeric',
  month: '2-digit',
  day: '2-digit',
  hour: '2-digit',
  minute: '2-digit',
  hourCycle: 'h23',
})

/**
 * 台北的日曆。所有「今天」「本月」都以 Asia/Taipei 為準（需求 §3.1），
 * 與使用者的瀏覽器在哪個時區無關——人在國外看盤，交易日仍然是台股的那一天。
 */
export class TaipeiCalendarDomain {
  private readonly today: string
  private readonly nowInput: string

  constructor(now: Date) {
    const parts = Object.fromEntries(TAIPEI_PARTS_FORMAT.formatToParts(now).map(part => [part.type, part.value]))
    this.today = `${parts.year}-${parts.month}-${parts.day}`
    this.nowInput = `${this.today}T${parts.hour}:${parts.minute}`
  }

  /** 交易日在今天之前：那一場已經收盤了。 */
  isPastSession(tradeDate: string): boolean {
    return tradeDate < this.today
  }

  /** 現在，排成回報成交表單時間欄位的格式（`YYYY-MM-DDTHH:mm`，台北時間）。 */
  nowInputText(): string {
    return this.nowInput
  }

  /**
   * 描述一個月份給選月份的元件用。沒給、格式不對或在未來的月份，一律改成本月：
   * 未來沒有東西可看，而打錯的網址不該變成一頁空白。
   */
  toMonthDto(requestedMonth?: string): MonthDto {
    const currentMonth = this.today.slice(0, 7)
    const month = requestedMonth !== undefined && MONTH_PATTERN.test(requestedMonth) && requestedMonth <= currentMonth
      ? requestedMonth
      : currentMonth
    const [year = 0, monthNumber = 1] = month.split('-').map(Number)
    const monthIndex = year * MONTHS_PER_YEAR + (monthNumber - 1)
    const monthAt = (index: number): string =>
      `${Math.floor(index / MONTHS_PER_YEAR)}-${String((index % MONTHS_PER_YEAR) + 1).padStart(2, '0')}`

    return new MonthDto(
      month,
      `${year} 年 ${monthNumber} 月`,
      monthAt(monthIndex - 1),
      month === currentMonth ? null : monthAt(monthIndex + 1),
    )
  }
}
