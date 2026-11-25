<script setup lang="ts">
import { ref } from 'vue'
import type { OrderDto } from '~/domain/models/dto/order-dto'
import { FillReportDto } from '~/domain/models/dto/fill-report-dto'
import AppAlert from '~/components/atoms/AppAlert.vue'
import AppButton from '~/components/atoms/AppButton.vue'
import AppInput from '~/components/atoms/AppInput.vue'

// 回報成交表單。初始值（限價、剩餘股數、現在時間）由 OrderDto 帶來；
// 打的字原樣交給 Application 驗證——這裡不判斷任何業務規則。
const props = defineProps<{
  order: OrderDto
  submitting: boolean
  errorMessage: string | null
}>()

const emit = defineEmits<{
  submit: [fillReport: FillReportDto]
  cancel: []
}>()

const priceText = ref(props.order.defaultFillPrice)
const sharesText = ref(props.order.defaultFillShares)
const feeText = ref('')
const filledAtText = ref(props.order.defaultFilledAt)

function submit(): void {
  emit('submit', new FillReportDto(
    props.order.id,
    props.order.tradeDate,
    props.order.remainingShares,
    priceText.value,
    sharesText.value,
    feeText.value,
    filledAtText.value,
  ))
}
</script>

<template>
  <form
    class="fill-report-form"
    novalidate
    @submit.prevent="submit"
  >
    <div class="fill-report-form__fields">
      <AppInput
        v-model="priceText"
        label="成交價"
        inputmode="decimal"
        data-testid="fill-price"
      />
      <AppInput
        v-model="sharesText"
        label="股數"
        inputmode="numeric"
        :hint="`剩餘 ${order.remainingShares} 股`"
        data-testid="fill-shares"
      />
      <AppInput
        v-model="feeText"
        label="手續費（選填）"
        inputmode="decimal"
        hint="留空由系統估算"
        data-testid="fill-fee"
      />
      <AppInput
        v-model="filledAtText"
        label="成交時間"
        type="datetime-local"
        data-testid="fill-filled-at"
      />
    </div>
    <AppAlert
      v-if="errorMessage !== null"
      variant="error"
      data-testid="fill-error"
    >
      {{ errorMessage }}
    </AppAlert>
    <div class="fill-report-form__actions">
      <AppButton
        type="button"
        variant="secondary"
        :disabled="submitting"
        @click="emit('cancel')"
      >
        取消
      </AppButton>
      <AppButton
        type="submit"
        :loading="submitting"
        data-testid="fill-submit"
      >
        送出
      </AppButton>
    </div>
  </form>
</template>

<style scoped lang="scss">
.fill-report-form {
  display: flex;
  flex-direction: column;
  gap: spacing('sm');
  border-radius: radius('md');
  background-color: color('surface-muted');
  padding: spacing('sm');

  &__fields {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: spacing('sm');
  }

  &__actions {
    display: flex;
    gap: spacing('xs');
    justify-content: flex-end;
  }
}
</style>
