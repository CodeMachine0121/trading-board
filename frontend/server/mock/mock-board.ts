import Decimal from 'decimal.js'

/**
 * 開發用的模擬板子：在網路層照契約 §4 回答，讓前端在後端完成之前就能整套操作。
 *
 * 它只在 `nuxt dev` 時存在（正式映像檔是 `nuxt generate` 出來的靜態檔，沒有伺服器），
 * 而且只活在記憶體裡——重開 dev server 就回到種子資料。前端程式碼裡沒有任何假實作：
 * 換成真的後端只是改 NUXT_PUBLIC_API_BASE_URL（見 data-access.md）。
 *
 * 規則照需求 §3 做到「看起來像真的」為止：手續費估算、預算不可為負、平均成本、
 * 已實現損益、月預算結轉。它不是後端的參考實作。
 */

export const MOCK_USERNAME = 'owner'
export const MOCK_PASSWORD = 'owner'
const MOCK_TOKEN_PREFIX = 'mock-token-'
const SESSION_HOURS = 12

const MONTHLY_ALLOCATION = new Decimal(10000)
const BROKER_FEE_RATE = new Decimal('0.001425')
const FEE_DISCOUNT = new Decimal('0.6')
const MINIMUM_FEE = new Decimal(1)
const TRANSACTION_TAX_RATE = new Decimal('0.003')
const MONTH_END_START_DAY = 22
const MILLISECONDS_PER_DAY = 86_400_000

const TAIPEI_DATE_FORMAT = new Intl.DateTimeFormat('en-CA', {
  timeZone: 'Asia/Taipei',
  year: 'numeric',
  month: '2-digit',
  day: '2-digit',
})

/** 拒絕：狀態碼與契約的錯誤內容。 */
export class MockRejection extends Error {
  constructor(readonly status: number, readonly code: string, message: string) {
    super(message)
  }
}

class MockFill {
  constructor(
    readonly id: string,
    readonly price: Decimal,
    readonly shares: number,
    readonly fee: Decimal,
    readonly amount: Decimal,
    readonly filledAt: Date,
  ) {}
}

class MockOrder {
  readonly fills: MockFill[] = []
  skipped = false

  constructor(
    readonly id: string,
    readonly side: 'buy' | 'sell',
    readonly symbol: string,
    readonly name: string,
    readonly limitPrice: Decimal,
    readonly shares: number,
    readonly reason: string,
    readonly maxChasePrice: Decimal = limitPrice,
    readonly stopLossPrice: Decimal = limitPrice,
    readonly takeProfitPrice: Decimal = limitPrice,
    readonly isMonthEndAllocation = false,
    readonly exitReason = 'takeProfit',
  ) {}

  filledShares(): number {
    return this.fills.reduce((total, fill) => total + fill.shares, 0)
  }
}

class MockCard {
  constructor(
    readonly tradeDate: string,
    readonly status: 'trade' | 'noTrade',
    readonly note: string,
    readonly orders: MockOrder[],
    readonly publishedAt: Date,
  ) {}
}

class MockPositionState {
  shares = 0
  totalCost = new Decimal(0)
  realizedPnl = new Decimal(0)

  constructor(
    readonly symbol: string,
    readonly name: string,
    readonly openedDate: string,
    public stopLossPrice: Decimal,
    public takeProfitPrice: Decimal,
  ) {}
}

class MockClosedPosition {
  constructor(
    readonly state: MockPositionState,
    readonly closedDate: string,
  ) {}
}

export class MockBoard {
  private readonly cards = new Map<string, MockCard>()
  private sequence = 0

  constructor(private readonly now: () => Date = () => new Date()) {
    this.seed()
  }

  // ── 身分 ─────────────────────────────────────────────

  signIn(username: string, password: string) {
    if (username !== MOCK_USERNAME || password !== MOCK_PASSWORD) {
      throw new MockRejection(401, 'invalidCredentials', '帳號或密碼錯誤')
    }

    return {
      accessToken: `${MOCK_TOKEN_PREFIX}${this.nextId()}`,
      expiresAt: new Date(this.now().getTime() + SESSION_HOURS * 3_600_000).toISOString(),
    }
  }

