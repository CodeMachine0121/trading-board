<script setup lang="ts">
import type { ClosedPositionDto } from '~/domain/models/dto/closed-position-dto'
import AppBadge from '~/components/atoms/AppBadge.vue'
import DetailList from '~/components/molecules/DetailList.vue'

// 某個月出清的股票。賺賠的顏色（紅賺綠賠）由 DTO 的 tone 決定。
defineProps<{
  positions: readonly ClosedPositionDto[]
}>()
</script>

<template>
  <p
    v-if="positions.length === 0"
    class="closed-position-list__empty"
    data-testid="closed-positions-empty"
  >
    這個月沒有出清的持股
  </p>
  <ul
    v-else
    class="closed-position-list"
  >
    <li
      v-for="(position, index) in positions"
      :key="index"
      class="closed-position-list__item"
      data-testid="closed-position"
    >
      <header class="closed-position-list__head">
        <h3 class="closed-position-list__title">
          {{ position.title }}
        </h3>
        <AppBadge
          :variant="position.tone"
          data-testid="closed-position-outcome"
        >
          {{ position.outcomeLabel }}
        </AppBadge>
        <strong
          class="closed-position-list__pnl"
          :class="`closed-position-list__pnl--${position.tone}`"
          data-testid="closed-position-pnl"
        >{{ position.realizedPnlText }}</strong>
      </header>
      <DetailList :rows="position.detailRows" />
    </li>
  </ul>
</template>

<style scoped lang="scss">
.closed-position-list {
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

  &__head {
    display: flex;
    flex-wrap: wrap;
    gap: spacing('xs');
    align-items: center;
  }

  &__title {
    color: color('text-strong');
    font-size: font-size('lg');
  }

  &__pnl {
    margin-left: auto;
    font-size: font-size('lg');

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
    padding: spacing('lg') 0;
    color: color('text-muted');
    text-align: center;
  }
}
</style>
