import { describe, expect, it } from 'vitest'
import { FillDraftInvalidError } from '~/domain/errors/fill-draft-invalid-error'
import { FillDraftDomain } from '~/domain/models/domains/fill-draft-domain'

const VALID_FILLED_AT = '2026-11-26T09:05'

describe('FillDraftDomain：回報成交表單在送出前的把關', () => {
  it('剩 30 股時回報剛好 30 股 → 通過，價格與股數原樣帶出', () => {
    const request = new FillDraftDomain('52', '30', '', VALID_FILLED_AT, 30).toRequestVo()

    expect(request.price.toFixed()).toBe('52')
    expect(request.shares).toBe(30)
  })

  it('成交時間以台北時間解讀', () => {
    const request = new FillDraftDomain('52.5', '1', '', VALID_FILLED_AT, 30).toRequestVo()

    expect(request.filledAt.toISOString()).toBe('2026-11-26T01:05:00.000Z')
  })

  it.each([
    ['手續費留空 → 不帶手續費（由板子估算）', '', undefined],
    ['手續費只有空白 → 視同留空', '   ', undefined],
    ['手續費填 0 → 就是零元', '0', '0'],
    ['手續費填 20 → 20', '20', '20'],
  ])('%s', (_, feeText, expectedFee) => {
    const request = new FillDraftDomain('52.5', '10', feeText, VALID_FILLED_AT, 30).toRequestVo()

    expect(request.fee?.toFixed()).toBe(expectedFee)
  })

  it.each([
    ['剩 30 股卻回報 31 股', '52', '31', '', VALID_FILLED_AT, '股數不可超過剩餘 30 股'],
    ['股數 0', '52', '0', '', VALID_FILLED_AT, '股數須為 1 以上的整數'],
    ['股數有小數', '52', '1.5', '', VALID_FILLED_AT, '股數須為 1 以上的整數'],
    ['股數空白', '52', '', '', VALID_FILLED_AT, '股數須為 1 以上的整數'],
    ['價格 0', '0', '10', '', VALID_FILLED_AT, '價格須大於 0'],
    ['價格不是數字', 'abc', '10', '', VALID_FILLED_AT, '價格須大於 0'],
    ['價格是負的', '-1', '10', '', VALID_FILLED_AT, '價格須大於 0'],
    ['手續費是負的', '52', '10', '-1', VALID_FILLED_AT, '手續費須為 0 以上的數字'],
    ['成交時間空白', '52', '10', '', '', '請填寫成交時間'],
    ['成交時間格式不對', '52', '10', '', '2026/11/26 09:05', '請填寫成交時間'],
    ['成交時間不存在（13 月）', '52', '10', '', '2026-13-26T09:05', '請填寫成交時間'],
  ])('%s → 拒絕：%s', (_, price, shares, fee, filledAt, message) => {
    const draft = new FillDraftDomain(price, shares, fee, filledAt, 30)

    expect(() => draft.toRequestVo()).toThrow(new FillDraftInvalidError(message))
  })
})
