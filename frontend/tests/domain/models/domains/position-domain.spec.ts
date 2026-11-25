import { describe, expect, it } from 'vitest'
import { aClosedPosition, anOpenPosition } from '../../../fixtures/entities'

describe('PositionDomain：持有中的股票', () => {
  it('聯電 90 股、平均成本 52.58 → 列出股數、成本、防守價、進場日與持有天數', () => {
    const position = anOpenPosition().toDomain().toOpenDto()

    expect(position.title).toBe('聯電 2303')
    expect(position.detailRows.map(row => [row.label, row.value])).toEqual([
      ['股數', '90 股'],
      ['平均成本', '52.58'],
      ['總成本', 'NT$4,732'],
      ['停損', '48'],
      ['停利', '60'],
      ['進場日', '2026/11/20（五）'],
      ['持有天數', '5 天'],
    ])
  })

  it('沒有停損停利價 → 顯示「—」', () => {
    const rows = anOpenPosition(null, null).toDomain().toOpenDto().detailRows

    expect(rows.find(row => row.label === '停損')?.value).toBe('—')
    expect(rows.find(row => row.label === '停利')?.value).toBe('—')
  })
})

describe('PositionDomain：已出清的股票（紅賺綠賠）', () => {
  it.each([
    ['已實現 +1,200 → 紅色、獲利', '1200', 'profit', '獲利', '+NT$1,200'],
    ['已實現 −800 → 綠色、虧損', '-800', 'loss', '虧損', '-NT$800'],
    ['損益兩平 → 中性色，但算虧損', '0', 'neutral', '虧損', 'NT$0'],
  ])('%s', (_, realizedPnl, tone, outcome, text) => {
    const position = aClosedPosition(realizedPnl).toDomain().toClosedDto()

    expect(position.tone).toBe(tone)
    expect(position.outcomeLabel).toBe(outcome)
    expect(position.realizedPnlText).toBe(text)
  })

  it.each([
    ['有出場日 → 列出進場日、出場日、持有天數', '2026-11-24', '2026/11/24（二）'],
    ['板子沒給出場日 → 顯示「—」', null, '—'],
  ])('%s', (_, closedDate, closedText) => {
    const rows = aClosedPosition('1200', closedDate).toDomain().toClosedDto().detailRows

    expect(rows.map(row => [row.label, row.value])).toEqual([
      ['進場日', '2026/11/18（三）'],
      ['出場日', closedText],
      ['持有天數', '6 天'],
    ])
  })
})
