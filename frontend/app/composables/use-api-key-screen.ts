import { ref, shallowRef } from 'vue'
import type { ApiKeyApplication } from '~/application/api-key-application'
import type { ApiKeyCreateDto } from '~/domain/models/dto/api-key-create-dto'
import type { CreatedApiKeyDto } from '~/domain/models/dto/created-api-key-dto'
import { failureMessageOf } from '~/composables/failure-message'
import { useAsyncView } from '~/composables/use-async-view'

/**
 * 「API 金鑰」畫面的全部狀態：清單、建立、只出現一次的完整金鑰、複製、撤銷。
 *
 * 完整金鑰只放在這裡：一開始建立下一把就先清掉，離開畫面（這個狀態被丟掉）它也就不在了——
 * 「只顯示一次」這條規則只寫在這一處，頁面只負責把狀態綁到元件上。
 */
export function useApiKeyScreen(apiKeyApplication: ApiKeyApplication) {
  const listView = useAsyncView(() => apiKeyApplication.listApiKeys())
  const creating = ref(false)
  const createErrorMessage = ref<string | null>(null)
  const createdKey = shallowRef<CreatedApiKeyDto | null>(null)
  const copied = ref(false)
  const copyErrorMessage = ref<string | null>(null)
  const revokingId = ref<string | null>(null)
  const revokeErrorMessage = ref<string | null>(null)

  async function createApiKey(apiKeyCreateDto: ApiKeyCreateDto): Promise<void> {
    creating.value = true
    createErrorMessage.value = null
    createdKey.value = null
    copied.value = false
    copyErrorMessage.value = null

    try {
      createdKey.value = await apiKeyApplication.createApiKey(apiKeyCreateDto)
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
      await apiKeyApplication.copyKey(key)
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
      const revoked = await apiKeyApplication.revokeApiKey(apiKeyId)
      listView.data.value = listView.data.value?.map(apiKey => apiKey.id === revoked.id ? revoked : apiKey) ?? null
    }
    catch (error: unknown) {
      revokeErrorMessage.value = failureMessageOf(error)
    }
    finally {
      revokingId.value = null
    }
  }

  return {
    apiKeys: listView.data,
    loading: listView.loading,
    errorMessage: listView.errorMessage,
    load: listView.reload,
    creating,
    createErrorMessage,
    createdKey,
    copied,
    copyErrorMessage,
    revokingId,
    revokeErrorMessage,
    createApiKey,
    copyKey,
    revokeApiKey,
  }
}
