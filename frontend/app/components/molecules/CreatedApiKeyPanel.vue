<script setup lang="ts">
import type { CreatedApiKeyDto } from '~/domain/models/dto/created-api-key-dto'
import AppAlert from '~/components/atoms/AppAlert.vue'
import AppButton from '~/components/atoms/AppButton.vue'

// 剛建立的金鑰：完整金鑰只在這裡出現這一次，附複製按鈕與提醒。
defineProps<{
  createdKey: CreatedApiKeyDto
  copied: boolean
  copyErrorMessage: string | null
}>()

const emit = defineEmits<{
  copy: [key: string]
}>()
</script>

<template>
  <section
    class="created-api-key-panel"
    data-testid="created-api-key"
  >
    <h3 class="created-api-key-panel__title">
      {{ createdKey.apiKey.memberName }} 的金鑰
    </h3>
    <AppAlert variant="info">
      {{ createdKey.warning }}
    </AppAlert>
    <code
      class="created-api-key-panel__key"
      data-testid="created-api-key-value"
    >{{ createdKey.key }}</code>
    <AppButton
      variant="secondary"
      data-testid="created-api-key-copy"
      @click="emit('copy', createdKey.key)"
    >
      {{ copied ? '已複製' : '複製金鑰' }}
    </AppButton>
    <AppAlert
      v-if="copyErrorMessage !== null"
      variant="error"
    >
      {{ copyErrorMessage }}
    </AppAlert>
  </section>
</template>

<style scoped lang="scss">
.created-api-key-panel {
  display: flex;
  flex-direction: column;
  gap: spacing('sm');
  padding: spacing('md');

  @include surface-card;

  &__title {
    color: color('text-strong');
    font-size: font-size('lg');
  }

  &__key {
    overflow-wrap: anywhere;
    border-radius: radius('md');
    background-color: color('surface-muted');
    padding: spacing('sm');
    font-family: font-family('numeric');
    font-size: font-size('sm');
    user-select: all;
  }
}
</style>
