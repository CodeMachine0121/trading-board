<script setup lang="ts">
import { onMounted, ref } from 'vue'
import AppButton from '~/components/atoms/AppButton.vue'
import AppShell from '~/components/templates/AppShell.vue'
import { useSignInFlow } from '~/composables/use-sign-in-flow'

// 登入後每一頁共用的外框：導覽五個畫面，右上角是登入中的電子郵件與登出。
const { $userSessionApplication } = useNuxtApp()
const { signOut } = useSignInFlow($userSessionApplication, navigateTo)
const ownerEmail = ref('')

const destinations = [
  { to: '/', label: '今日' },
  { to: '/history', label: '歷史' },
  { to: '/positions', label: '持股' },
  { to: '/summary', label: '月報' },
  { to: '/api-keys', label: 'API 金鑰' },
]

onMounted(async () => {
  try {
    ownerEmail.value = (await $userSessionApplication.fetchCurrentOwner()).email
  }
  catch (error: unknown) {
    // 頁首少了電子郵件不妨礙使用；登入失效的話，請求那一層已經把人送回登入畫面。
    console.warn('could not read the signed-in owner', error)
  }
})
</script>

<template>
  <AppShell>
    <template #brand>
      Trading Board
    </template>
    <template #actions>
      <span
        v-if="ownerEmail !== ''"
        class="layout-owner"
        data-testid="owner-email"
      >{{ ownerEmail }}</span>
      <AppButton
        variant="ghost"
        size="sm"
        data-testid="sign-out"
        @click="signOut"
      >
        登出
      </AppButton>
    </template>
    <template #nav>
      <NuxtLink
        v-for="destination in destinations"
        :key="destination.to"
        :to="destination.to"
        class="layout-nav__link"
        active-class="layout-nav__link--active"
        :exact-active-class="destination.to === '/' ? 'layout-nav__link--active' : undefined"
      >
        {{ destination.label }}
      </NuxtLink>
    </template>
    <slot />
  </AppShell>
</template>

<style scoped lang="scss">
.layout-owner {
  overflow: hidden;
  max-width: control-size('owner-email-max');
  color: color('text-muted');
  font-size: font-size('sm');
  text-overflow: ellipsis;
  white-space: nowrap;
}

.layout-nav__link {
  flex: 1;
  border-bottom: 2px solid transparent;
  padding: spacing('xs') spacing('sm');
  color: color('text-muted');
  font-weight: font-weight('medium');
  text-align: center;
  text-decoration: none;
  white-space: nowrap;

  @include tap-target;
  @include focus-ring;

  &--active {
    border-bottom-color: color('primary');
    color: color('primary');
  }
}
</style>
