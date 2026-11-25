import { createHash } from 'node:crypto'
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
  it('登入 → 一次性授權碼 → 換到通行憑證與續期憑證；續期會換新、舊的再用就整段登入失效', () => {
    const board = newBoard()
    const verifier = 'v'.repeat(43)
    const challenge = createHash('sha256').update(verifier).digest('base64url')
    const { redirectTo } = board.authorize({
      email: ' Owner@Example.com ', password: 'owner', client_id: 'trading-board-web',
      redirect_uri: 'http://localhost:3000/auth/callback', code_challenge: challenge, code_challenge_method: 'S256', state: 's1',
    })
    const returned = new URL(redirectTo)
    expect(returned.searchParams.get('state')).toBe('s1')

    const first = board.token({ grant_type: 'authorization_code', code: returned.searchParams.get('code'),
      code_verifier: verifier, redirect_uri: 'http://localhost:3000/auth/callback' }, undefined)
    expect(board.assertAuthorized(`Bearer ${first.body.access_token}`, undefined)).toBe('owner')
    expect(first.body.expires_in).toBe(900)

    const second = board.token({ grant_type: 'refresh_token' }, first.refreshToken)
    expect(rejectionOf(() => board.token({ grant_type: 'refresh_token' }, first.refreshToken))?.code).toBe('invalid_grant')
    expect(rejectionOf(() => board.token({ grant_type: 'refresh_token' }, second.refreshToken))?.code).toBe('invalid_grant')
  })

  it('密碼錯 → 401；連錯 5 次 → 429；不明的應用程式 → 400', () => {
    const board = newBoard()
    const attempt = (password: string, clientId = 'trading-board-web') => rejectionOf(() => board.authorize({
      email: 'owner@example.com', password, client_id: clientId, redirect_uri: 'http://localhost:3000/auth/callback',
      code_challenge: 'c'.repeat(43), code_challenge_method: 'S256',
    }))

    expect(attempt('x', 'evil')?.status).toBe(400)
    expect([1, 2, 3, 4, 5].map(() => attempt('x')?.status)).toEqual([401, 401, 401, 401, 401])
    expect(attempt('owner')?.status).toBe(429)
  })

  it('沒帶憑證 → 401；成員金鑰可以讀，但不能管金鑰；撤銷後就不能用', () => {
    const board = newBoard()
    const created = board.createApiKey({ memberName: 'stock-scanner', note: 'n' })

    expect(rejectionOf(() => board.assertAuthorized(undefined, undefined))?.status).toBe(401)
    expect(board.assertAuthorized(undefined, created.key)).toBe('stock-scanner')
    expect(rejectionOf(() => board.assertOwner('stock-scanner'))?.status).toBe(403)
    expect(rejectionOf(() => board.createApiKey({ memberName: 'STOCK-scanner' }))?.status).toBe(409)
    expect(board.revokeApiKey(created.id).status).toBe('revoked')
    expect(rejectionOf(() => board.assertAuthorized(undefined, created.key))?.status).toBe(401)
    expect(board.listApiKeys().map(apiKey => apiKey.memberName)).toEqual(['stock-scanner'])
  })

  it('最新的交易卡是明天的：聯電買進、晶豪科賣出', () => {
    const card = newBoard().latestCard()

    expect(card.tradeDate).toBe('2026-11-26')
    expect(card.publishedBy).toBe('desk-chief')
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
