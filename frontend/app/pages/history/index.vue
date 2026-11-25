<script setup lang="ts">
import AppAlert from '~/components/atoms/AppAlert.vue'
import AppSpinner from '~/components/atoms/AppSpinner.vue'
import MonthStepper from '~/components/molecules/MonthStepper.vue'
import CardSummaryList from '~/components/organisms/CardSummaryList.vue'
import { useAsyncView } from '~/composables/use-async-view'

// 歷史：選一個月 → 那個月的交易卡 → 點進去看明細。月份記在網址上（?month=），重新整理不會跑掉。
const { $tradeCardApplication } = useNuxtApp()
const route = useRoute()

const requestedMonth = computed(() => typeof route.query.month === 'string' ? route.query.month : undefined)
const { data: cardList, loading, errorMessage, reload } = useAsyncView(
  () => $tradeCardApplication.listCards(requestedMonth.value))

watch(requestedMonth, () => void reload())
onMounted(() => void reload())
</script>

<template>
  <div class="history-page">
    <MonthStepper
      v-if="cardList !== null"
      :month="cardList.month"
      @change="month => navigateTo({ query: { month } })"
    />
    <AppSpinner v-if="loading" />
    <AppAlert
      v-else-if="errorMessage !== null"
      variant="error"
    >
      {{ errorMessage }}
    </AppAlert>
    <CardSummaryList
      v-else-if="cardList !== null"
      :cards="cardList.cards"
      @select="tradeDate => navigateTo(`/history/${tradeDate}`)"
    />
  </div>
</template>

<style scoped lang="scss">
.history-page {
  display: flex;
  flex-direction: column;
  gap: spacing('md');
}
</style>
