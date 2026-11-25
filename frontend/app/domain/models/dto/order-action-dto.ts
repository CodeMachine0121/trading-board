/** 對一筆委託做事（例如略過）時要帶的：是哪一筆、在哪一張交易卡上。 */
export class OrderActionDto {
  constructor(
    readonly orderId: string,
    readonly tradeDate: string,
  ) {}
}
