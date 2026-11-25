import { MockBoard, MockRejection } from '../../../../mock/mock-board'

/**
 * 開發用模擬板子的唯一入口：`/mock/api/v1/**` 照契約 §4 分派到 MockBoard。
 * 只在 `nuxt dev` 時存在；狀態留在這個行程的記憶體裡。
 */
const board = new MockBoard()

export default defineEventHandler(async (event) => {
  const segments = (getRouterParam(event, 'path') ?? '').split('/').filter(segment => segment !== '')
  const method = event.method
  const query = getQuery(event)

  try {
    if (method === 'POST' && segments.join('/') === 'auth/login') {
      const body = await readBody<{ username?: string, password?: string }>(event)

      return board.signIn(body?.username ?? '', body?.password ?? '')
    }

    board.assertAuthorized(getHeader(event, 'authorization'), getHeader(event, 'x-api-key'))

    const [resource, identifier, action] = segments
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

    return { error: { code: error.code, message: error.message } }
  }
})
