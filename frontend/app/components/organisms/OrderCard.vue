<script setup lang="ts">
import { ref, watch } from 'vue'
import type { FillReportDto } from '~/domain/models/dto/fill-report-dto'
import type { OrderDto } from '~/domain/models/dto/order-dto'
import { OrderActionDto } from '~/domain/models/dto/order-action-dto'
import AppAlert from '~/components/atoms/AppAlert.vue'
import AppBadge from '~/components/atoms/AppBadge.vue'
import AppButton from '~/components/atoms/AppButton.vue'
import DetailList from '~/components/molecules/DetailList.vue'
import FillReportForm from '~/components/molecules/FillReportForm.vue'

// 一筆委託（買進或賣出都是它）：明細、狀態、已回報的成交，以及「回報成交」「略過」。
// 能不能操作、要列哪些欄位，全由 OrderDto 決定。
const props = defineProps<{
  order: OrderDto
  busy: boolean
  failureMessage: string | null
}>()

const emit = defineEmits<{
  report: [fillReport: FillReportDto]
  skip: [orderAction: OrderActionDto]
}>()

const formOpen = ref(false)

// 板子收下之後會換來一份新的委託：表單完成任務，收起來。
watch(() => props.order, () => {
  formOpen.value = false
})
</script>

<template>
  <article
    class="order-card"
    :class="`order-card--${order.side}`"
    :data-testid="`order-${order.id}`"
  >
    <header class="order-card__head">
      <h3 class="order-card__title">
        {{ order.title }}
      </h3>
      <AppBadge
        v-for="tag in order.tags"
        :key="tag"
        variant="warning"
      >
        {{ tag }}
      </AppBadge>
      <AppBadge
        class="order-card__status"
        :variant="order.statusTone"
        data-testid="order-status"
      >
        {{ order.statusLabel }}
      </AppBadge>
    </header>

    <DetailList :rows="order.detailRows" />

    <p
      v-if="order.reason !== ''"
      class="order-card__reason"
    >
      {{ order.reason }}
    </p>

    <section class="order-card__fills">
      <p class="order-card__filled-shares">
        {{ order.filledSharesText }}
      </p>
      <ul
        v-if="order.fills.length > 0"
        class="order-card__fill-list"
      >
        <li
          v-for="fill in order.fills"
          :key="fill.id"
          class="order-card__fill"
          data-testid="order-fill"
        >
          <span>{{ fill.filledAtText }}</span>
          <span>{{ fill.priceText }} × {{ fill.sharesText }}</span>
          <span>金額 {{ fill.amountText }}</span>
          <span>手續費 {{ fill.feeText }}</span>
        </li>
      </ul>
    </section>

    <FillReportForm
      v-if="formOpen"
      :order="order"
      :submitting="busy"
      :error-message="failureMessage"
      @submit="fillReport => emit('report', fillReport)"
      @cancel="formOpen = false"
    />
    <template v-else>
      <AppAlert
        v-if="failureMessage !== null"
        variant="error"
        data-testid="order-error"
      >
        {{ failureMessage }}
      </AppAlert>
      <div
        v-if="order.actionable"
        class="order-card__actions"
      >
        <AppButton
          variant="secondary"
          :loading="busy"
          data-testid="order-skip"
          @click="emit('skip', new OrderActionDto(order.id, order.tradeDate))"
        >
          略過
        </AppButton>
        <AppButton
          :disabled="busy"
          data-testid="order-report"
          @click="formOpen = true"
        >
          回報成交
        </AppButton>
      </div>
    </template>
  </article>
</template>

<style scoped lang="scss">
.order-card {
  display: flex;
  flex-direction: column;
  gap: spacing('sm');
  border-left-width: spacing('2xs');
  padding: spacing('md');

  @include surface-card;

  &--buy {
    border-left-color: color('profit');
  }

  &--sell {
    border-left-color: color('loss');
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

  &__status {
    margin-left: auto;
  }

  &__reason {
    border-radius: radius('md');
    background-color: color('surface-muted');
    padding: spacing('xs') spacing('sm');
    color: color('text');
    font-size: font-size('sm');
    white-space: pre-line;
  }

  &__fills {
    display: flex;
    flex-direction: column;
    gap: spacing('2xs');
    font-size: font-size('sm');
  }

  &__filled-shares {
    color: color('text-muted');
  }

  &__fill {
    display: flex;
    flex-wrap: wrap;
    gap: spacing('2xs') spacing('sm');
    border-top: 1px solid color('border');
    padding-top: spacing('2xs');

    @include numeric;
  }

  &__actions {
    display: flex;
    gap: spacing('xs');
    justify-content: flex-end;
  }
}
</style>
