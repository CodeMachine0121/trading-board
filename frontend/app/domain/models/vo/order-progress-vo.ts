import type { Tone } from '~/domain/models/vo/tone'

/** 一筆委託現在走到哪裡：畫面上的狀態、語氣、還剩幾股、還能不能回報或略過。 */
export class OrderProgressVo {
  constructor(
    readonly statusLabel: string,
    readonly statusTone: Tone,
    readonly remainingShares: number,
    readonly actionable: boolean,
  ) {}
}
