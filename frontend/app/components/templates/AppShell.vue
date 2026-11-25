<script setup lang="ts">
// 版面骨架：頂列（名稱、導覽、動作）與內容區。只有插槽，不綁任何資料。
// 手機上導覽排成頂列下方的一整排分頁，寬螢幕時併進頂列。
</script>

<template>
  <div class="app-shell">
    <header class="app-shell__header">
      <div class="app-shell__bar">
        <div class="app-shell__brand">
          <slot name="brand" />
        </div>
        <div class="app-shell__actions">
          <slot name="actions" />
        </div>
      </div>
      <nav
        class="app-shell__nav"
        aria-label="主要導覽"
      >
        <slot name="nav" />
      </nav>
    </header>
    <main class="app-shell__main">
      <slot />
    </main>
  </div>
</template>

<style scoped lang="scss">
.app-shell {
  display: flex;
  flex-direction: column;
  min-height: 100vh;

  &__header {
    position: sticky;
    top: 0;
    z-index: z-index('header');
    border-bottom: 1px solid color('border');
    background-color: color('surface');
  }

  &__bar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin: 0 auto;
    padding: spacing('xs') spacing('md');
    max-width: control-size('content-max');
  }

  &__brand {
    color: color('text-strong');
    font-weight: font-weight('bold');
  }

  &__nav {
    display: flex;
    margin: 0 auto;
    max-width: control-size('content-max');
  }

  &__main {
    display: flex;
    flex-direction: column;
    gap: spacing('md');
    margin: 0 auto;
    padding: spacing('md');
    width: 100%;
    max-width: control-size('content-max');
  }
}
</style>
