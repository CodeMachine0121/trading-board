<script setup lang="ts">
import AppAlert from '~/components/atoms/AppAlert.vue'
import AppButton from '~/components/atoms/AppButton.vue'
import AppSpinner from '~/components/atoms/AppSpinner.vue'
import TradeCardView from '~/components/organisms/TradeCardView.vue'
import { useAsyncView } from '~/composables/use-async-view'
import { useTradeCardActions } from '~/composables/use-trade-card-actions'

// 一張過去的交易卡：與今日同一個元件。過期的委託唯讀，部分成交的仍可補回報。
const { $tradeCardApplication } = useNuxtApp()
const route = useRoute()
const tradeDate = String(route.params.tradeDate)

const cardView = useAsyncView(() => $tradeCardApplication.fetchCard(tradeDate))
const { busyOrderId, failedOrderId, failureMessage, reportFill, skipOrder }
  = useTradeCardActions($tradeCardApplication, cardView.data)
const { data: card, loading, errorMessage } = cardView

onMounted(() => void cardView.reload())
</script>

<template>
  <div class="history-card-page">
    <div>
      <AppButton
        variant="ghost"
        size="sm"
        @click="navigateTo({ path: '/history', query: { month: tradeDate.slice(0, 7) } })"
      >
        ‹ 回到歷史
      </AppButton>
    </div>
    <AppSpinner v-if="loading && card === null" />
    <AppAlert
      v-else-if="errorMessage !== null"
      variant="error"
    >
      {{ errorMessage }}
    </AppAlert>
    <p
      v-else-if="card === null"
      class="history-card-page__empty"
    >
      找不到這一天的交易卡
    </p>
    <TradeCardView
      v-else
      :card="card"
      :busy-order-id="busyOrderId"
      :failed-order-id="failedOrderId"
      :failure-message="failureMessage"
      @report="reportFill"
      @skip="skipOrder"
    />
  </div>
</template>

<style scoped lang="scss">
.history-card-page {
  display: flex;
  flex-direction: column;
  gap: spacing('md');
}

.history-card-page__empty {
  padding: spacing('xl') 0;
  color: color('text-muted');
  text-align: center;
}
</style>
