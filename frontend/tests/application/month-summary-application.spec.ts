import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { MonthSummaryApplication } from '~/application/month-summary-application'
import type { IMonthSummaryProxy } from '~/domain/interface/i-month-summary-proxy'
import { MonthSummaryService } from '~/domain/service/month-summary-service'
import { aMonthSummary } from '../fixtures/entities'

function setUp() {
  const monthSummaryProxy = { fetchMonthSummary: vi.fn<IMonthSummaryProxy['fetchMonthSummary']>() } satisfies IMonthSummaryProxy

  return { monthSummaryProxy, application: new MonthSummaryApplication(new MonthSummaryService(monthSummaryProxy)) }
}

beforeEach(() => {
  vi.useFakeTimers({ toFake: ['Date'] })
  vi.setSystemTime(new Date('2026-11-25T12:30:00Z'))
})

afterEach(() => {
  vi.useRealTimers()
})

describe('MonthSummaryApplication', () => {
  it('今日的預算條 → 查本月（台北 2026-11）', async () => {
    const { monthSummaryProxy, application } = setUp()
    monthSummaryProxy.fetchMonthSummary.mockResolvedValue(aMonthSummary())

    const strip = await application.fetchBudgetStrip()

    expect(monthSummaryProxy.fetchMonthSummary).toHaveBeenCalledWith('2026-11')
    expect(strip?.remainingText).toBe('NT$5,275')
  })

  it('板子還沒有本月的資料 → 沒有預算條', async () => {
    const { monthSummaryProxy, application } = setUp()
    monthSummaryProxy.fetchMonthSummary.mockResolvedValue(null)

    expect(await application.fetchBudgetStrip()).toBeNull()
  })

  it('11 月月報：3 檔出清、2 檔獲利 → 勝率 66.67%', async () => {
    const { monthSummaryProxy, application } = setUp()
    monthSummaryProxy.fetchMonthSummary.mockResolvedValue(aMonthSummary({}, 'strict', { closed: 3, wins: 2, losses: 1 }))

    const page = await application.fetchMonthSummary('2026-11')

    expect(page.month.label).toBe('2026 年 11 月')
    expect(page.summary?.performanceRows.find(row => row.label === '勝率')?.value).toBe('66.67%')
  })

  it('板子第一個月之前的月份 → 這個月沒有資料（summary 是 null）', async () => {
    const { monthSummaryProxy, application } = setUp()
    monthSummaryProxy.fetchMonthSummary.mockResolvedValue(null)

    const page = await application.fetchMonthSummary('2026-10')

    expect(monthSummaryProxy.fetchMonthSummary).toHaveBeenCalledWith('2026-10')
    expect(page.summary).toBeNull()
  })
})