  assertAuthorized(authorization: string | undefined, apiKey: string | undefined): void {
    if (apiKey !== undefined && apiKey !== '') {
      return
    }
    if (authorization === undefined || !authorization.startsWith(`Bearer ${MOCK_TOKEN_PREFIX}`)) {
      throw new MockRejection(401, 'unauthorized', '請先登入')
    }
  }

  // ── 交易卡 ───────────────────────────────────────────

  latestCard() {
    const latest = [...this.cards.keys()].sort().at(-1)
    if (latest === undefined) {
      throw new MockRejection(404, 'notFound', '尚未發布任何交易卡')
    }

    return this.card(latest)
  }

  card(tradeDate: string) {
    const card = this.cards.get(tradeDate)
    if (card === undefined) {
      throw new MockRejection(404, 'notFound', '找不到這一天的交易卡')
    }

    return {
      tradeDate: card.tradeDate,
      status: card.status,
      note: card.note,
      buyOrders: card.orders.filter(order => order.side === 'buy').map(order => ({
        ...this.orderWire(order),
        maxChasePrice: order.maxChasePrice.toFixed(),
        estimatedAmount: order.limitPrice.times(order.shares).toFixed(),
        stopLossPrice: order.stopLossPrice.toFixed(),
        takeProfitPrice: order.takeProfitPrice.toFixed(),
        isMonthEndAllocation: order.isMonthEndAllocation,
      })),
      sellOrders: card.orders.filter(order => order.side === 'sell').map(order => ({
        ...this.orderWire(order),
        exitReason: order.exitReason,
      })),
      publishedAt: card.publishedAt.toISOString(),
    }
  }

  listCards(month: string) {
    return [...this.cards.values()]
      .filter(card => card.tradeDate.startsWith(month))
      .sort((left, right) => right.tradeDate.localeCompare(left.tradeDate))
      .map(card => ({
        tradeDate: card.tradeDate,
        status: card.status,
        buyCount: card.orders.filter(order => order.side === 'buy').length,
        sellCount: card.orders.filter(order => order.side === 'sell').length,
        filledOrderCount: card.orders.filter(order => order.filledShares() === order.shares).length,
      }))
  }

  // ── 委託 ─────────────────────────────────────────────

  reportFill(orderId: string, body: { price?: unknown, shares?: unknown, fee?: unknown, filledAt?: unknown }) {
    const order = this.order(orderId)
    const shares = Number(body.shares)
    const filledAt = new Date(String(body.filledAt))
    if (order.skipped) {
      throw new MockRejection(409, 'orderSkipped', '已略過的委託不能回報成交')
    }
    if (typeof body.price !== 'string' || !/^\d+(\.\d+)?$/.test(body.price) || new Decimal(body.price).lte(0)) {
      throw new MockRejection(400, 'validation', '價格須大於 0')
    }
    if (!Number.isInteger(shares) || shares < 1) {
      throw new MockRejection(400, 'validation', '股數須為 1 以上的整數')
    }
    if (order.filledShares() + shares > order.shares) {
      throw new MockRejection(400, 'validation', `成交股數超過委託剩餘的 ${order.shares - order.filledShares()} 股`)
    }
    if (Number.isNaN(filledAt.getTime())) {
      throw new MockRejection(400, 'validation', '成交時間格式不正確')
    }

    const price = new Decimal(body.price)
    const amount = price.times(shares)
    const fee = typeof body.fee === 'string' && body.fee !== '' ? new Decimal(body.fee) : this.estimateFee(order.side, amount)
    const fill = new MockFill(this.nextId(), price, shares, fee, amount, filledAt)

    if (order.side === 'buy') {
      const month = this.taipeiDate(filledAt).slice(0, 7)
      const remaining = this.budget(month).remaining
      if (remaining.minus(amount).minus(fee).isNegative()) {
        throw new MockRejection(409, 'budgetExceeded', `超出本月剩餘預算（剩餘 NT$${remaining.toFixed()}）`)
      }
    }
    else {
      const held = this.positions().open.find(position => position.symbol === order.symbol)?.shares ?? 0
      if (held < shares) {
        throw new MockRejection(409, 'insufficientShares', `持股不足（目前持有 ${held} 股）`)
      }
    }

    order.fills.push(fill)

    return this.orderWire(order)
  }

