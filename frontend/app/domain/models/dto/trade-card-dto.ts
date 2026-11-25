import type { OrderDto } from '~/domain/models/dto/order-dto'

/** 一張交易卡在畫面上的全部樣子。 */
export class TradeCardDto {
  constructor(
    readonly tradeDate: string,
    readonly tradeDateLabel: string,
    readonly isNoTrade: boolean,
    readonly note: string,
    readonly buyOrders: readonly OrderDto[],
    readonly sellOrders: readonly OrderDto[],
    readonly emptyBuyText: string,
    readonly emptySellText: string,
    readonly publishedAtText: string,
  ) {}
}
