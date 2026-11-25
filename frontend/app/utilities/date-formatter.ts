/**
 * 日期的顯示格式。純技術性轉換，不碰領域規則。
 *
 * 交易日是「日曆上的一天」（YYYY-MM-DD），不是一個時刻，所以星期幾用 UTC 算，
 * 不受瀏覽器所在時區影響；時刻（成交時間、發布時間）一律以台北時間顯示。
 */

const WEEKDAY_LABELS = ['日', '一', '二', '三', '四', '五', '六']

// 拆成數字再自己組：各語系的完整格式會夾帶不換行空白之類的字元，畫面與測試都不穩定。
const TAIPEI_DATE_TIME_PARTS = new Intl.DateTimeFormat('en-CA', {
  timeZone: 'Asia/Taipei',
  year: 'numeric',
  month: '2-digit',
  day: '2-digit',
  hour: '2-digit',
  minute: '2-digit',
  hourCycle: 'h23',
})

/** `2026-11-26` → `2026/11/26（四）`。 */
export function formatCalendarDate(calendarDate: string): string {
  const [year = 0, month = 1, day = 1] = calendarDate.split('-').map(Number)
  const weekday = new Date(Date.UTC(year, month - 1, day)).getUTCDay()

  return `${calendarDate.replaceAll('-', '/')}（${WEEKDAY_LABELS[weekday]}）`
}

/** 一個時刻以台北時間顯示：`2026/11/26 09:05`。 */
export function formatTaipeiDateTime(moment: Date): string {
  const parts = Object.fromEntries(TAIPEI_DATE_TIME_PARTS.formatToParts(moment).map(part => [part.type, part.value]))

  return `${parts.year}/${parts.month}/${parts.day} ${parts.hour}:${parts.minute}`
}
