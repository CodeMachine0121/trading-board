import type { H3Event } from 'h3'
import { MockBoard, MockRejection } from '../../../../mock/mock-board'
import type { MockTokenGrant } from '../../../../mock/mock-board'

/**
 * 開發用模擬板子的唯一入口：`/mock/api/v1/**` 照契約分派到 MockBoard。
 * 只在 `nuxt dev` 時存在；狀態留在這個行程的記憶體裡。
 */
const board = new MockBoard()

/** 續期憑證的 cookie：和真的板子一樣 HttpOnly、SameSite=Strict、只送到 oauth 端點（dev 是 http，所以不設 Secure）。 */
const REFRESH_COOKIE = 'trading_board_refresh_token'
const REFRESH_COOKIE_PATH = '/mock/api/v1/oauth'

function keepRefreshCookie(event: H3Event, grant: MockTokenGrant) {
  setCookie(event, REFRESH_COOKIE, grant.refreshToken, {
    httpOnly: true, sameSite: 'strict', path: REFRESH_COOKIE_PATH, maxAge: 30 * 86_400,
  })
  setResponseHeader(event, 'Cache-Control', 'no-store')

  return grant.body
}

function forgetRefreshCookie(event: H3Event) {
  deleteCookie(event, REFRESH_COOKIE, { httpOnly: true, sameSite: 'strict', path: REFRESH_COOKIE_PATH })
}

export default defineEventHandler(async (event) => {
  const segments = (getRouterParam(event, 'path') ?? '').split('/').filter(segment => segment !== '')
  const method = event.method
  const query = getQuery(event)
  const route = segments.join('/')

  try {
    if (method === 'POST' && route === 'oauth/authorize') {
      return board.authorize(await readBody(event) ?? {})
    }
    if (method === 'POST' && route === 'oauth/token') {
      try {
        return keepRefreshCookie(event, board.token(await readBody(event) ?? {}, getCookie(event, REFRESH_COOKIE)))
      }
      catch (error: unknown) {
        forgetRefreshCookie(event)
        throw error
      }
    }
    if (method === 'POST' && route === 'oauth/revoke') {
      board.revoke(getCookie(event, REFRESH_COOKIE))
      forgetRefreshCookie(event)

      return ''
    }

    const caller = board.assertAuthorized(getHeader(event, 'authorization'), getHeader(event, 'x-api-key'))

    const [resource, identifier, action] = segments
    if (method === 'GET' && route === 'me') {
      board.assertOwner(caller)
      return board.me()
    }
    if (resource === 'api-keys') {
      board.assertOwner(caller)
      if (method === 'GET' && identifier === undefined) {
        return board.listApiKeys()
      }
      if (method === 'POST' && identifier === undefined) {
        setResponseStatus(event, 201)
        return board.createApiKey(await readBody(event) ?? {})
      }
      if (method === 'POST' && identifier !== undefined && action === 'revoke') {
        return board.revokeApiKey(identifier)
      }
    }
    if (method === 'GET' && resource === 'cards' && identifier === undefined) {
      return board.listCards(String(query.month ?? ''))
    }
    if (method === 'GET' && resource === 'cards' && identifier === 'latest') {
      return board.latestCard()
    }
    if (method === 'GET' && resource === 'cards' && identifier !== undefined) {
      return board.card(identifier)
    }
    if (method === 'POST' && resource === 'orders' && identifier !== undefined && action === 'fills') {
      return board.reportFill(identifier, await readBody(event))
    }
    if (method === 'POST' && resource === 'orders' && identifier !== undefined && action === 'skip') {
      return board.skipOrder(identifier)
    }
    if (method === 'GET' && resource === 'positions') {
      return board.listPositions(String(query.status ?? 'open'), typeof query.month === 'string' ? query.month : undefined)
    }
    if (method === 'GET' && resource === 'months' && identifier !== undefined && action === 'summary') {
      return board.monthSummary(identifier)
    }

    throw new MockRejection(404, 'notFound', '模擬板子沒有這條路')
  }
  catch (error: unknown) {
    if (!(error instanceof MockRejection)) {
      throw error
    }
    setResponseStatus(event, error.status)

    return error.oauth
      ? { error: error.code, error_description: error.message }
      : { error: { code: error.code, message: error.message } }
  }
})
