import { describe, expect, it } from 'vitest'
import { aMonthSummary } from '../../../fixtures/entities'

describe('MonthSummaryDomain：今日畫面的預算條', () => {
  it.each([
    ['嚴格期：月配額 10,000、已投入 4,725', { spent: '4725', remaining: '5275' }, 'strict', 'NT$5,275', 52.75, '嚴格期', 'info'],
    ['月底配置期', { spent: '4725', remaining: '5275' }, 'monthEnd', 'NT$5,275', 52.75, '月底配置期', 'warning'],
    ['全部花完', { spent: '10000', remaining: '0' }, 'strict', 'NT$0', 0, '嚴格期', 'info'],
    ['上月結轉 2,000、賣出回收 3,000、還沒花', { carriedOver: '2000', recycledProceeds: '3000', spent: '0', remaining: '15000' }, 'strict', 'NT$15,000', 100, '嚴格期', 'info'],
  ] as const)('%s', (_, amounts, mode, remainingText, percent, modeLabel, modeTone) => {
    const strip = aMonthSummary(amounts, mode).toDomain().toBudgetStripDto()

    expect(strip.remainingText).toBe(remainingText)
    expect(strip.allocationText).toBe('NT$10,000')
    expect(strip.remainingPercent).toBe(percent)
    expect(strip.modeLabel).toBe(modeLabel)
    expect(strip.modeTone).toBe(modeTone)
  })

  it.each([
    ['剩餘比可用還多（資料異常）→ 進度條封頂 100', { remaining: '20000', spent: '0' }, 100],
    ['剩餘是負的（資料異常）→ 進度條歸零', { remaining: '-10', spent: '10010' }, 0],
    ['可用的錢是 0 → 進度條是空的', { allocation: '0', remaining: '0', spent: '0' }, 0],
    ['上月結轉 2,000、已投入 6,000 → 分母含結轉：6,000 ÷ 12,000 = 50', { carriedOver: '2000', spent: '6000', remaining: '6000' }, 50],
    ['剩餘 1/3 → 四捨五入到兩位', { allocation: '3', remaining: '1', spent: '2' }, 33.33],
  ])('%s', (_, amounts, percent) => {
    expect(aMonthSummary(amounts).toDomain().toBudgetStripDto().remainingPercent).toBe(percent)
  })
})

describe('MonthSummaryDomain：月報', () => {
  it('列出六項預算明細與期間', () => {
    const summary = aMonthSummary(
      { allocation: '10000', carriedOver: '0', recycledProceeds: '3300', spent: '9800', remaining: '3500' }, 'monthEnd',
    ).toDomain().toDto()

    expect(summary.budgetRows.map(row => [row.label, row.value])).toEqual([
      ['月配額', 'NT$10,000'],
      ['上月結轉', 'NT$0'],
      ['賣出回收', 'NT$3,300'],
      ['已投入', 'NT$9,800'],
      ['剩餘', 'NT$3,500'],
    ])
    expect(summary.modeLabel).toBe('月底配置期')
  })

  it('3 檔出清、2 檔獲利 → 勝率 66.67%，並列出獲利與虧損檔數', () => {
    const summary = aMonthSummary({}, 'strict', { realizedPnl: '1200', closed: 3, wins: 2, losses: 1, returnPct: '8.5' })
      .toDomain().toDto()

    expect(summary.performanceRows.map(row => [row.label, row.value])).toEqual([
      ['已出清', '3 檔'],
      ['獲利', '2 檔'],
      ['虧損', '1 檔'],
      ['勝率', '66.67%'],
      ['已實現報酬率', '8.50%'],
    ])
    expect(summary.realizedPnlText).toBe('+NT$1,200')
    expect(summary.realizedPnlTone).toBe('profit')
  })

  it.each([
    ['當月沒有出清 → 勝率與報酬率顯示「—」', 0, '5'],
    ['有出清但板子沒給報酬率 → 報酬率顯示「—」', 2, null],
  ])('%s', (_, closed, returnPct) => {
    const rows = aMonthSummary({}, 'strict', { closed, wins: closed, returnPct }).toDomain().toDto().performanceRows

    expect(rows.find(row => row.label === '已實現報酬率')?.value).toBe('—')
    expect(rows.find(row => row.label === '勝率')?.value).toBe(closed === 0 ? '—' : '100.00%')
  })

  it.each([
    ['已實現虧損 → 綠色', '-800', '-NT$800', 'loss'],
    ['損益兩平 → 中性色', '0', 'NT$0', 'neutral'],
  ])('%s', (_, realizedPnl, text, tone) => {
    const summary = aMonthSummary({}, 'strict', { realizedPnl }).toDomain().toDto()

    expect(summary.realizedPnlText).toBe(text)
    expect(summary.realizedPnlTone).toBe(tone)
  })
})
