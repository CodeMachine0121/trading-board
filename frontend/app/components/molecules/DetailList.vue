<script setup lang="ts">
import type { DetailRowDto } from '~/domain/models/dto/detail-row-dto'

// 一組「名稱：值」。要列哪幾列、什麼語氣，都由 Domain 決定好放在 DTO 裡。
defineProps<{
  rows: readonly DetailRowDto[]
}>()
</script>

<template>
  <dl class="detail-list">
    <div
      v-for="row in rows"
      :key="row.label"
      class="detail-list__row"
    >
      <dt class="detail-list__label">
        {{ row.label }}
      </dt>
      <dd
        class="detail-list__value"
        :class="row.tone === null ? null : `detail-list__value--${row.tone}`"
      >
        {{ row.value }}
      </dd>
    </div>
  </dl>
</template>

<style scoped lang="scss">
.detail-list {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: spacing('xs') spacing('md');

  @include respond-to('md') {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }

  &__row {
    display: flex;
    flex-direction: column;
  }

  &__label {
    @include dense-label;
  }

  &__value {
    color: color('text-strong');
    font-weight: font-weight('medium');

    @include numeric;

    &--profit {
      color: color('profit');
    }

    &--loss {
      color: color('loss');
    }

    &--warning {
      color: color('warning');
    }
  }
}
</style>
