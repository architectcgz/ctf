<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { getContest } from '@/api/admin'
import type { ContestDetailData } from '@/api/contracts'
import AWDOperationsPanel from '@/components/platform/contest/AWDOperationsPanel.vue'
import ContestOperationsTopbarPanel from '@/components/platform/contest/ContestOperationsTopbarPanel.vue'
import AppLoading from '@/components/common/AppLoading.vue'
import { useToast } from '@/composables/useToast'

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

function exitToList() {
  void router.push({ name: 'ContestManage' })
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
  <div class="contest-ops-shell">
    <div
      v-if="loading"
      class="ops-loading-overlay"
    >
      <AppLoading>正在建立指挥链路...</AppLoading>
    </div>

    <main class="ops-content">
      <ContestOperationsTopbarPanel
        v-if="contest"
        :contest-title="contest.title"
        @back="exitToList"
        @open-studio="goToStudio"
      />

      <div class="ops-canvas custom-scrollbar">
        <AWDOperationsPanel
          v-if="contest"
          :key="`${contest.id}-${activeTab}`"
          :contests="[contest]"
          :selected-contest-id="contest.id"
          :hide-contest-selector="true"
          :initial-tab="activeTab"
        />
      </div>
    </main>
  </div>
</template>

<style scoped>
.contest-ops-shell {
  display: flex;
  flex-direction: column;
  height: 100vh;
  width: 100%;
  background: var(--color-bg-base);
  overflow: hidden;
}

.ops-content { flex: 1; display: flex; flex-direction: column; min-width: 0; }

.ops-canvas { flex: 1; overflow-y: auto; }

.ops-loading-overlay {
  position: absolute; inset: 0; z-index: 100;
  background: var(--color-bg-base); display: flex; align-items: center; justify-content: center;
}
</style>
