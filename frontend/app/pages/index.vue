<script setup lang="ts">
import AppAlert from '~/components/atoms/AppAlert.vue'
import AppButton from '~/components/atoms/AppButton.vue'
import AppSpinner from '~/components/atoms/AppSpinner.vue'
import BudgetStrip from '~/components/molecules/BudgetStrip.vue'
import TradeCardView from '~/components/organisms/TradeCardView.vue'
import { useAsyncView } from '~/composables/use-async-view'
import { useTradeCardActions } from '~/composables/use-trade-card-actions'

// 今日：本月預算那一條，以及最新的一張交易卡。
const { $tradeCardApplication, $monthSummaryApplication } = useNuxtApp()

const budgetView = useAsyncView(() => $monthSummaryApplication.fetchBudgetStrip())
const cardView = useAsyncView(() => $tradeCardApplication.fetchLatestCard())

// 回報成交會動到預算，所以成功之後順便重讀那一條。
const { busyOrderId, failedOrderId, failureMessage, reportFill, skipOrder }
  = useTradeCardActions($tradeCardApplication, cardView.data, () => void budgetView.reload())

const { data: budget, errorMessage: budgetErrorMessage } = budgetView
const { data: card, loading: cardLoading, errorMessage: cardErrorMessage } = cardView

onMounted(() => {
  void budgetView.reload()
  void cardView.reload()
})
</script>

<template>
  <div class="today-page">
    <BudgetStrip
      v-if="budget !== null"
      :budget="budget"
    />
    <AppAlert
      v-else-if="budgetErrorMessage !== null"
      variant="error"
    >
      {{ budgetErrorMessage }}
    </AppAlert>

    <AppSpinner v-if="cardLoading && card === null" />
    <AppAlert
      v-else-if="cardErrorMessage !== null"
      variant="error"
    >
      {{ cardErrorMessage }}
      <AppButton
        size="sm"
        variant="secondary"
        @click="cardView.reload()"
      >
        重試
      </AppButton>
    </AppAlert>
    <p
      v-else-if="card === null"
      class="today-page__empty"
    >
      尚未發布任何交易卡
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
.today-page {
  display: flex;
  flex-direction: column;
  gap: spacing('md');
}

.today-page__empty {
  padding: spacing('xl') 0;
  color: color('text-muted');
  text-align: center;
}
</style>
