<script setup lang="ts">
import AppButton from '~/components/atoms/AppButton.vue'
import AppShell from '~/components/templates/AppShell.vue'
import { LOGIN_PATH } from '~/plugins/dependencies'

// 登入後每一頁共用的外框：導覽四個畫面，右上角登出。
const { $userSessionApplication } = useNuxtApp()

const destinations = [
  { to: '/', label: '今日' },
  { to: '/history', label: '歷史' },
  { to: '/positions', label: '持股' },
  { to: '/summary', label: '月報' },
]

async function signOut(): Promise<void> {
  $userSessionApplication.signOut()
  await navigateTo(LOGIN_PATH)
}
</script>

<template>
  <AppShell>
    <template #brand>
      Trading Board
    </template>
    <template #actions>
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
.layout-nav__link {
  flex: 1;
  border-bottom: 2px solid transparent;
  padding: spacing('xs') spacing('sm');
  color: color('text-muted');
  font-weight: font-weight('medium');
  text-align: center;
  text-decoration: none;

  @include tap-target;
  @include focus-ring;

  &--active {
    border-bottom-color: color('primary');
    color: color('primary');
  }
}
</style>
