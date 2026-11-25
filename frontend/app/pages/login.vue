<script setup lang="ts">
import { ref } from 'vue'
import type { CredentialsDto } from '~/domain/models/dto/credentials-dto'
import { TradingBoardError } from '~/domain/errors/trading-board-error'
import LoginForm from '~/components/organisms/LoginForm.vue'

definePageMeta({ layout: false })

const { $userSessionApplication } = useNuxtApp()
const submitting = ref(false)
const errorMessage = ref<string | null>(null)

async function signIn(credentialsDto: CredentialsDto): Promise<void> {
  submitting.value = true
  errorMessage.value = null

  try {
    await $userSessionApplication.signIn(credentialsDto)
    await navigateTo('/')
  }
  catch (error: unknown) {
    if (!(error instanceof TradingBoardError)) {
      throw error
    }
    errorMessage.value = error.message
  }
  finally {
    submitting.value = false
  }
}
</script>

<template>
  <main class="login-page">
    <LoginForm
      :submitting="submitting"
      :error-message="errorMessage"
      @submit="signIn"
    />
  </main>
</template>

<style scoped lang="scss">
.login-page {
  display: flex;
  justify-content: center;
  align-items: center;
  margin: 0 auto;
  padding: spacing('md');
  min-height: 100vh;
  max-width: control-size('form-max');
}
</style>
