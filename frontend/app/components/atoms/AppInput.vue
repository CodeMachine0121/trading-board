<script setup lang="ts">
import { useId } from 'vue'

// 全站唯一的輸入欄位：標籤、輸入框、說明一組。原生屬性（required、autocomplete、
// inputmode、min…）走 attribute fallthrough，落在 <input> 上。
defineOptions({ inheritAttrs: false })

withDefaults(defineProps<{
  label: string
  type?: 'text' | 'email' | 'password' | 'datetime-local'
  hint?: string
}>(), {
  type: 'text',
  hint: '',
})

const value = defineModel<string>({ required: true })
const inputId = useId()
</script>

<template>
  <div class="app-input">
    <label
      class="app-input__label"
      :for="inputId"
    >{{ label }}</label>
    <input
      :id="inputId"
      v-model="value"
      class="app-input__control"
      :type="type"
      v-bind="$attrs"
    >
    <p
      v-if="hint !== ''"
      class="app-input__hint"
    >
      {{ hint }}
    </p>
  </div>
</template>

<style scoped lang="scss">
.app-input {
  display: flex;
  flex-direction: column;
  gap: spacing('2xs');

  &__label {
    @include dense-label;
  }

  &__control {
    border: 1px solid color('border-strong');
    border-radius: radius('md');
    background-color: color('surface');
    padding: spacing('xs') spacing('sm');
    width: 100%;
    color: color('text-strong');

    @include tap-target;
    @include focus-ring;
  }

  &__hint {
    color: color('text-muted');
    font-size: font-size('xs');
  }
}
</style>
