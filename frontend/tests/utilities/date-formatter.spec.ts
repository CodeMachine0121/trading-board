import { describe, expect, it } from 'vitest'
import { formatCalendarDate, formatTaipeiDateTime } from '~/utilities/date-formatter'

describe('date-formatter', () => {
  it.each([
    ['2026-11-26', '2026/11/26（四）'],
    ['2026-11-29', '2026/11/29（日）'],
  ])('formatCalendarDate(%s) → %s', (date, text) => {
    expect(formatCalendarDate(date)).toBe(text)
  })

  it('formatTaipeiDateTime 以台北時間顯示', () => {
    expect(formatTaipeiDateTime(new Date('2026-11-26T01:05:00Z'))).toBe('2026/11/26 09:05')
  })
})
