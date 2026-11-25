<script setup lang="ts">
import AppAlert from '~/components/atoms/AppAlert.vue'
import AppSpinner from '~/components/atoms/AppSpinner.vue'
import MonthStepper from '~/components/molecules/MonthStepper.vue'
import MonthSummaryPanel from '~/components/organisms/MonthSummaryPanel.vue'
import { useAsyncView } from '~/composables/use-async-view'

// 月報：選一個月 → 預算明細與已實現績效。
const { $monthSummaryApplication } = useNuxtApp()
const route = useRoute()

const requestedMonth = computed(() => typeof route.query.month === 'string' ? route.query.month : undefined)
const { data: page, loading, errorMessage, reload } = useAsyncView(
  () => $monthSummaryApplication.fetchMonthSummary(requestedMonth.value))

watch(requestedMonth, () => void reload())
onMounted(() => void reload())
</script>

<template>
  <div class="summary-page">
    <MonthStepper
      v-if="page !== null"
      :month="page.month"
      @change="month => navigateTo({ query: { month } })"
    />
    <AppSpinner v-if="loading" />
    <AppAlert
      v-else-if="errorMessage !== null"
      variant="error"
    >
      {{ errorMessage }}
    </AppAlert>
    <MonthSummaryPanel
      v-else-if="page !== null"
      :summary="page.summary"
    />
  </div>
</template>

<style scoped lang="scss">
.summary-page {
  display: flex;
  flex-direction: column;
  gap: spacing('md');
}
</style>
