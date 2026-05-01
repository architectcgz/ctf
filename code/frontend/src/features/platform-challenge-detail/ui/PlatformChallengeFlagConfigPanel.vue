<script setup lang="ts">
import type { PlatformChallengeFlagDraft } from '../model'
import PlatformChallengeFlagActionBar from './PlatformChallengeFlagActionBar.vue'
import PlatformChallengeFlagFieldGrid from './PlatformChallengeFlagFieldGrid.vue'
import PlatformChallengeFlagNoticeStack from './PlatformChallengeFlagNoticeStack.vue'

interface Props {
  draft: PlatformChallengeFlagDraft
}

defineProps<Props>()

const emit = defineEmits<{
  save: []
  'update:draft': [value: Partial<Pick<PlatformChallengeFlagDraft, 'flagPrefix' | 'flagRegex' | 'flagType' | 'flagValue'>>]
}>()
</script>

<template>
  <section class="journal-panel challenge-flag-panel p-5 md:p-6">
    <div class="flex flex-col gap-3 md:flex-row md:items-start md:justify-between">
      <p class="challenge-flag-panel__copy">
        支持静态 Flag、动态前缀、正则判题和人工审核四种模式。保存后即时刷新当前题目配置。
      </p>
      <div class="flag-summary-chip">
        {{ draft.flagDraftSummary }}
      </div>
    </div>

    <PlatformChallengeFlagFieldGrid
      :draft="draft"
      @update:draft="emit('update:draft', $event)"
    />
    <PlatformChallengeFlagNoticeStack :draft="draft" />
    <PlatformChallengeFlagActionBar
      :draft="draft"
      @save="emit('save')"
    />
  </section>
</template>

<style scoped>
.challenge-flag-panel {
  display: grid;
  gap: var(--space-5);
}

.challenge-flag-panel__copy {
  margin: var(--space-2) 0 0;
  font-size: var(--font-size-0-88);
  line-height: 1.7;
  color: var(--journal-muted);
}

.flag-summary-chip {
  display: inline-flex;
  align-items: center;
  border-radius: 999px;
  border: 1px solid color-mix(in srgb, var(--journal-accent) 20%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 10%, transparent);
  padding: var(--space-2) var(--space-3-5);
  font-size: var(--font-size-0-80);
  font-weight: 600;
  color: var(--journal-accent);
}
</style>
