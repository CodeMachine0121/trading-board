import { DetailRowDto } from '~/domain/models/dto/detail-row-dto'
import { FillDto } from '~/domain/models/dto/fill-dto'
import { OrderDto } from '~/domain/models/dto/order-dto'
import { TradeCardDto } from '~/domain/models/dto/trade-card-dto'

/** 元件測試用的 DTO：元件只看得到 DTO，所以直接造 DTO，不經過 domain。 */

export function anOrderDto(
  id = 'buy-1',
  actionable = true,
  remainingShares = 90,
  side: 'buy' | 'sell' = 'buy',
  fills: FillDto[] = [],
  reason = '底部突破',
): OrderDto {
  return new OrderDto(
    id,
    '2026-11-26',
    side,
    side === 'buy' ? '聯電 2303' : '晶豪科 3006',
    side === 'buy' ? ['月底配置'] : [],
    actionable ? '待成交' : '已成交',
    actionable ? 'info' : 'primary',
    `已成交 ${90 - remainingShares} / 90 股`,
    [new DetailRowDto('限價', '52.5'), new DetailRowDto('停損', '48', 'loss')],
    reason,
    fills,
    actionable,
    remainingShares,
    '52.5',
    String(remainingShares),
    '2026-11-25T20:30',
  )
}

export function aFillDto(): FillDto {
  return new FillDto('fill-1', '52.4', '30 股', 'NT$2', 'NT$1,572', '2026/11/26 09:05')
}

export function aTradeCardDto(
  buyOrders: OrderDto[] = [anOrderDto('buy-1')],
  sellOrders: OrderDto[] = [anOrderDto('sell-1', true, 50, 'sell')],
  note = '',
): TradeCardDto {
  return new TradeCardDto('2026-11-26', '2026/11/26（四）', false, note, buyOrders, sellOrders, '無買進委託', '無賣出委託',
    '發布於 2026/11/25 20:30')
}

export function aNoTradeCardDto(note = '大盤轉弱'): TradeCardDto {
  return new TradeCardDto('2026-11-26', '2026/11/26（四）', true, note, [], [], '無買進委託', '無賣出委託', '')
}
