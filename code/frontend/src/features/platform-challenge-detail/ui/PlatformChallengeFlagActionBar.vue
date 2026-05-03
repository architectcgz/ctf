<script setup lang="ts">
import type { PlatformChallengeFlagDraft } from '../model'

interface Props {
  draft: PlatformChallengeFlagDraft
}

defineProps<Props>()

const emit = defineEmits<{
  save: []
}>()
</script>

<template>
  <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
    <div class="text-sm text-[var(--journal-muted)]">
      当前配置：{{ draft.flagConfigSummary }}
    </div>
    <button
      :disabled="draft.saving"
      class="ui-btn ui-btn--primary"
      type="button"
      @click="emit('save')"
    >
      {{ draft.saving ? '保存中...' : '保存配置' }}
    </button>
  </div>
</template>

<style scoped>
.ui-btn {
  --ui-btn-height: 2.45rem;
  --ui-btn-radius: 0.75rem;
  --ui-btn-padding: var(--space-2) var(--space-4);
  --ui-btn-font-size: var(--font-size-0-875);
  --ui-btn-font-weight: 600;
  --ui-btn-primary-border: color-mix(in srgb, var(--journal-accent) 18%, transparent);
  --ui-btn-primary-background: var(--journal-accent);
  --ui-btn-primary-hover-background: color-mix(in srgb, var(--journal-accent) 88%, var(--color-bg-base));
  --ui-btn-ghost-color: var(--journal-ink);
  --ui-btn-ghost-hover-color: var(--journal-accent);
  --ui-btn-ghost-hover-background: color-mix(in srgb, var(--journal-accent) 4%, var(--journal-surface));
  --ui-btn-focus-ring: color-mix(in srgb, var(--journal-accent) 18%, transparent);
}
</style>
