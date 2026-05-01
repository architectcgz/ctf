<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useRoute } from 'vue-router'

import { getContest } from '@/api/admin/contests'
import type { ContestDetailData } from '@/api/contracts'
import AWDOperationsPanel from '@/components/platform/contest/AWDOperationsPanel.vue'
import AWDServiceAlertBanner from '@/components/platform/contest/AWDServiceAlertBanner.vue'
import AppLoading from '@/components/common/AppLoading.vue'
import { useBackofficeBreadcrumbDetail } from '@/composables/useBackofficeBreadcrumbDetail'
import { useToast } from '@/composables/useToast'

const route = useRoute()
const toast = useToast()
const { setBreadcrumbDetailTitle } = useBackofficeBreadcrumbDetail()

const contestId = computed(() => String(route.params.id ?? ''))
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
          >
            <template
              #service-alerts="{
                serviceAlerts,
                selectedAlertKey,
                getServiceAlertClass,
                applyServiceAlertFilter,
              }"
            >
              <AWDServiceAlertBanner
                class="contest-ops-service-alerts"
                :alerts="serviceAlerts"
                :selected-alert-key="selectedAlertKey"
                :get-alert-class="getServiceAlertClass"
                @select-alert="applyServiceAlertFilter"
              />
            </template>
          </AWDOperationsPanel>
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
  --workspace-directory-section-padding: var(--space-4) var(--space-5-5);
  background: transparent;
}

.contest-ops-service-alerts {
  margin-bottom: var(--space-8);
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
