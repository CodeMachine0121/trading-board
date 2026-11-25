<script setup lang="ts">
import type { FillReportDto } from '~/domain/models/dto/fill-report-dto'
import type { OrderActionDto } from '~/domain/models/dto/order-action-dto'
import type { TradeCardDto } from '~/domain/models/dto/trade-card-dto'
import OrderCard from '~/components/organisms/OrderCard.vue'

// 一張交易卡：今日不買就是一條橫幅；否則買進與賣出兩段委託。
// 今日與歷史明細用的是同一個它——過期的委託唯讀，那是 OrderDto 決定的。
defineProps<{
  card: TradeCardDto
  busyOrderId: string | null
  failedOrderId: string | null
  failureMessage: string | null
}>()

const emit = defineEmits<{
  report: [fillReport: FillReportDto]
  skip: [orderAction: OrderActionDto]
}>()
</script>

<template>
  <section class="trade-card-view">
    <header class="trade-card-view__head">
      <h2 class="trade-card-view__date">
        {{ card.tradeDateLabel }}
      </h2>
      <p class="trade-card-view__published">
        {{ card.publishedAtText }}
      </p>
    </header>

    <div
      v-if="card.isNoTrade"
      class="trade-card-view__no-trade"
      data-testid="no-trade-banner"
    >
      <strong class="trade-card-view__no-trade-title">今日不買</strong>
      <p v-if="card.note !== ''">
        {{ card.note }}
      </p>
    </div>

    <template v-else>
      <p
        v-if="card.note !== ''"
        class="trade-card-view__note"
      >
        {{ card.note }}
      </p>
      <section
        class="trade-card-view__section"
        data-testid="buy-section"
      >
        <h3 class="trade-card-view__section-title trade-card-view__section-title--buy">
          買進
        </h3>
        <p
          v-if="card.buyOrders.length === 0"
          class="trade-card-view__empty"
        >
          {{ card.emptyBuyText }}
        </p>
        <OrderCard
          v-for="order in card.buyOrders"
          :key="order.id"
          :order="order"
          :busy="busyOrderId === order.id"
          :failure-message="failedOrderId === order.id ? failureMessage : null"
          @report="fillReport => emit('report', fillReport)"
          @skip="orderAction => emit('skip', orderAction)"
        />
      </section>
      <section
        class="trade-card-view__section"
        data-testid="sell-section"
      >
        <h3 class="trade-card-view__section-title trade-card-view__section-title--sell">
          賣出
        </h3>
        <p
          v-if="card.sellOrders.length === 0"
          class="trade-card-view__empty"
        >
          {{ card.emptySellText }}
        </p>
        <OrderCard
          v-for="order in card.sellOrders"
          :key="order.id"
          :order="order"
          :busy="busyOrderId === order.id"
          :failure-message="failedOrderId === order.id ? failureMessage : null"
          @report="fillReport => emit('report', fillReport)"
          @skip="orderAction => emit('skip', orderAction)"
        />
      </section>
    </template>
  </section>
</template>

<style scoped lang="scss">
.trade-card-view {
  display: flex;
  flex-direction: column;
  gap: spacing('md');

  &__head {
    display: flex;
    flex-wrap: wrap;
    gap: spacing('xs');
    justify-content: space-between;
    align-items: baseline;
  }

  &__date {
    color: color('text-strong');
    font-size: font-size('xl');
  }

  &__published {
    color: color('text-muted');
    font-size: font-size('xs');
  }

  &__no-trade {
    display: flex;
    flex-direction: column;
    gap: spacing('xs');
    border-radius: radius('lg');
    background-color: color('warning-soft');
    padding: spacing('lg') spacing('md');
    color: color('warning');
    text-align: center;
  }

  &__no-trade-title {
    font-size: font-size('2xl');
  }

  &__note {
    color: color('text-muted');
    font-size: font-size('sm');
  }

  &__section {
    display: flex;
    flex-direction: column;
    gap: spacing('sm');
  }

  &__section-title {
    font-size: font-size('lg');

    &--buy {
      color: color('profit');
    }

    &--sell {
      color: color('loss');
    }
  }

  &__empty {
    color: color('text-muted');
  }
}
</style>
