import type { Tone } from '~/domain/models/vo/tone'

/**
 * 今日畫面頂端那一條本月預算。
 * `remainingPercent` 是 0–100：本月可用的錢（月配額＋上月結轉＋賣出回收）還剩幾成。
 */
export class BudgetStripDto {
  constructor(
    readonly remainingText: string,
    readonly allocationText: string,
    readonly remainingPercent: number,
    readonly modeLabel: string,
    readonly modeTone: Tone,
  ) {}
}
