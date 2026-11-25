/** 一筆已回報的成交，全部排成畫面上的字。 */
export class FillDto {
  constructor(
    readonly id: string,
    readonly priceText: string,
    readonly sharesText: string,
    readonly feeText: string,
    readonly amountText: string,
    readonly filledAtText: string,
  ) {}
}
