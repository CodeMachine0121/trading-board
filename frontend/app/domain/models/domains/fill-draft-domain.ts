import Decimal from 'decimal.js'
import { FillDraftInvalidError } from '~/domain/errors/fill-draft-invalid-error'
import { FillRequestVo } from '~/domain/models/vo/fill-request-vo'

const DECIMAL_PATTERN = /^\d+(\.\d+)?$/
const WHOLE_NUMBER_PATTERN = /^\d+$/
const DATE_TIME_INPUT_PATTERN = /^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}$/
/** 成交時間欄位以台北時間解讀，不管瀏覽器在哪個時區。 */
const TAIPEI_OFFSET = '+08:00'

/**
 * 主人在回報成交表單裡打的字（需求 §3.4）。
 *
 * 在送給板子之前先擋掉一定會被拒絕的內容：價格要大於 0、股數是 1 到剩餘股數之間的整數、
 * 手續費可以不填（由板子估算），填了就不能是負的。
 */
export class FillDraftDomain {
  constructor(
    private readonly priceText: string,
    private readonly sharesText: string,
    private readonly feeText: string,
    private readonly filledAtText: string,
    private readonly remainingShares: number,
  ) {}

  /** 驗證通過就是一份可以送出去的回報；否則以使用者看得懂的理由拒絕。 */
  toRequestVo(): FillRequestVo {
    const priceText = this.priceText.trim()
    if (!DECIMAL_PATTERN.test(priceText) || new Decimal(priceText).lessThanOrEqualTo(0)) {
      throw new FillDraftInvalidError('價格須大於 0')
    }

    const sharesText = this.sharesText.trim()
    const shares = Number(sharesText)
    if (!WHOLE_NUMBER_PATTERN.test(sharesText) || shares < 1) {
      throw new FillDraftInvalidError('股數須為 1 以上的整數')
    }
    if (shares > this.remainingShares) {
      throw new FillDraftInvalidError(`股數不可超過剩餘 ${this.remainingShares} 股`)
    }

    const feeText = this.feeText.trim()
    if (feeText !== '' && !DECIMAL_PATTERN.test(feeText)) {
      throw new FillDraftInvalidError('手續費須為 0 以上的數字')
    }

    const filledAt = new Date(`${this.filledAtText}:00${TAIPEI_OFFSET}`)
    if (!DATE_TIME_INPUT_PATTERN.test(this.filledAtText) || Number.isNaN(filledAt.getTime())) {
      throw new FillDraftInvalidError('請填寫成交時間')
    }

    return new FillRequestVo(
      new Decimal(priceText),
      shares,
      feeText === '' ? undefined : new Decimal(feeText),
      filledAt,
    )
  }
}
