<script setup lang="ts">
// 載入中。說明文字給螢幕閱讀器，畫面上只有轉圈。
withDefaults(defineProps<{
  label?: string
}>(), {
  label: '載入中',
})
</script>

<template>
  <div
    class="app-spinner"
    role="status"
  >
    <span
      class="app-spinner__circle"
      aria-hidden="true"
    />
    <span class="app-spinner__label">{{ label }}</span>
  </div>
</template>

<style scoped lang="scss">
.app-spinner {
  display: flex;
  justify-content: center;
  padding: spacing('xl') 0;

  &__circle {
    animation: app-spinner-rotate 0.8s linear infinite;
    border: 3px solid color('track');
    border-top-color: color('primary');
    border-radius: radius('pill');
    width: spacing('xl');
    height: spacing('xl');
  }

  &__label {
    @include visually-hidden;
  }
}

@keyframes app-spinner-rotate {
  to {
    transform: rotate(360deg);
  }
}
</style>
