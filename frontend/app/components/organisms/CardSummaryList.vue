<script setup lang="ts">
import type { TradeCardSummaryDto } from '~/domain/models/dto/trade-card-summary-dto'
import AppBadge from '~/components/atoms/AppBadge.vue'

// 一個月的交易卡清單。點一列就打開那一張——要去哪裡由頁面決定。
defineProps<{
  cards: readonly TradeCardSummaryDto[]
}>()

const emit = defineEmits<{
  select: [tradeDate: string]
}>()
</script>

<template>
  <p
    v-if="cards.length === 0"
    class="card-summary-list__empty"
    data-testid="card-list-empty"
  >
    這個月沒有交易卡
  </p>
  <ul
    v-else
    class="card-summary-list"
  >
    <li
      v-for="card in cards"
      :key="card.tradeDate"
    >
      <button
        type="button"
        class="card-summary-list__row"
        data-testid="card-summary-row"
        @click="emit('select', card.tradeDate)"
      >
        <span class="card-summary-list__date">{{ card.tradeDateLabel }}</span>
        <AppBadge :variant="card.statusTone">
          {{ card.statusLabel }}
        </AppBadge>
        <span class="card-summary-list__counts">{{ card.orderCountText }}</span>
        <span class="card-summary-list__counts">{{ card.filledOrderCountText }}</span>
      </button>
    </li>
  </ul>
</template>

<style scoped lang="scss">
.card-summary-list {
  display: flex;
  flex-direction: column;
  gap: spacing('xs');

  &__row {
    display: flex;
    flex-wrap: wrap;
    gap: spacing('xs') spacing('sm');
    align-items: center;
    cursor: pointer;
    padding: spacing('sm') spacing('md');
    width: 100%;
    color: color('text');
    text-align: left;

    @include surface-card;
    @include tap-target;
    @include focus-ring;

    &:hover {
      background-color: color('surface-muted');
    }
  }

  &__date {
    color: color('text-strong');
    font-weight: font-weight('bold');
  }

  &__counts {
    color: color('text-muted');
    font-size: font-size('sm');
  }

  &__empty {
    padding: spacing('xl') 0;
    color: color('text-muted');
    text-align: center;
  }
}
</style>
