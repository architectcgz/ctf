<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { getContest } from '@/api/admin'
import type { ContestDetailData } from '@/api/contracts'
import AWDOperationsPanel from '@/components/platform/contest/AWDOperationsPanel.vue'
import AppLoading from '@/components/common/AppLoading.vue'
import { useToast } from '@/composables/useToast'
import { Settings } from 'lucide-vue-next'

type ContestOperationsTab = 'matrix' | 'attacks' | 'traffic' | 'scoreboard'

const CONTEST_OPERATIONS_TABS: ContestOperationsTab[] = ['matrix', 'attacks', 'traffic', 'scoreboard']

const route = useRoute()
const router = useRouter()
const toast = useToast()

const contestId = computed(() => String(route.params.id ?? ''))
const activeTab = ref<ContestOperationsTab>(resolveActiveTab(route.query.activeTab))

const loading = ref(true)
const contest = ref<ContestDetailData | null>(null)

async function loadContest() {
  if (!contestId.value) return
  loading.value = true
  try {
    contest.value = await getContest(contestId.value)
  } catch (err) {
    toast.error('加载竞赛信息失败')
  } finally {
    loading.value = false
  }
}

function goToStudio() {
  void router.push({ name: 'ContestEdit', params: { id: contestId.value } })
}

function resolveActiveTab(tab: unknown): ContestOperationsTab {
  if (typeof tab === 'string' && CONTEST_OPERATIONS_TABS.includes(tab as ContestOperationsTab)) {
    return tab as ContestOperationsTab
  }

  return 'matrix'
}

// Watch for query changes to allow sidebar deep links to work
watch(() => route.query.activeTab, (newTab) => {
  activeTab.value = resolveActiveTab(newTab)
})

onMounted(() => {
  void loadContest()
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
            :key="`${contest.id}-${activeTab}`"
            :contests="[contest]"
            :selected-contest-id="contest.id"
            :hide-contest-selector="true"
            :initial-tab="activeTab"
            @open:contest-edit="goToStudio"
          />
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
}

.contest-ops-workspace {
  --workspace-directory-section-padding: var(--space-5) var(--space-5-5);
  background: transparent;
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
