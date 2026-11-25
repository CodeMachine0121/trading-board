<script setup lang="ts">
// 全站唯一的按鈕。長相由使用端以 variant / size / block 決定；
// 原生行為（type、disabled、@click）走 attribute fallthrough。
withDefaults(defineProps<{
  variant?: 'primary' | 'secondary' | 'ghost' | 'danger'
  size?: 'md' | 'sm'
  block?: boolean
  loading?: boolean
}>(), {
  variant: 'primary',
  size: 'md',
  block: false,
  loading: false,
})
</script>

<template>
  <button
    class="app-button"
    :class="[
      `app-button--${variant}`,
      `app-button--${size}`,
      { 'app-button--block': block },
    ]"
    :disabled="loading || undefined"
    :aria-busy="loading || undefined"
  >
    <slot />
  </button>
</template>

<style scoped lang="scss">
.app-button {
  display: inline-flex;
  gap: spacing('xs');
  justify-content: center;
  align-items: center;
  transition: background-color duration('fast') ease, border-color duration('fast') ease;
  cursor: pointer;
  border: 1px solid transparent;
  border-radius: radius('md');
  padding: spacing('xs') spacing('md');
  font-weight: font-weight('medium');
  font-size: font-size('md');

  @include tap-target;
  @include focus-ring;

  &:disabled {
    cursor: not-allowed;
    opacity: 0.6;
  }

  &--primary {
    background-color: color('primary');
    color: color('text-inverse');

    &:hover:not(:disabled) {
      background-color: color('primary-strong');
    }
  }

  &--secondary {
    border-color: color('border-strong');
    background-color: color('surface');
    color: color('text');

    &:hover:not(:disabled) {
      background-color: color('surface-muted');
    }
  }

  &--ghost {
    background-color: transparent;
    color: color('primary');

    &:hover:not(:disabled) {
      background-color: color('primary-soft');
    }
  }

  &--danger {
    border-color: color('danger');
    background-color: color('surface');
    color: color('danger');

    &:hover:not(:disabled) {
      background-color: color('danger-soft');
    }
  }

  &--sm {
    padding: spacing('2xs') spacing('sm');
    font-size: font-size('sm');
  }

  &--block {
    width: 100%;
  }
}
</style>
