<script setup lang="ts">
import type { BudgetStripDto } from '~/domain/models/dto/budget-strip-dto'
import AppBadge from '~/components/atoms/AppBadge.vue'
import AppProgressBar from '~/components/atoms/AppProgressBar.vue'

// 今日畫面頂端的本月預算：剩餘 / 月配額、還剩幾成、現在是嚴格期還是月底配置期。
defineProps<{
  budget: BudgetStripDto
}>()
</script>

<template>
  <section
    class="budget-strip"
    aria-label="本月預算"
  >
    <div class="budget-strip__head">
      <p class="budget-strip__amount">
        <span class="budget-strip__label">剩餘</span>
        <strong
          class="budget-strip__remaining"
          data-testid="budget-remaining"
        >{{ budget.remainingText }}</strong>
        <span class="budget-strip__allocation">/ 月配額 {{ budget.allocationText }}</span>
      </p>
      <AppBadge
        :variant="budget.modeTone"
        data-testid="budget-mode"
      >
        {{ budget.modeLabel }}
      </AppBadge>
    </div>
    <AppProgressBar
      :percent="budget.remainingPercent"
      label="本月可用資金剩餘比例"
    />
  </section>
</template>

<style scoped lang="scss">
.budget-strip {
  display: flex;
  flex-direction: column;
  gap: spacing('xs');
  padding: spacing('sm') spacing('md');

  @include surface-card;

  &__head {
    display: flex;
    gap: spacing('xs');
    justify-content: space-between;
    align-items: center;
  }

  &__amount {
    display: flex;
    flex-wrap: wrap;
    gap: spacing('2xs');
    align-items: baseline;
  }

  &__label {
    @include dense-label;
  }

  &__remaining {
    color: color('text-strong');
    font-size: font-size('xl');

    @include numeric;
  }

  &__allocation {
    color: color('text-muted');
    font-size: font-size('sm');
  }
}
</style>
