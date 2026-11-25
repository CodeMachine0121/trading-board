<script setup lang="ts">
// 一條 0–100 的進度條。數字由使用端算好，這裡只畫。
// 寬度經由 CSS 的 v-bind() 交給樣式，不寫 inline style。
defineProps<{
  percent: number
  label: string
}>()
</script>

<template>
  <div
    class="app-progress-bar"
    role="progressbar"
    :aria-label="label"
    :aria-valuenow="percent"
    aria-valuemin="0"
    aria-valuemax="100"
  >
    <div class="app-progress-bar__fill" />
  </div>
</template>

<style scoped lang="scss">
.app-progress-bar {
  overflow: hidden;
  border-radius: radius('pill');
  background-color: color('track');
  height: control-size('bar');

  &__fill {
    transition: width duration('normal') ease;
    border-radius: radius('pill');
    background-color: color('primary');
    width: calc(v-bind('percent') * 1%);
    height: 100%;
  }
}
</style>
