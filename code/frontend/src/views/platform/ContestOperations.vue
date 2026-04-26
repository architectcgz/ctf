<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { getContest } from '@/api/admin'
import type { ContestDetailData } from '@/api/contracts'
import AWDOperationsPanel from '@/components/platform/contest/AWDOperationsPanel.vue'
import AppLoading from '@/components/common/AppLoading.vue'
import { useBackofficeBreadcrumbDetail } from '@/composables/useBackofficeBreadcrumbDetail'
import { useRouteQueryTabs } from '@/composables/useRouteQueryTabs'
import { useToast } from '@/composables/useToast'

type ContestOperationsPanelKey = 'inspector' | 'instances'

const panelTabs = [
  {
    key: 'inspector' as const,
    label: '轮次态势',
    tabId: 'contest-ops-tab-inspector',
    panelId: 'contest-ops-panel-inspector',
  },
  {
    key: 'instances' as const,
    label: '实例编排',
    tabId: 'contest-ops-tab-instances',
    panelId: 'contest-ops-panel-instances',
  },
]

const route = useRoute()
const router = useRouter()
const toast = useToast()
const { setBreadcrumbDetailTitle } = useBackofficeBreadcrumbDetail()

const contestId = computed(() => String(route.params.id ?? ''))
const panelTabOrder = panelTabs.map((tab) => tab.key) as ContestOperationsPanelKey[]
const {
  activeTab: activePanel,
  setTabButtonRef,
  selectTab: switchPanel,
  handleTabKeydown,
} = useRouteQueryTabs<ContestOperationsPanelKey>({
  route,
  router,
  orderedTabs: panelTabOrder,
  defaultTab: 'inspector',
  routeName: 'ContestOperations',
  routeParams: route.params,
})

const loading = ref(true)
const contest = ref<ContestDetailData | null>(null)
const runtimeStageReady = computed(
  () =>
    contest.value?.status === 'running' ||
    contest.value?.status === 'frozen' ||
    contest.value?.status === 'ended'
)
const inspectorRuntimeContent = computed(() =>
  runtimeStageReady.value ? 'round-inspector' : 'readiness'
)

async function loadContest() {
  if (!contestId.value) {
    setBreadcrumbDetailTitle()
    return
  }
  loading.value = true
  try {
    contest.value = await getContest(contestId.value)
    setBreadcrumbDetailTitle(contest.value.title)
  } catch (err) {
    setBreadcrumbDetailTitle()
    toast.error('加载竞赛信息失败')
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  void loadContest()
})

onUnmounted(() => {
  setBreadcrumbDetailTitle()
})
</script>

<template>
  <section
    class="contest-ops-shell workspace-shell journal-shell journal-shell-admin journal-notes-card journal-hero flex min-h-full flex-1 flex-col"
  >
    <div
      v-if="loading"
      class="ops-loading-overlay"
    >
      <AppLoading>正在建立指挥链路...</AppLoading>
    </div>

    <div class="workspace-grid">
      <main class="content-pane contest-ops-content">
        <section
          v-if="contest"
          class="workspace-directory-section contest-ops-workspace"
        >
          <nav
            class="top-tabs"
            role="tablist"
            aria-label="AWD 运维视图切换"
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
              @click="switchPanel(tab.key)"
              @keydown="handleTabKeydown($event, index)"
            >
              {{ tab.label }}
            </button>
          </nav>

          <section
            v-if="activePanel === 'inspector'"
            id="contest-ops-panel-inspector"
            class="tab-panel contest-ops-tab-panel active"
            role="tabpanel"
            aria-labelledby="contest-ops-tab-inspector"
          >
            <AWDOperationsPanel
              :key="`${contest.id}-inspector`"
              :contests="[contest]"
              :selected-contest-id="contest.id"
              :hide-contest-selector="true"
              :hide-studio-link="true"
              :hide-readiness-actions="true"
              :hide-operation-tabs="true"
              operation-panel="inspector"
              :runtime-content="inspectorRuntimeContent"
            />
          </section>

          <section
            v-if="activePanel === 'instances'"
            id="contest-ops-panel-instances"
            class="tab-panel contest-ops-tab-panel active"
            role="tabpanel"
            aria-labelledby="contest-ops-tab-instances"
          >
            <AWDOperationsPanel
              :key="`${contest.id}-instances`"
              :contests="[contest]"
              :selected-contest-id="contest.id"
              :hide-contest-selector="true"
              :hide-studio-link="true"
              :hide-readiness-actions="true"
              :hide-operation-tabs="true"
              operation-panel="instances"
              runtime-content="instances"
            />
          </section>
        </section>
      </main>
    </div>
  </section>
</template>

<style scoped>
.contest-ops-shell {
  position: relative;
  --workspace-shell-bg: var(--journal-surface);
  --workspace-shell-border: color-mix(in srgb, var(--journal-border) 84%, transparent);
}

.contest-ops-content {
  display: flex;
  flex-direction: column;
  padding: 0;
}

.contest-ops-workspace {
  --workspace-directory-section-padding: var(--space-5) var(--space-5-5);
  --page-top-tabs-gap: var(--space-7);
  --page-top-tabs-margin: 0;
  --page-top-tabs-padding: 0;
  --page-top-tabs-border: color-mix(in srgb, var(--journal-ink) 10%, transparent);
  --page-top-tab-min-height: 52px;
  --page-top-tab-padding: var(--space-2-5) 0 var(--space-3-5);
  --page-top-tab-font-size: var(--font-size-15);
  --page-top-tab-active-color: color-mix(in srgb, var(--journal-accent) 74%, var(--journal-ink));
  --page-top-tab-active-border: color-mix(in srgb, var(--journal-accent) 86%, var(--journal-ink));
  background: transparent;
}

.contest-ops-tab-panel {
  padding-top: var(--space-6);
}

.ops-loading-overlay {
  position: absolute;
  inset: 0;
  z-index: 100;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--journal-surface);
}

@media (max-width: 767px) {
  .contest-ops-workspace {
    --workspace-directory-section-padding: var(--space-4-5) var(--space-4);
  }
}
</style>
