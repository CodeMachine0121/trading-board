<script setup lang="ts">
// 小標籤。語氣由使用端決定；profit 是紅、loss 是綠（台股慣例），定義在 token 裡。
withDefaults(defineProps<{
  variant?: 'primary' | 'info' | 'warning' | 'neutral' | 'profit' | 'loss'
}>(), {
  variant: 'neutral',
})
</script>

<template>
  <span
    class="app-badge"
    :class="`app-badge--${variant}`"
  >
    <slot />
  </span>
</template>

<style scoped lang="scss">
.app-badge {
  display: inline-flex;
  align-items: center;
  border-radius: radius('pill');
  padding: spacing('3xs') spacing('xs');
  font-weight: font-weight('medium');
  font-size: font-size('xs');
  line-height: line-height('tight');
  white-space: nowrap;

  &--primary {
    background-color: color('primary-soft');
    color: color('primary');
  }

  &--info {
    background-color: color('info-soft');
    color: color('info');
  }

  &--warning {
    background-color: color('warning-soft');
    color: color('warning');
  }

  &--neutral {
    background-color: color('neutral-soft');
    color: color('neutral');
  }

  &--profit {
    background-color: color('profit-soft');
    color: color('profit');
  }

  &--loss {
    background-color: color('loss-soft');
    color: color('loss');
  }
}
</style>
