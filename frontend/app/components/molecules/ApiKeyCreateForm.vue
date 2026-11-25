<script setup lang="ts">
import { ref } from 'vue'
import { ApiKeyCreateDto } from '~/domain/models/dto/api-key-create-dto'
import AppAlert from '~/components/atoms/AppAlert.vue'
import AppButton from '~/components/atoms/AppButton.vue'
import AppInput from '~/components/atoms/AppInput.vue'

// 建立一把成員金鑰：成員名稱（必填）與備註（選填）。規則由板子說了算，這裡只擋空白。
defineProps<{
  submitting: boolean
  errorMessage: string | null
}>()

const emit = defineEmits<{
  submit: [apiKeyCreateDto: ApiKeyCreateDto]
}>()

const memberName = ref('')
const note = ref('')
</script>

<template>
  <form
    class="api-key-create-form"
    @submit.prevent="emit('submit', new ApiKeyCreateDto(memberName, note))"
  >
    <AppInput
      v-model="memberName"
      label="成員名稱"
      hint="例如 stock-scanner、desk-chief（1 到 40 個字）"
      maxlength="40"
      required
      data-testid="api-key-member-name"
    />
    <AppInput
      v-model="note"
      label="備註（選填）"
      maxlength="200"
      data-testid="api-key-note"
    />
    <AppAlert
      v-if="errorMessage !== null"
      variant="error"
      data-testid="api-key-create-error"
    >
      {{ errorMessage }}
    </AppAlert>
    <AppButton
      type="submit"
      :loading="submitting"
      data-testid="api-key-create"
    >
      建立金鑰
    </AppButton>
  </form>
</template>

<style scoped lang="scss">
.api-key-create-form {
  display: flex;
  flex-direction: column;
  gap: spacing('sm');
  padding: spacing('md');

  @include surface-card;
}
</style>