  skipOrder(orderId: string) {
    const order = this.order(orderId)
    if (order.filledShares() === order.shares) {
      throw new MockRejection(409, 'orderFilled', '已全部成交的委託不能略過')
    }
    order.skipped = true

    return this.orderWire(order)
  }

  // ── 持股與月報 ───────────────────────────────────────

  listPositions(status: string, month: string | undefined) {
    const { open, closed } = this.positions()
    const today = this.taipeiDate(this.now())

    if (status === 'closed') {
      return closed
        .filter(position => month === undefined || position.closedDate.startsWith(month))
        .map(position => this.positionWire(position.state, position.closedDate, position.closedDate))
    }

    return open.map(state => this.positionWire(state, null, today))
  }

  monthSummary(month: string) {
    const firstMonth = this.firstMonth()
    if (!/^\d{4}-\d{2}$/.test(month) || firstMonth === null || month < firstMonth) {
      throw new MockRejection(404, 'notFound', '這個月沒有資料')
    }

    const budget = this.budget(month)
    const sells = this.sellResults().filter(result => result.month === month)
    const realizedPnl = sells.reduce((total, result) => total.plus(result.pnl), new Decimal(0))
    const costBasis = sells.reduce((total, result) => total.plus(result.cost), new Decimal(0))
    const closed = this.positions().closed.filter(position => position.closedDate.startsWith(month))
    const winCount = closed.filter(position => position.state.realizedPnl.isPositive() && !position.state.realizedPnl.isZero()).length
    const day = Number(this.taipeiDate(this.now()).slice(8, 10))
    const isCurrentMonth = this.taipeiDate(this.now()).startsWith(month)

    return {
      month,
      allocation: MONTHLY_ALLOCATION.toFixed(),
      carriedOver: budget.carriedOver.toFixed(),
      recycledProceeds: budget.recycledProceeds.toFixed(),
      spent: budget.spent.toFixed(),
      remaining: budget.remaining.toFixed(),
      mode: !isCurrentMonth || day >= MONTH_END_START_DAY ? 'monthEnd' : 'strict',
      realizedPnl: realizedPnl.toFixed(),
      closedPositionCount: closed.length,
      winCount,
      lossCount: closed.length - winCount,
      winRate: closed.length === 0 ? '0' : new Decimal(winCount).dividedBy(closed.length).toDecimalPlaces(4).toFixed(),
      realizedReturnPct: costBasis.isZero() ? null : realizedPnl.dividedBy(costBasis).times(100).toDecimalPlaces(2).toFixed(),
    }
  }

  // ── 內部 ─────────────────────────────────────────────

  private budget(month: string) {
    const firstMonth = this.firstMonth() ?? month
    const fills = this.allFills()
    let carriedOver = new Decimal(0)
    let result = { carriedOver, recycledProceeds: new Decimal(0), spent: new Decimal(0), remaining: MONTHLY_ALLOCATION }

    for (let cursor = firstMonth; cursor <= month; cursor = this.nextMonth(cursor)) {
      const inMonth = fills.filter(entry => this.taipeiDate(entry.fill.filledAt).startsWith(cursor))
      const recycledProceeds = inMonth
        .filter(entry => entry.order.side === 'sell')
        .reduce((total, entry) => total.plus(entry.fill.amount).minus(entry.fill.fee), new Decimal(0))
      const spent = inMonth
        .filter(entry => entry.order.side === 'buy')
        .reduce((total, entry) => total.plus(entry.fill.amount).plus(entry.fill.fee), new Decimal(0))
      const remaining = MONTHLY_ALLOCATION.plus(carriedOver).plus(recycledProceeds).minus(spent)
      result = { carriedOver, recycledProceeds, spent, remaining }
      carriedOver = remaining
    }

    return result
  }

