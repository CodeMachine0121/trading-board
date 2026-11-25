import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { PositionApplication } from '~/application/position-application'
import type { IPositionProxy } from '~/domain/interface/i-position-proxy'
import { PositionService } from '~/domain/service/position-service'
import { aClosedPosition, anOpenPosition } from '../fixtures/entities'

function setUp() {
  const positionProxy = {
    listOpenPositions: vi.fn<IPositionProxy['listOpenPositions']>(),
    listClosedPositions: vi.fn<IPositionProxy['listClosedPositions']>(),
  } satisfies IPositionProxy

  return { positionProxy, application: new PositionApplication(new PositionService(positionProxy)) }
}

beforeEach(() => {
  vi.useFakeTimers({ toFake: ['Date'] })
  vi.setSystemTime(new Date('2026-11-25T12:30:00Z'))
})

afterEach(() => {
  vi.useRealTimers()
})

describe('PositionApplication', () => {
  it('持有中：聯電 90 股', async () => {
    const { positionProxy, application } = setUp()
    positionProxy.listOpenPositions.mockResolvedValue([anOpenPosition()])

    const positions = await application.listOpenPositions()

    expect(positions.map(position => position.title)).toEqual(['聯電 2303'])
  })

  it('什麼都沒持有 → 空清單', async () => {
    const { positionProxy, application } = setUp()
    positionProxy.listOpenPositions.mockResolvedValue([])

    expect(await application.listOpenPositions()).toEqual([])
  })

  it('沒指定月份 → 查本月出清的，紅賺綠賠', async () => {
    const { positionProxy, application } = setUp()
    positionProxy.listClosedPositions.mockResolvedValue([aClosedPosition('1200'), aClosedPosition('-800')])

    const list = await application.listClosedPositions()

    expect(positionProxy.listClosedPositions).toHaveBeenCalledWith('2026-11')
    expect(list.month.value).toBe('2026-11')
    expect(list.positions.map(position => [position.outcomeLabel, position.tone])).toEqual([['獲利', 'profit'], ['虧損', 'loss']])
  })

  it('指定月份 → 查那個月', async () => {
    const { positionProxy, application } = setUp()
    positionProxy.listClosedPositions.mockResolvedValue([])

    await application.listClosedPositions('2026-09')

    expect(positionProxy.listClosedPositions).toHaveBeenCalledWith('2026-09')
  })
})
