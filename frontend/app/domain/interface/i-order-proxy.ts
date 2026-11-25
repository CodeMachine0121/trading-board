import type { FillRequestVo } from '~/domain/models/vo/fill-request-vo'

/** 對板子上的一筆委託回報成交或略過。板子拒絕時拋 BoardRejectedError。 */
export interface IOrderProxy {
  reportFill(orderId: string, fillRequest: FillRequestVo): Promise<void>
  skipOrder(orderId: string): Promise<void>
}
