<script setup lang="ts">
import LoginForm from '~/components/organisms/LoginForm.vue'
import { useSignInFlow } from '~/composables/use-sign-in-flow'

definePageMeta({ layout: false })

// 帳密對了，板子回一個要轉去的網址（帶一次性授權碼）；整頁跳過去，由 /auth/callback 收尾。
const { submitting, errorMessage, signIn } = useSignInFlow(useNuxtApp().$userSessionApplication, navigateTo)
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
