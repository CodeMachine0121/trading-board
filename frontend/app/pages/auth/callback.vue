<script setup lang="ts">
import { onMounted } from 'vue'
import { SignInReturnDto } from '~/domain/models/dto/sign-in-return-dto'
import AppAlert from '~/components/atoms/AppAlert.vue'
import AppSpinner from '~/components/atoms/AppSpinner.vue'
import { LOGIN_PATH, useSignInFlow } from '~/composables/use-sign-in-flow'

definePageMeta({ layout: false })

// 板子把登入結果轉回這裡：比對標記、拿一次性授權碼換通行憑證，成功就進首頁。
const route = useRoute()
const { errorMessage, completeSignIn } = useSignInFlow(useNuxtApp().$userSessionApplication, navigateTo)

onMounted(() => completeSignIn(new SignInReturnDto(
  typeof route.query.code === 'string' ? route.query.code : '',
  typeof route.query.state === 'string' ? route.query.state : '',
)))
</script>

<template>
  <main class="sign-in-callback">
    <template v-if="errorMessage !== null">
      <AppAlert
        variant="error"
        data-testid="callback-error"
      >
        {{ errorMessage }}
      </AppAlert>
      <NuxtLink
        :to="LOGIN_PATH"
        class="sign-in-callback__link"
        data-testid="callback-sign-in"
      >
        重新登入
      </NuxtLink>
    </template>
    <AppSpinner v-else />
  </main>
</template>

<style scoped lang="scss">
.sign-in-callback {
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  gap: spacing('md');
  margin: 0 auto;
  padding: spacing('md');
  min-height: 100vh;
  max-width: control-size('form-max');

  &__link {
    color: color('primary');
    font-weight: font-weight('medium');

    @include tap-target;
    @include focus-ring;
  }
}
</style>
