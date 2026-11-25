import { ref, type Ref } from 'vue'
import type { TradeCardApplication } from '~/application/trade-card-application'
import type { FillReportDto } from '~/domain/models/dto/fill-report-dto'
import type { OrderActionDto } from '~/domain/models/dto/order-action-dto'
import type { TradeCardDto } from '~/domain/models/dto/trade-card-dto'
import { failureMessageOf } from '~/composables/failure-message'

/**
 * 交易卡上「回報成交」與「略過」的那一段畫面狀態：哪一筆正在送、哪一筆失敗了、為什麼。
 *
 * 今日與歷史的卡片明細都用同一張卡，所以這一段只寫一次。成功時把卡換成板子現在的樣子，
 * 再通知呼叫端（今日畫面要順便重讀本月預算）；失敗時卡維持原樣，理由掛在那一筆委託上。
 */
export function useTradeCardActions(
  tradeCardApplication: TradeCardApplication,
  card: Ref<TradeCardDto | null>,
  onCardChanged: () => void = () => {},
) {
  const busyOrderId = ref<string | null>(null)
  const failedOrderId = ref<string | null>(null)
  const failureMessage = ref<string | null>(null)

  async function runOrderAction(orderId: string, action: () => Promise<TradeCardDto | null>): Promise<void> {
    busyOrderId.value = orderId
    failedOrderId.value = null
    failureMessage.value = null

    try {
      const updatedCard = await action()
      if (updatedCard !== null) {
        card.value = updatedCard
      }
      onCardChanged()
    }
    catch (error: unknown) {
      failedOrderId.value = orderId
      failureMessage.value = failureMessageOf(error)
    }
    finally {
      busyOrderId.value = null
    }
  }

  async function reportFill(fillReportDto: FillReportDto): Promise<void> {
    await runOrderAction(fillReportDto.orderId, () => tradeCardApplication.reportFill(fillReportDto))
  }

  async function skipOrder(orderActionDto: OrderActionDto): Promise<void> {
    await runOrderAction(orderActionDto.orderId, () => tradeCardApplication.skipOrder(orderActionDto))
  }

  return { busyOrderId, failedOrderId, failureMessage, reportFill, skipOrder }
}
