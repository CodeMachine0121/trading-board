/**
 * 回報成交表單原樣交給 Application 的內容——全部是使用者打的字，還沒驗證。
 * `filledAtText` 是 `YYYY-MM-DDTHH:mm`，以台北時間解讀。
 */
export class FillReportDto {
  constructor(
    readonly orderId: string,
    readonly tradeDate: string,
    readonly remainingShares: number,
    readonly priceText: string,
    readonly sharesText: string,
    readonly feeText: string,
    readonly filledAtText: string,
  ) {}
}
