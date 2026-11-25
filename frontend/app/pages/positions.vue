<script setup lang="ts">
import AppAlert from '~/components/atoms/AppAlert.vue'
import AppSpinner from '~/components/atoms/AppSpinner.vue'
import MonthStepper from '~/components/molecules/MonthStepper.vue'
import ClosedPositionList from '~/components/organisms/ClosedPositionList.vue'
import OpenPositionList from '~/components/organisms/OpenPositionList.vue'
import { useAsyncView } from '~/composables/use-async-view'

// 持股：還持有的，以及選定月份出清的（紅賺綠賠）。
const { $positionApplication } = useNuxtApp()
const route = useRoute()

const requestedMonth = computed(() => typeof route.query.month === 'string' ? route.query.month : undefined)
const openView = useAsyncView(() => $positionApplication.listOpenPositions())
const closedView = useAsyncView(() => $positionApplication.listClosedPositions(requestedMonth.value))
const { data: openPositions, loading: openLoading, errorMessage: openErrorMessage } = openView
const { data: closedList, loading: closedLoading, errorMessage: closedErrorMessage } = closedView

watch(requestedMonth, () => void closedView.reload())
onMounted(() => {
  void openView.reload()
  void closedView.reload()
})
</script>

<template>
  <div class="positions-page">
    <section class="positions-page__section">
      <h2 class="positions-page__title">
        持有中
      </h2>
      <AppSpinner v-if="openLoading" />
      <AppAlert
        v-else-if="openErrorMessage !== null"
        variant="error"
      >
        {{ openErrorMessage }}
      </AppAlert>
      <OpenPositionList
        v-else-if="openPositions !== null"
        :positions="openPositions"
      />
    </section>

    <section class="positions-page__section">
      <h2 class="positions-page__title">
        已出清
      </h2>
      <MonthStepper
        v-if="closedList !== null"
        :month="closedList.month"
        @change="month => navigateTo({ query: { month } })"
      />
      <AppSpinner v-if="closedLoading" />
      <AppAlert
        v-else-if="closedErrorMessage !== null"
        variant="error"
      >
        {{ closedErrorMessage }}
      </AppAlert>
      <ClosedPositionList
        v-else-if="closedList !== null"
        :positions="closedList.positions"
      />
    </section>
  </div>
</template>

<style scoped lang="scss">
.positions-page {
  display: flex;
  flex-direction: column;
  gap: spacing('md');
}

.positions-page__section {
  display: flex;
  flex-direction: column;
  gap: spacing('sm');
}

.positions-page__title {
  color: color('text-strong');
  font-size: font-size('xl');
}
</style>
