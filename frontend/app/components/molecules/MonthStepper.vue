<script setup lang="ts">
import type { MonthDto } from '~/domain/models/dto/month-dto'
import AppButton from '~/components/atoms/AppButton.vue'

// ‹ 2026 年 11 月 ›。上一個月永遠可以去；下一個月在本月時不存在（未來沒有東西可看）。
defineProps<{
  month: MonthDto
}>()

const emit = defineEmits<{
  change: [month: string]
}>()
</script>

<template>
  <div class="month-stepper">
    <AppButton
      variant="ghost"
      size="sm"
      aria-label="上一個月"
      data-testid="previous-month"
      @click="emit('change', month.previous)"
    >
      ‹
    </AppButton>
    <span
      class="month-stepper__label"
      aria-live="polite"
    >{{ month.label }}</span>
    <AppButton
      v-if="month.next !== null"
      variant="ghost"
      size="sm"
      aria-label="下一個月"
      data-testid="next-month"
      @click="emit('change', month.next)"
    >
      ›
    </AppButton>
    <AppButton
      v-else
      variant="ghost"
      size="sm"
      aria-label="下一個月"
      data-testid="next-month"
      disabled
    >
      ›
    </AppButton>
  </div>
</template>

<style scoped lang="scss">
.month-stepper {
  display: flex;
  gap: spacing('sm');
  justify-content: center;
  align-items: center;

  &__label {
    min-width: control-size('month-label');
    color: color('text-strong');
    font-weight: font-weight('bold');
    text-align: center;
  }
}
</style>
