<script setup lang="ts">
import { ref } from 'vue'
import type { CredentialsDto } from '~/domain/models/dto/credentials-dto'
import { failureMessageOf } from '~/composables/failure-message'
import LoginForm from '~/components/organisms/LoginForm.vue'

definePageMeta({ layout: false })

const { $userSessionApplication } = useNuxtApp()
const submitting = ref(false)
const errorMessage = ref<string | null>(null)

// 帳密對了，板子回一個要轉去的網址（帶一次性授權碼）；整頁跳過去，由 /auth/callback 收尾。
async function signIn(credentialsDto: CredentialsDto): Promise<void> {
  submitting.value = true
  errorMessage.value = null

  try {
    const redirectTo = await $userSessionApplication.beginSignIn(credentialsDto)
    await navigateTo(redirectTo, { external: true })
  }
  catch (error: unknown) {
    errorMessage.value = failureMessageOf(error)
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
