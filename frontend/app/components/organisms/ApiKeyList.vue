<script setup lang="ts">
import { ref } from 'vue'
import type { ApiKeyDto } from '~/domain/models/dto/api-key-dto'
import AppBadge from '~/components/atoms/AppBadge.vue'
import AppButton from '~/components/atoms/AppButton.vue'

// 所有成員金鑰（新的在前）。撤銷要在畫面上再確認一次——撤銷之後無法復原。
defineProps<{
  apiKeys: readonly ApiKeyDto[]
  revokingId: string | null
}>()

const emit = defineEmits<{
  revoke: [apiKeyId: string]
}>()

const confirmingId = ref<string | null>(null)

function confirmRevoke(apiKeyId: string): void {
  confirmingId.value = null
  emit('revoke', apiKeyId)
}
</script>

<template>
  <p
    v-if="apiKeys.length === 0"
    class="api-key-list__empty"
    data-testid="api-key-list-empty"
  >
    還沒有任何金鑰
  </p>
  <ul
    v-else
    class="api-key-list"
  >
    <li
      v-for="apiKey in apiKeys"
      :key="apiKey.id"
      class="api-key-list__item"
      data-testid="api-key-row"
    >
      <div class="api-key-list__heading">
        <strong class="api-key-list__member">{{ apiKey.memberName }}</strong>
        <AppBadge
          :variant="apiKey.statusTone"
          data-testid="api-key-status"
        >
          {{ apiKey.statusLabel }}
        </AppBadge>
      </div>
      <p
        v-if="apiKey.note !== ''"
        class="api-key-list__note"
      >
        {{ apiKey.note }}
      </p>
      <dl class="api-key-list__details">
        <dt>金鑰開頭</dt>
        <dd><code>{{ apiKey.displayPrefixText }}</code></dd>
        <dt>建立於</dt>
        <dd>{{ apiKey.createdAtText }}</dd>
        <dt>最後使用</dt>
        <dd data-testid="api-key-last-used">
          {{ apiKey.lastUsedAtText }}
        </dd>
      </dl>
      <div
        v-if="apiKey.canRevoke"
        class="api-key-list__actions"
      >
        <template v-if="confirmingId === apiKey.id">
          <span class="api-key-list__confirm-text">撤銷後無法復原，確定？</span>
          <AppButton
            variant="danger"
            size="sm"
            :loading="revokingId === apiKey.id"
            data-testid="api-key-revoke-confirm"
            @click="confirmRevoke(apiKey.id)"
          >
            確定撤銷
          </AppButton>
          <AppButton
            variant="ghost"
            size="sm"
            data-testid="api-key-revoke-cancel"
            @click="confirmingId = null"
          >
            取消
          </AppButton>
        </template>
        <AppButton
          v-else
          variant="secondary"
          size="sm"
          data-testid="api-key-revoke"
          @click="confirmingId = apiKey.id"
        >
          撤銷
        </AppButton>
      </div>
    </li>
  </ul>
</template>

<style scoped lang="scss">
.api-key-list {
  display: flex;
  flex-direction: column;
  gap: spacing('sm');
  margin: 0;
  padding: 0;
  list-style: none;

  &__empty {
    color: color('text-muted');
  }

  &__item {
    display: flex;
    flex-direction: column;
    gap: spacing('xs');
    padding: spacing('md');

    @include surface-card;
  }

  &__heading {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: spacing('xs');
  }

  &__member {
    overflow-wrap: anywhere;
    color: color('text-strong');
  }

  &__note {
    color: color('text-muted');
    font-size: font-size('sm');
  }

  &__details {
    display: grid;
    grid-template-columns: auto 1fr;
    gap: spacing('3xs') spacing('sm');
    margin: 0;
    font-size: font-size('sm');
  }

  &__actions {
    display: flex;
    flex-wrap: wrap;
    justify-content: flex-end;
    align-items: center;
    gap: spacing('xs');
  }

  &__confirm-text {
    color: color('danger');
    font-size: font-size('sm');
  }
}
</style>
