<script setup lang="ts">
import type { AdminChallengeListItem, FlagType } from '@/api/contracts'
import AdminChallengeProfilePanel from '@/components/platform/challenge/AdminChallengeProfilePanel.vue'
import ChallengeWriteupManagePanel from '@/components/platform/writeup/ChallengeWriteupManagePanel.vue'

type ChallengePanelKey = 'detail' | 'writeup'

type ChallengePanelTab = {
  key: ChallengePanelKey
  label: string
  tabId: string
  panelId: string
}

defineProps<{
  loading: boolean
  panelTabs: ReadonlyArray<ChallengePanelTab>
  activePanel: ChallengePanelKey
  setTabButtonRef: (key: ChallengePanelKey, element: HTMLButtonElement | null) => void
  challenge: AdminChallengeListItem | null
  downloadingAttachment: boolean
  flagConfigSummary: string
  flagDraftSummary: string
  flagType: FlagType
  flagValue: string
  flagRegex: string
  flagPrefix: string
  saving: boolean
  isSharedInstanceChallenge: boolean
  challengeId: string
}>()

const emit = defineEmits<{
  select: [panel: ChallengePanelKey]
  keydown: [payload: { event: KeyboardEvent; index: number }]
  downloadAttachment: []
  saveFlagConfig: []
  'update:flag-type': [value: FlagType]
  'update:flag-value': [value: string]
  'update:flag-regex': [value: string]
  'update:flag-prefix': [value: string]
}>()

function handleSelect(panel: ChallengePanelKey): void {
  emit('select', panel)
}

function handleTabKeydown(event: KeyboardEvent, index: number): void {
  emit('keydown', { event, index })
}
</script>

<template>
  <nav
    class="top-tabs"
    role="tablist"
    aria-label="题目管理视图切换"
  >
    <button
      v-for="(tab, index) in panelTabs"
      :id="tab.tabId"
      :key="tab.key"
      :ref="(element) => setTabButtonRef(tab.key, element as HTMLButtonElement | null)"
      type="button"
      role="tab"
      class="top-tab"
      :class="{ active: activePanel === tab.key }"
      :aria-selected="activePanel === tab.key ? 'true' : 'false'"
      :aria-controls="tab.panelId"
      :tabindex="activePanel === tab.key ? 0 : -1"
      @click="handleSelect(tab.key)"
      @keydown="handleTabKeydown($event, index)"
    >
      {{ tab.label }}
    </button>
  </nav>

  <main class="content-pane">
    <div
      v-if="loading"
      class="flex items-center justify-center py-12"
    >
      <div
        class="h-8 w-8 animate-spin rounded-full border-4 border-[var(--journal-border)] border-t-[var(--journal-accent)]"
      />
    </div>

    <template v-else-if="challenge">
      <AdminChallengeProfilePanel
        v-show="activePanel === 'detail'"
        id="admin-challenge-panel-detail"
        class="tab-panel challenge-panel"
        role="tabpanel"
        aria-labelledby="admin-challenge-tab-detail"
        :aria-hidden="activePanel === 'detail' ? 'false' : 'true'"
        :challenge="challenge"
        :downloading-attachment="downloadingAttachment"
        :flag-config-summary="flagConfigSummary"
        :flag-draft-summary="flagDraftSummary"
        :flag-type="flagType"
        :flag-value="flagValue"
        :flag-regex="flagRegex"
        :flag-prefix="flagPrefix"
        :saving="saving"
        :is-shared-instance-challenge="isSharedInstanceChallenge"
        @download-attachment="emit('downloadAttachment')"
        @save-flag-config="emit('saveFlagConfig')"
        @update:flag-type="emit('update:flag-type', $event)"
        @update:flag-value="emit('update:flag-value', $event)"
        @update:flag-regex="emit('update:flag-regex', $event)"
        @update:flag-prefix="emit('update:flag-prefix', $event)"
      />

      <section
        v-show="activePanel === 'writeup'"
        id="admin-challenge-panel-writeup"
        class="tab-panel challenge-panel"
        role="tabpanel"
        aria-labelledby="admin-challenge-tab-writeup"
        :aria-hidden="activePanel === 'writeup' ? 'false' : 'true'"
      >
        <ChallengeWriteupManagePanel
          :challenge-id="challengeId"
          :challenge-title="challenge.title"
        />
      </section>
    </template>
  </main>
</template>

<style scoped>
.content-pane {
  display: flex;
  flex: 1 1 auto;
  flex-direction: column;
}

.tab-panel {
  display: grid;
  gap: var(--space-5);
}

.challenge-panel {
  padding-top: var(--space-6);
}
</style>
