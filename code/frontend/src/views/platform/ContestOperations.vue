<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { ChevronLeft, Settings, Zap } from 'lucide-vue-next'
import { useRoute, useRouter } from 'vue-router'

import { getContest } from '@/api/admin'
import type { ContestDetailData } from '@/api/contracts'
import AWDOperationsPanel from '@/components/platform/contest/AWDOperationsPanel.vue'
import AppLoading from '@/components/common/AppLoading.vue'
import { useToast } from '@/composables/useToast'

const route = useRoute()
const router = useRouter()
const toast = useToast()

const contestId = computed(() => String(route.params.id ?? ''))
const activeTab = ref<'matrix' | 'attacks' | 'traffic' | 'scoreboard'>((route.query.activeTab as any) || 'matrix')

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

// Watch for query changes to allow sidebar deep links to work
watch(() => route.query.activeTab, (newTab) => {
  if (newTab) {
    activeTab.value = newTab as any
  }
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
      <header
        v-if="contest"
        class="ops-topbar"
      >
        <div class="topbar-left">
          <button
            class="icon-btn"
            title="退出"
            @click="exitToList"
          >
            <ChevronLeft class="h-5 w-5" />
          </button>
          <div class="divider" />
          <div class="brand">
            <div class="brand-overline">
              Command Center
            </div>
            <h1 class="brand-title">
              {{ contest.title }}
            </h1>
          </div>
        </div>

        <div class="topbar-right">
          <button
            class="studio-link-btn"
            @click="goToStudio"
          >
            <Settings class="h-4 w-4" />
            <span>进入竞赛工作室</span>
          </button>
        </div>
      </header>

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
  background: #fdfdfd;
  overflow: hidden;
}

.ops-content { flex: 1; display: flex; flex-direction: column; min-width: 0; }

.ops-topbar {
  height: 4.5rem;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 2rem;
  background: white;
  border-bottom: 1px solid #e2e8f0;
  z-index: 50;
}

.topbar-left { display: flex; align-items: center; gap: 1.5rem; }
.divider { width: 1px; height: 1.5rem; background: #e2e8f0; }

.brand-overline { font-size: 10px; font-weight: 800; text-transform: uppercase; color: #3b82f6; letter-spacing: 0.15em; }
.brand-title { font-size: 1rem; font-weight: 900; color: #0f172a; margin: 0; }

.studio-link-btn {
  display: inline-flex; align-items: center; gap: 0.65rem; height: 2.5rem; padding: 0 1.25rem;
  border-radius: 0.85rem; background: #f8fafc; border: 1px solid #e2e8f0;
  font-size: 13px; font-weight: 700; color: #475569; cursor: pointer; transition: all 0.2s ease;
}
.studio-link-btn:hover { background: white; border-color: #3b82f6; color: #2563eb; }

.ops-canvas { flex: 1; overflow-y: auto; }

.ops-loading-overlay {
  position: absolute; inset: 0; z-index: 100;
  background: white; display: flex; align-items: center; justify-content: center;
}

.icon-btn {
  width: 2.25rem; height: 2.25rem; display: flex; align-items: center; justify-content: center;
  border-radius: 0.6rem; color: #94a3b8; cursor: pointer; transition: all 0.2s ease;
}
.icon-btn:hover { background: #f1f5f9; color: #0f172a; }
</style>
