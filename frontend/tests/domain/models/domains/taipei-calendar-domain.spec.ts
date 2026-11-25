import { describe, expect, it } from 'vitest'
import { TaipeiCalendarDomain } from '~/domain/models/domains/taipei-calendar-domain'

/** 台北 2026-11-25 20:30（UTC 12:30）。 */
const NOV_25_EVENING_TAIPEI = new Date('2026-11-25T12:30:00Z')

describe('TaipeiCalendarDomain：以台北為準的今天與月份', () => {
  it.each([
    ['昨天的場次已經過了', '2026-11-24', true],
    ['今天的場次還沒過', '2026-11-25', false],
    ['明天的場次還沒過', '2026-11-26', false],
  ])('%s', (_, tradeDate, isPast) => {
    expect(new TaipeiCalendarDomain(NOV_25_EVENING_TAIPEI).isPastSession(tradeDate)).toBe(isPast)
  })

  it('UTC 已是 11/25 16:30、台北已是 11/26 00:30 → 11/25 的場次已經過了', () => {
    expect(new TaipeiCalendarDomain(new Date('2026-11-25T16:30:00Z')).isPastSession('2026-11-25')).toBe(true)
  })

  it('現在以台北時間排成表單的時間格式', () => {
    expect(new TaipeiCalendarDomain(NOV_25_EVENING_TAIPEI).nowInputText()).toBe('2026-11-25T20:30')
  })

  it('沒指定月份 → 本月：2026 年 11 月，沒有下一個月', () => {
    const month = new TaipeiCalendarDomain(NOV_25_EVENING_TAIPEI).toMonthDto()

    expect(month).toEqual({ value: '2026-11', label: '2026 年 11 月', previous: '2026-10', next: null })
  })

  it.each([
    ['過去的月份 → 有上一個月與下一個月', '2026-10', '2026-10', '2026-09', '2026-11'],
    ['一月 → 上一個月是去年十二月', '2026-01', '2026-01', '2025-12', '2026-02'],
    ['十二月 → 下一個月是明年一月', '2025-12', '2025-12', '2025-11', '2026-01'],
    ['未來的月份 → 改成本月', '2027-01', '2026-11', '2026-10', null],
    ['格式不對 → 改成本月', 'abc', '2026-11', '2026-10', null],
    ['13 月 → 改成本月', '2026-13', '2026-11', '2026-10', null],
  ])('%s', (_, requested, value, previous, next) => {
    const month = new TaipeiCalendarDomain(NOV_25_EVENING_TAIPEI).toMonthDto(requested)

    expect([month.value, month.previous, month.next]).toEqual([value, previous, next])
  })
})