  private positions() {
    const open = new Map<string, MockPositionState>()
    const closed: MockClosedPosition[] = []

    for (const { order, fill } of this.allFills()) {
      const fillDate = this.taipeiDate(fill.filledAt)
      if (order.side === 'buy') {
        const state = open.get(order.symbol)
          ?? new MockPositionState(order.symbol, order.name, fillDate, order.stopLossPrice, order.takeProfitPrice)
        state.shares += fill.shares
        state.totalCost = state.totalCost.plus(fill.amount).plus(fill.fee)
        state.stopLossPrice = order.stopLossPrice
        state.takeProfitPrice = order.takeProfitPrice
        open.set(order.symbol, state)
        continue
      }

      const state = open.get(order.symbol)
      if (state === undefined) {
        continue
      }
      const cost = state.totalCost.dividedBy(state.shares).times(fill.shares)
      state.realizedPnl = state.realizedPnl.plus(fill.amount.minus(fill.fee).minus(cost))
      state.totalCost = state.totalCost.minus(cost)
      state.shares -= fill.shares
      if (state.shares === 0) {
        closed.push(new MockClosedPosition(state, fillDate))
        open.delete(order.symbol)
      }
    }

    return { open: [...open.values()], closed }
  }

  private sellResults() {
    const averageCosts = new Map<string, { shares: number, cost: Decimal }>()
    const results: { month: string, pnl: Decimal, cost: Decimal }[] = []

    for (const { order, fill } of this.allFills()) {
      const held = averageCosts.get(order.symbol) ?? { shares: 0, cost: new Decimal(0) }
      if (order.side === 'buy') {
        averageCosts.set(order.symbol, { shares: held.shares + fill.shares, cost: held.cost.plus(fill.amount).plus(fill.fee) })
        continue
      }
      const cost = held.cost.dividedBy(held.shares).times(fill.shares)
      averageCosts.set(order.symbol, { shares: held.shares - fill.shares, cost: held.cost.minus(cost) })
      results.push({ month: this.taipeiDate(fill.filledAt).slice(0, 7), pnl: fill.amount.minus(fill.fee).minus(cost), cost })
    }

    return results
  }

  private positionWire(state: MockPositionState, closedDate: string | null, untilDate: string) {
    return {
      symbol: state.symbol,
      name: state.name,
      shares: state.shares,
      averageCost: state.shares === 0 ? '0' : state.totalCost.dividedBy(state.shares).toDecimalPlaces(2).toFixed(),
      totalCost: state.totalCost.toDecimalPlaces(0).toFixed(),
      stopLossPrice: state.stopLossPrice.toFixed(),
      takeProfitPrice: state.takeProfitPrice.toFixed(),
      openedDate: state.openedDate,
      closedDate,
      daysHeld: Math.round((Date.parse(untilDate) - Date.parse(state.openedDate)) / MILLISECONDS_PER_DAY),
      realizedPnl: state.realizedPnl.toDecimalPlaces(0).toFixed(),
    }
  }

  private orderWire(order: MockOrder) {
    const filledShares = order.filledShares()
    const status = order.skipped
      ? 'skipped'
      : filledShares === order.shares ? 'filled' : filledShares > 0 ? 'partiallyFilled' : 'pending'

    return {
      id: order.id,
      symbol: order.symbol,
      name: order.name,
      limitPrice: order.limitPrice.toFixed(),
      shares: order.shares,
      reason: order.reason,
      status,
      filledShares,
      fills: order.fills.map(fill => ({
        id: fill.id,
        price: fill.price.toFixed(),
        shares: fill.shares,
        fee: fill.fee.toFixed(),
        amount: fill.amount.toFixed(),
        filledAt: fill.filledAt.toISOString(),
      })),
    }
  }

  private order(orderId: string): MockOrder {
    for (const card of this.cards.values()) {
      const order = card.orders.find(candidate => candidate.id === orderId)
      if (order !== undefined) {
        return order
      }
    }

    throw new MockRejection(404, 'notFound', '找不到這筆委託')
  }

  private allFills() {
    return [...this.cards.values()]
      .flatMap(card => card.orders.flatMap(order => order.fills.map(fill => ({ order, fill }))))
      .sort((left, right) => left.fill.filledAt.getTime() - right.fill.filledAt.getTime())
  }

  private firstMonth(): string | null {
    const first = [...this.cards.keys()].sort().at(0)

    return first === undefined ? null : first.slice(0, 7)
  }

  private estimateFee(side: 'buy' | 'sell', amount: Decimal): Decimal {
    const brokerFee = Decimal.max(
      amount.times(BROKER_FEE_RATE).times(FEE_DISCOUNT).toDecimalPlaces(0, Decimal.ROUND_HALF_UP),
      MINIMUM_FEE,
    )

    return side === 'buy'
      ? brokerFee
      : brokerFee.plus(amount.times(TRANSACTION_TAX_RATE).toDecimalPlaces(0, Decimal.ROUND_HALF_UP))
  }

