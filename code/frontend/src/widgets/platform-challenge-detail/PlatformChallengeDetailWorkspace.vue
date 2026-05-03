<script setup lang="ts">
import type { AdminChallengeListItem } from '@/api/contracts'
import AdminChallengeTopbarPanel from '@/components/platform/challenge/AdminChallengeTopbarPanel.vue'
import AdminChallengeWorkspaceTabs from '@/components/platform/challenge/AdminChallengeWorkspaceTabs.vue'
import type { PlatformChallengeFlagDraft, PlatformChallengeFlagDraftPatch } from '@/features/platform-challenge-detail'

type ChallengePanelKey = 'detail' | 'writeup'

type ChallengePanelTab = {
  key: ChallengePanelKey
  label: string
  tabId: string
  panelId: string
}

defineProps<{
  workspaceLabel: string
  hasChallengeId: boolean
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
  openTopology: []
  openChallengeList: []
  select: [panel: ChallengePanelKey]
  keydown: [payload: { event: KeyboardEvent; index: number }]
  downloadAttachment: []
  saveFlagConfig: []
  updateFlagDraft: [value: PlatformChallengeFlagDraftPatch]
}>()
</script>

<template>
  <AdminChallengeTopbarPanel
    :workspace-label="workspaceLabel"
    :has-challenge-id="hasChallengeId"
    @open-topology="emit('openTopology')"
    @open-challenge-list="emit('openChallengeList')"
  />

  <AdminChallengeWorkspaceTabs
    :loading="loading"
    :panel-tabs="panelTabs"
    :active-panel="activePanel"
    :set-tab-button-ref="setTabButtonRef"
    :challenge="challenge"
    :downloading-attachment="downloadingAttachment"
    :flag-draft="flagDraft"
    :challenge-id="challengeId"
    @select="emit('select', $event)"
    @keydown="emit('keydown', $event)"
    @download-attachment="emit('downloadAttachment')"
    @save-flag-config="emit('saveFlagConfig')"
    @update:flag-draft="emit('updateFlagDraft', $event)"
  />
</template>
