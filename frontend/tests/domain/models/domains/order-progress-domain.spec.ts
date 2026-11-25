import { describe, expect, it } from 'vitest'
import { OrderProgressDomain } from '~/domain/models/domains/order-progress-domain'
import type { OrderStatus } from '~/domain/models/vo/order-status'

describe('OrderProgressDomain：一筆委託走到哪裡、還能不能操作', () => {
  it.each<[string, OrderStatus, number, boolean, string, string, number, boolean]>([
    ['今天的場次、一股都還沒成交 → 待成交，可操作', 'pending', 0, false, '待成交', 'info', 90, true],
    ['成交 30/90 → 部分成交，可操作', 'partiallyFilled', 30, false, '部分成交', 'warning', 60, true],
    ['成交 90/90 → 已成交，不可操作', 'filled', 90, false, '已成交', 'primary', 0, false],
    ['主人略過 → 已略過，不可操作', 'skipped', 0, false, '已略過', 'neutral', 90, false],
    ['昨天的場次、一股都沒成交 → 已過期，不可操作', 'pending', 0, true, '已過期', 'neutral', 90, false],
    ['昨天的場次、成交 30/90 → 仍是部分成交，可補回報', 'partiallyFilled', 30, true, '部分成交', 'warning', 60, true],
    ['昨天的場次、全部成交 → 已成交，不是過期', 'filled', 90, true, '已成交', 'primary', 0, false],
    ['略過之前已成交一部分 → 仍是已略過', 'skipped', 30, false, '已略過', 'neutral', 60, false],
  ])('%s', (_, status, filledShares, sessionIsPast, label, tone, remaining, actionable) => {
    const progress = new OrderProgressDomain(status, 90, filledShares, sessionIsPast).toVo()

    expect(progress.statusLabel).toBe(label)
    expect(progress.statusTone).toBe(tone)
    expect(progress.remainingShares).toBe(remaining)
    expect(progress.actionable).toBe(actionable)
  })

  it.each([
    ['回報的成交股數超過委託 → 夾回委託股數', 120, 0],
    ['成交股數是負的 → 夾回 0', -5, 90],
  ])('%s', (_, filledShares, remaining) => {
    expect(new OrderProgressDomain('partiallyFilled', 90, filledShares, false).toVo().remainingShares).toBe(remaining)
  })
})
