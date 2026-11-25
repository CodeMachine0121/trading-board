<script setup lang="ts">
import { onMounted } from 'vue'
import AppAlert from '~/components/atoms/AppAlert.vue'
import AppSpinner from '~/components/atoms/AppSpinner.vue'
import ApiKeyCreateForm from '~/components/molecules/ApiKeyCreateForm.vue'
import CreatedApiKeyPanel from '~/components/molecules/CreatedApiKeyPanel.vue'
import ApiKeyList from '~/components/organisms/ApiKeyList.vue'
import { useApiKeyScreen } from '~/composables/use-api-key-screen'

// 「API 金鑰」：替每位顧問團隊成員建立自己的金鑰、看清單、撤銷。狀態與「只顯示一次」的規則在 useApiKeyScreen。
const { $apiKeyApplication } = useNuxtApp()
const {
  apiKeys, loading, errorMessage, load, creating, createErrorMessage, createdKey, copied, copyErrorMessage,
  revokingId, revokeErrorMessage, createApiKey, copyKey, revokeApiKey,
} = useApiKeyScreen($apiKeyApplication)

onMounted(() => void load())
</script>

<template>
  <div class="api-keys-page">
    <h2 class="api-keys-page__title">
      API 金鑰
    </h2>
    <ApiKeyCreateForm
      :submitting="creating"
      :error-message="createErrorMessage"
      @submit="createApiKey"
    />
    <CreatedApiKeyPanel
      v-if="createdKey !== null"
      :created-key="createdKey"
      :copied="copied"
      :copy-error-message="copyErrorMessage"
      @copy="copyKey"
    />
    <AppAlert
      v-if="revokeErrorMessage !== null"
      variant="error"
    >
      {{ revokeErrorMessage }}
    </AppAlert>
    <AppSpinner v-if="loading && apiKeys === null" />
    <AppAlert
      v-else-if="errorMessage !== null"
      variant="error"
    >
      {{ errorMessage }}
    </AppAlert>
    <ApiKeyList
      v-else-if="apiKeys !== null"
      :api-keys="apiKeys"
      :revoking-id="revokingId"
      @revoke="revokeApiKey"
    />
  </div>
</template>

<style scoped lang="scss">
.api-keys-page {
  display: flex;
  flex-direction: column;
  gap: spacing('md');

  &__title {
    color: color('text-strong');
    font-size: font-size('xl');
  }
}
</style>
