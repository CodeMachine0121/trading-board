<script setup lang="ts">
// 一段提示。error 會被螢幕閱讀器立刻念出來（role="alert"）。
withDefaults(defineProps<{
  variant?: 'error' | 'info'
}>(), {
  variant: 'info',
})
</script>

<template>
  <div
    class="app-alert"
    :class="`app-alert--${variant}`"
    :role="variant === 'error' ? 'alert' : 'status'"
  >
    <slot />
  </div>
</template>

<style scoped lang="scss">
.app-alert {
  display: flex;
  flex-wrap: wrap;
  gap: spacing('sm');
  align-items: center;
  border-radius: radius('md');
  padding: spacing('sm') spacing('md');
  font-size: font-size('sm');

  &--error {
    background-color: color('danger-soft');
    color: color('danger');
  }

  &--info {
    background-color: color('info-soft');
    color: color('info');
  }
}
</style>
