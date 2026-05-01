<template>
  <section
    class="workspace-shell journal-shell journal-shell-admin journal-hero flex min-h-full flex-1 flex-col"
  >
    <AdminChallengeTopbarPanel
      :workspace-label="workspaceLabel"
      :has-challenge-id="Boolean(challengeId)"
      @open-topology="openTopology"
      @open-challenge-list="openChallengeList"
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
      @select="switchPanel"
      @keydown="handleTabKeydown($event.event, $event.index)"
      @download-attachment="downloadAttachment"
      @save-flag-config="saveFlagConfig"
      @update:flag-draft="updateFlagDraft"
    />
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import AdminChallengeTopbarPanel from '@/components/platform/challenge/AdminChallengeTopbarPanel.vue'
import AdminChallengeWorkspaceTabs from '@/components/platform/challenge/AdminChallengeWorkspaceTabs.vue'
import { useRouteQueryTabs } from '@/composables/useRouteQueryTabs'
import { usePlatformChallengeDetailPage } from '@/features/platform-challenge-detail'

type ChallengePanelKey = 'detail' | 'writeup'

const panelTabs = [
  {
    key: 'detail' as const,
    label: '题目管理',
    tabId: 'admin-challenge-tab-detail',
    panelId: 'admin-challenge-panel-detail',
  },
  {
    key: 'writeup' as const,
    label: '题解管理',
    tabId: 'admin-challenge-tab-writeup',
    panelId: 'admin-challenge-panel-writeup',
  },
]

const route = useRoute()
const router = useRouter()
const panelTabOrder = panelTabs.map((tab) => tab.key) as ChallengePanelKey[]
const {
  challenge,
  challengeId,
  downloadingAttachment,
  downloadAttachment,
  flagDraft,
  loading,
  openChallengeList,
  openTopology,
  saveFlagConfig,
  updateFlagDraft,
  workspaceLabel,
} = usePlatformChallengeDetailPage()
const {
  activeTab: activePanel,
  setTabButtonRef,
  selectTab: switchPanel,
  handleTabKeydown,
} = useRouteQueryTabs<ChallengePanelKey>({
  route,
  router,
  orderedTabs: panelTabOrder,
  defaultTab: 'detail',
  routeName: 'PlatformChallengeDetail',
  routeParams: route.params,
})
</script>

<style scoped>
.journal-shell {
  --workspace-topbar-tabs-gap: 0;
  --workspace-tabs-offset-top: var(--workspace-topbar-tabs-gap);
  --workspace-tabs-panel-gap: var(--space-2);
  --journal-topbar-padding-bottom: var(--workspace-topbar-tabs-gap);
  --page-top-tabs-gap: var(--space-7);
  --page-top-tabs-margin: 0 calc(var(--space-6) * -1) 0;
  --page-top-tabs-padding: 0 var(--space-6);
  --page-top-tabs-border: color-mix(in srgb, var(--journal-ink) 10%, transparent);
  --page-top-tab-min-height: 42px;
  --page-top-tab-padding: var(--space-1-5) 0 var(--space-2);
  --page-top-tab-font-size: var(--font-size-14);
  --page-top-tab-active-color: color-mix(in srgb, var(--journal-accent) 74%, var(--journal-ink));
  --page-top-tab-active-border: color-mix(in srgb, var(--journal-accent) 86%, var(--journal-ink));
  --journal-shell-hero-radial-strength: 7%;
  --journal-shell-hero-radial-size: 22rem;
  --journal-shell-hero-end: var(--journal-surface);
  --journal-shell-hero-shadow: 0 22px 50px var(--color-shadow-soft);
}

</style>