  private nextMonth(month: string): string {
    const [year = 0, monthNumber = 1] = month.split('-').map(Number)

    return monthNumber === 12 ? `${year + 1}-01` : `${year}-${String(monthNumber + 1).padStart(2, '0')}`
  }

  private taipeiDate(moment: Date): string {
    return TAIPEI_DATE_FORMAT.format(moment)
  }

  private dayOffset(days: number): string {
    return this.taipeiDate(new Date(this.now().getTime() + days * MILLISECONDS_PER_DAY))
  }

  private nextId(): string {
    this.sequence += 1

    return `mock-${String(this.sequence).padStart(4, '0')}`
  }

  /** 種子：一檔已出清（獲利）、一檔持有中、一張今日不買、一筆過期的委託、明天的交易卡。 */
  private seed(): void {
    const at = (days: number, time: string) => new Date(`${this.dayOffset(days)}T${time}:00+08:00`)
    const fillOf = (order: MockOrder, price: string, shares: number, filledAt: Date) => {
      const amount = new Decimal(price).times(shares)
      order.fills.push(new MockFill(this.nextId(), new Decimal(price), shares, this.estimateFee(order.side, amount), amount, filledAt))
    }

    const psmcBuy = new MockOrder(this.nextId(), 'buy', '6770', '力積電', new Decimal('20'), 100,
      '底部盤整後放量突破月線。', new Decimal('20.5'), new Decimal('18.5'), new Decimal('23'))
    fillOf(psmcBuy, '20', 100, at(-6, '09:05'))
    this.cards.set(this.dayOffset(-6), new MockCard(this.dayOffset(-6), 'trade', '', [psmcBuy], at(-7, '20:30')))

    const psmcSell = new MockOrder(this.nextId(), 'sell', '6770', '力積電', new Decimal('22.5'), 100,
      '接近前高壓力，先落袋。', undefined, undefined, undefined, false, 'takeProfit')
    fillOf(psmcSell, '22.5', 100, at(-5, '10:12'))
    this.cards.set(this.dayOffset(-5), new MockCard(this.dayOffset(-5), 'trade', '', [psmcSell], at(-6, '20:30')))

    const esmtBuy = new MockOrder(this.nextId(), 'buy', '3006', '晶豪科', new Decimal('60'), 50,
      '記憶體報價回升，頸線突破。', new Decimal('61.5'), new Decimal('56'), new Decimal('68'))
    fillOf(esmtBuy, '60', 50, at(-3, '09:20'))
    const winbondBuy = new MockOrder(this.nextId(), 'buy', '2344', '華邦電', new Decimal('45'), 60,
      '量縮回測支撐。', new Decimal('46'), new Decimal('42'), new Decimal('52'))
    this.cards.set(this.dayOffset(-3), new MockCard(this.dayOffset(-3), 'trade', '', [esmtBuy, winbondBuy], at(-4, '20:30')))

    this.cards.set(this.dayOffset(-2), new MockCard(this.dayOffset(-2), 'noTrade', '大盤轉弱，觀望一天。', [], at(-3, '20:30')))

    const umcBuy = new MockOrder(this.nextId(), 'buy', '2303', '聯電', new Decimal('52.5'), 90,
      '底部打底三個月後帶量突破，回測不破。', new Decimal('54'), new Decimal('48'), new Decimal('60'))
    const nanyaBuy = new MockOrder(this.nextId(), 'buy', '2408', '南亞科', new Decimal('48'), 40,
      '月底資金配置，等回測 5 日線。', new Decimal('49'), new Decimal('44'), new Decimal('55'), true)
    const esmtSell = new MockOrder(this.nextId(), 'sell', '3006', '晶豪科', new Decimal('68'), 50,
      '到達停利價。', undefined, undefined, undefined, false, 'takeProfit')
    this.cards.set(this.dayOffset(1), new MockCard(this.dayOffset(1), 'trade', '', [umcBuy, nanyaBuy, esmtSell], at(0, '20:30')))
  }
}
