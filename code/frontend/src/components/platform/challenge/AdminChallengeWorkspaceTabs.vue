<script setup lang="ts">
import type { AdminChallengeListItem } from '@/api/contracts'
import AdminChallengeProfilePanel from '@/components/platform/challenge/AdminChallengeProfilePanel.vue'
import ChallengeWriteupManagePanel from '@/components/platform/writeup/ChallengeWriteupManagePanel.vue'
import type { PlatformChallengeFlagDraft } from '@/features/platform-challenge-detail'

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
  flagDraft: PlatformChallengeFlagDraft
  challengeId: string
}>()

const emit = defineEmits<{
  select: [panel: ChallengePanelKey]
  keydown: [payload: { event: KeyboardEvent; index: number }]
  downloadAttachment: []
  saveFlagConfig: []
  'update:flag-draft': [value: Partial<Pick<PlatformChallengeFlagDraft, 'flagPrefix' | 'flagRegex' | 'flagType' | 'flagValue'>>]
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
        :flag-draft="flagDraft"
        @download-attachment="emit('downloadAttachment')"
        @save-flag-config="emit('saveFlagConfig')"
        @update:flag-draft="emit('update:flag-draft', $event)"
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
  padding-top: var(--workspace-tabs-panel-gap);
}

.tab-panel {
  display: grid;
  gap: var(--space-5);
}

.challenge-panel {
  padding-top: 0;
}
</style>
