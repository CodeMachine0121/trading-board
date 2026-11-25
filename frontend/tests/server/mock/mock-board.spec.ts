import { describe, expect, it } from 'vitest'
import { MockBoard, MockRejection } from '../../../server/mock/mock-board'

/** 台北 2026-11-25 20:30。 */
const NOW = new Date('2026-11-25T12:30:00Z')

function newBoard() {
  return new MockBoard(() => NOW)
}

function rejectionOf(action: () => unknown): MockRejection | null {
  try {
    action()

    return null
  }
  catch (error: unknown) {
    return error instanceof MockRejection ? error : null
  }
}

describe('開發用模擬板子（只在 nuxt dev 時存在）', () => {
  it('owner / owner 登入得到憑證；密碼錯誤 401', () => {
    const board = newBoard()

    expect(board.signIn('owner', 'owner').accessToken).toMatch(/^mock-token-/)
    expect(rejectionOf(() => board.signIn('owner', 'x'))?.status).toBe(401)
  })

  it('沒帶憑證 → 401；帶 API key 也可以讀', () => {
    const board = newBoard()

    expect(rejectionOf(() => board.assertAuthorized(undefined, undefined))?.status).toBe(401)
    expect(rejectionOf(() => board.assertAuthorized(undefined, 'agent-key'))).toBeNull()
  })

  it('最新的交易卡是明天的：聯電買進、晶豪科賣出', () => {
    const card = newBoard().latestCard()

    expect(card.tradeDate).toBe('2026-11-26')
    expect(card.buyOrders.map(order => order.name)).toContain('聯電')
    expect(card.sellOrders.map(order => order.name)).toEqual(['晶豪科'])
  })

  it('回報聯電 90 股 → 已成交，本月剩餘預算減少（含估算手續費）', () => {
    const board = newBoard()
    const umc = board.latestCard().buyOrders.find(order => order.symbol === '2303')
    const before = board.monthSummary('2026-11').remaining

    const order = board.reportFill(umc?.id ?? '', { price: '52.5', shares: 90, filledAt: '2026-11-26T01:05:00Z' })

    expect(order.status).toBe('filled')
    expect(order.fills[0]?.fee).toBe('4')
    expect(Number(board.monthSummary('2026-11').remaining)).toBe(Number(before) - 4725 - 4)
  })

  it('超過剩餘股數 → 400；超過預算 → 409', () => {
    const board = newBoard()
    const umc = board.latestCard().buyOrders.find(order => order.symbol === '2303')

    expect(rejectionOf(() => board.reportFill(umc?.id ?? '', { price: '52.5', shares: 91, filledAt: NOW.toISOString() }))?.status).toBe(400)
    expect(rejectionOf(() => board.reportFill(umc?.id ?? '', { price: '5250', shares: 90, filledAt: NOW.toISOString() }))?.status).toBe(409)
  })

  it('持股與月報：晶豪科持有中、力積電已出清獲利', () => {
    const board = newBoard()

    expect(board.listPositions('open', undefined).map(position => position.symbol)).toEqual(['3006'])
    const [closed] = board.listPositions('closed', '2026-11')
    expect([closed?.symbol, Number(closed?.realizedPnl) > 0]).toEqual(['6770', true])
    expect(board.monthSummary('2026-11').winCount).toBe(1)
    expect(rejectionOf(() => board.monthSummary('2026-10'))?.status).toBe(404)
  })
})
