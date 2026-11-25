<script setup lang="ts">
import type { OpenPositionDto } from '~/domain/models/dto/open-position-dto'
import DetailList from '~/components/molecules/DetailList.vue'

// 還持有的股票。
defineProps<{
  positions: readonly OpenPositionDto[]
}>()
</script>

<template>
  <p
    v-if="positions.length === 0"
    class="open-position-list__empty"
    data-testid="open-positions-empty"
  >
    目前沒有持股
  </p>
  <ul
    v-else
    class="open-position-list"
  >
    <li
      v-for="position in positions"
      :key="position.symbol"
      class="open-position-list__item"
      data-testid="open-position"
    >
      <h3 class="open-position-list__title">
        {{ position.title }}
      </h3>
      <DetailList :rows="position.detailRows" />
    </li>
  </ul>
</template>

<style scoped lang="scss">
.open-position-list {
  display: flex;
  flex-direction: column;
  gap: spacing('sm');

  &__item {
    display: flex;
    flex-direction: column;
    gap: spacing('sm');
    padding: spacing('md');

    @include surface-card;
  }

  &__title {
    color: color('text-strong');
    font-size: font-size('lg');
  }

  &__empty {
    padding: spacing('lg') 0;
    color: color('text-muted');
    text-align: center;
  }
}
</style>
