<script setup lang="ts">
import type { MonthSummaryDto } from '~/domain/models/dto/month-summary-dto'
import AppBadge from '~/components/atoms/AppBadge.vue'
import DetailList from '~/components/molecules/DetailList.vue'

// 一個月的預算明細與已實現績效。`null` 代表那個月板子沒有資料。
defineProps<{
  summary: MonthSummaryDto | null
}>()
</script>

<template>
  <p
    v-if="summary === null"
    class="month-summary-panel__empty"
    data-testid="month-summary-empty"
  >
    這個月沒有資料
  </p>
  <div
    v-else
    class="month-summary-panel"
  >
    <section class="month-summary-panel__block">
      <header class="month-summary-panel__head">
        <h2 class="month-summary-panel__title">
          預算
        </h2>
        <AppBadge :variant="summary.modeTone">
          {{ summary.modeLabel }}
        </AppBadge>
      </header>
      <DetailList :rows="summary.budgetRows" />
    </section>
    <section class="month-summary-panel__block">
      <header class="month-summary-panel__head">
        <h2 class="month-summary-panel__title">
          已實現損益
        </h2>
        <strong
          class="month-summary-panel__pnl"
          :class="`month-summary-panel__pnl--${summary.realizedPnlTone}`"
          data-testid="month-realized-pnl"
        >{{ summary.realizedPnlText }}</strong>
      </header>
      <DetailList :rows="summary.performanceRows" />
    </section>
  </div>
</template>

<style scoped lang="scss">
.month-summary-panel {
  display: flex;
  flex-direction: column;
  gap: spacing('md');

  &__block {
    display: flex;
    flex-direction: column;
    gap: spacing('sm');
    padding: spacing('md');

    @include surface-card;
  }

  &__head {
    display: flex;
    gap: spacing('xs');
    justify-content: space-between;
    align-items: center;
  }

  &__title {
    color: color('text-strong');
    font-size: font-size('lg');
  }

  &__pnl {
    font-size: font-size('xl');

    @include numeric;

    &--profit {
      color: color('profit');
    }

    &--loss {
      color: color('loss');
    }

    &--neutral {
      color: color('neutral');
    }
  }

  &__empty {
    padding: spacing('xl') 0;
    color: color('text-muted');
    text-align: center;
  }
}
</style>
