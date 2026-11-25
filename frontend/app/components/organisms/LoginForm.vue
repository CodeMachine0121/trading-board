<script setup lang="ts">
import { ref } from 'vue'
import { CredentialsDto } from '~/domain/models/dto/credentials-dto'
import AppAlert from '~/components/atoms/AppAlert.vue'
import AppButton from '~/components/atoms/AppButton.vue'
import AppInput from '~/components/atoms/AppInput.vue'

// 登入表單。帳密原樣交出去，對不對由板子說。
defineProps<{
  submitting: boolean
  errorMessage: string | null
}>()

const emit = defineEmits<{
  submit: [credentials: CredentialsDto]
}>()

const username = ref('')
const password = ref('')
</script>

<template>
  <form
    class="login-form"
    @submit.prevent="emit('submit', new CredentialsDto(username, password))"
  >
    <h1 class="login-form__title">
      Trading Board
    </h1>
    <AppInput
      v-model="username"
      label="帳號"
      autocomplete="username"
      required
      data-testid="login-username"
    />
    <AppInput
      v-model="password"
      label="密碼"
      type="password"
      autocomplete="current-password"
      required
      data-testid="login-password"
    />
    <AppAlert
      v-if="errorMessage !== null"
      variant="error"
      data-testid="login-error"
    >
      {{ errorMessage }}
    </AppAlert>
    <AppButton
      type="submit"
      block
      :loading="submitting"
    >
      登入
    </AppButton>
  </form>
</template>

<style scoped lang="scss">
.login-form {
  display: flex;
  flex-direction: column;
  gap: spacing('md');
  padding: spacing('lg');

  @include surface-card;

  &__title {
    color: color('text-strong');
    font-size: font-size('2xl');
    text-align: center;
  }
}
</style>
