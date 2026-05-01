<script setup lang="ts">
import type { FlagType } from '@/api/contracts'
import type { PlatformChallengeFlagDraft } from '../model'

interface Props {
  draft: PlatformChallengeFlagDraft
}

defineProps<Props>()

const emit = defineEmits<{
  'update:draft': [value: Partial<Pick<PlatformChallengeFlagDraft, 'flagPrefix' | 'flagRegex' | 'flagType' | 'flagValue'>>]
}>()

function updateFlagType(event: Event): void {
  emit('update:draft', { flagType: (event.target as HTMLSelectElement).value as FlagType })
}

function updateFlagValue(event: Event): void {
  emit('update:draft', { flagValue: (event.target as HTMLInputElement).value })
}

function updateFlagRegex(event: Event): void {
  emit('update:draft', { flagRegex: (event.target as HTMLInputElement).value })
}

function updateFlagPrefix(event: Event): void {
  emit('update:draft', { flagPrefix: (event.target as HTMLInputElement).value })
}
</script>

<template>
  <div class="grid gap-4 md:grid-cols-2">
    <label class="flag-field">
      <span class="flag-field-label">判题模式</span>
      <select
        :value="draft.flagType"
        class="flag-field-input"
        @change="updateFlagType"
      >
        <option value="static">静态 Flag</option>
        <option value="dynamic">动态前缀</option>
        <option value="regex">正则匹配</option>
        <option value="manual_review">人工审核</option>
      </select>
    </label>

    <label
      v-if="draft.flagType === 'dynamic' || draft.flagType === 'regex'"
      class="flag-field"
    >
      <span class="flag-field-label">Flag 前缀</span>
      <input
        :value="draft.flagPrefix"
        type="text"
        placeholder="例如：flag"
        class="flag-field-input"
        @input="updateFlagPrefix"
      >
    </label>

    <label
      v-if="draft.flagType === 'static'"
      class="flag-field md:col-span-2"
    >
      <span class="flag-field-label">静态 Flag</span>
      <input
        :value="draft.flagValue"
        type="text"
        placeholder="例如：flag{demo}"
        class="flag-field-input font-mono"
        @input="updateFlagValue"
      >
    </label>

    <label
      v-if="draft.flagType === 'regex'"
      class="flag-field md:col-span-2"
    >
      <span class="flag-field-label">正则表达式</span>
      <input
        :value="draft.flagRegex"
        type="text"
        placeholder="例如：^flag\\{demo-[0-9]+\\}$"
        class="flag-field-input font-mono"
        @input="updateFlagRegex"
      >
    </label>
  </div>
</template>

<style scoped>
.flag-field {
  display: flex;
  flex-direction: column;
  gap: var(--space-2-5);
}

.flag-field-label {
  font-size: var(--font-size-0-82);
  font-weight: 600;
  color: var(--journal-ink);
}

.flag-field-input {
  min-height: 2.9rem;
  border: 1px solid var(--journal-border);
  border-radius: 1rem;
  background: color-mix(in srgb, var(--journal-surface) 96%, transparent);
  padding: var(--space-3) var(--space-4);
  font-size: var(--font-size-0-92);
  color: var(--journal-ink);
  outline: none;
  transition:
    border-color 0.15s ease,
    box-shadow 0.15s ease;
}

.flag-field-input:focus {
  border-color: color-mix(in srgb, var(--journal-accent) 42%, transparent);
  box-shadow: 0 0 0 4px color-mix(in srgb, var(--journal-accent) 12%, transparent);
}
</style>
