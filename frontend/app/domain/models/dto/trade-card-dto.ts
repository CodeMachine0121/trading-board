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
    /** 「發布者：desk-chief」；不知道是誰發布的就是空字串，畫面不顯示這一行。 */
    readonly publisherText: string = '',
  ) {}
}
