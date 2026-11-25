import type { IOrderProxy } from '~/domain/interface/i-order-proxy'
import type { FillRequestVo } from '~/domain/models/vo/fill-request-vo'
import { BoardApiProxy } from '~/infrastructure/proxy/board-api-proxy'

/**
 * Proxy：`POST /orders/{orderId}/fills`、`POST /orders/{orderId}/skip`。
 *
 * 板子會回「更新後的委託」，但這裡不讀它：呼叫端接著會重讀整張交易卡（見 TradeCardService）。
 */
export class OrderProxy extends BoardApiProxy implements IOrderProxy {
  async reportFill(orderId: string, fillRequest: FillRequestVo): Promise<void> {
    await this.requestBoard<unknown>(`/orders/${encodeURIComponent(orderId)}/fills`, {
      method: 'POST',
      body: {
        price: fillRequest.price.toFixed(),
        shares: fillRequest.shares,
        filledAt: fillRequest.filledAt.toISOString(),
        // 沒填手續費就整個不送：板子會依規則估算（需求 §3.4），送 0 的意思是「零元」。
        ...(fillRequest.fee === undefined ? {} : { fee: fillRequest.fee.toFixed() }),
      },
    })
  }

  async skipOrder(orderId: string): Promise<void> {
    await this.requestBoard<unknown>(`/orders/${encodeURIComponent(orderId)}/skip`, { method: 'POST' })
  }
}
