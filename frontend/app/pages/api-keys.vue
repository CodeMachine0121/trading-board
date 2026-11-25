<script setup lang="ts">
import { onMounted, ref, shallowRef } from 'vue'
import type { ApiKeyCreateDto } from '~/domain/models/dto/api-key-create-dto'
import type { CreatedApiKeyDto } from '~/domain/models/dto/created-api-key-dto'
import AppAlert from '~/components/atoms/AppAlert.vue'
import AppSpinner from '~/components/atoms/AppSpinner.vue'
import ApiKeyCreateForm from '~/components/molecules/ApiKeyCreateForm.vue'
import CreatedApiKeyPanel from '~/components/molecules/CreatedApiKeyPanel.vue'
import ApiKeyList from '~/components/organisms/ApiKeyList.vue'
import { failureMessageOf } from '~/composables/failure-message'
import { useAsyncView } from '~/composables/use-async-view'

// 「API 金鑰」：替每位顧問團隊成員建立自己的金鑰、看清單、撤銷。
// 完整金鑰只放在這一頁的狀態裡：建立下一把或離開這一頁，它就不在了。
const { $apiKeyApplication } = useNuxtApp()
const listView = useAsyncView(() => $apiKeyApplication.listApiKeys())
const { data: apiKeys, loading, errorMessage } = listView

const creating = ref(false)
const createErrorMessage = ref<string | null>(null)
const createdKey = shallowRef<CreatedApiKeyDto | null>(null)
const copied = ref(false)
const copyErrorMessage = ref<string | null>(null)
const revokingId = ref<string | null>(null)
const revokeErrorMessage = ref<string | null>(null)

onMounted(() => void listView.reload())

async function createApiKey(apiKeyCreateDto: ApiKeyCreateDto): Promise<void> {
  creating.value = true
  createErrorMessage.value = null
  createdKey.value = null
  copied.value = false
  copyErrorMessage.value = null

  try {
    createdKey.value = await $apiKeyApplication.createApiKey(apiKeyCreateDto)
    await listView.reload()
  }
  catch (error: unknown) {
    createErrorMessage.value = failureMessageOf(error)
  }
  finally {
    creating.value = false
  }
}

async function copyKey(key: string): Promise<void> {
  copyErrorMessage.value = null
  try {
    await $apiKeyApplication.copyKey(key)
    copied.value = true
  }
  catch (error: unknown) {
    copyErrorMessage.value = failureMessageOf(error)
  }
}

async function revokeApiKey(apiKeyId: string): Promise<void> {
  revokingId.value = apiKeyId
  revokeErrorMessage.value = null

  try {
    const revoked = await $apiKeyApplication.revokeApiKey(apiKeyId)
    apiKeys.value = (apiKeys.value ?? []).map(apiKey => apiKey.id === revoked.id ? revoked : apiKey)
  }
  catch (error: unknown) {
    revokeErrorMessage.value = failureMessageOf(error)
  }
  finally {
    revokingId.value = null
  }
}
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
