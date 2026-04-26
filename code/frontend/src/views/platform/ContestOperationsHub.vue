<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'

import { getContests } from '@/api/admin'
import type { ContestDetailData } from '@/api/contracts'
import ContestOperationsHubHeroPanel from '@/components/platform/contest/ContestOperationsHubHeroPanel.vue'
import ContestOperationsHubWorkspacePanel from '@/components/platform/contest/ContestOperationsHubWorkspacePanel.vue'

const router = useRouter()

const loading = ref(true)
const loadError = ref('')
const contests = ref<ContestDetailData[]>([])

const awdContests = computed(() => contests.value.filter((item) => item.mode === 'awd'))
const operableContests = computed(() =>
  awdContests.value.filter((item) =>
    ['running', 'frozen', 'registering'].includes(item.status)
  )
)
const runningContestCount = computed(
  () => operableContests.value.filter((item) => item.status === 'running').length
)
const frozenContestCount = computed(
  () => operableContests.value.filter((item) => item.status === 'frozen').length
)
const preferredContest = computed(
  () =>
    operableContests.value.find((item) => item.status === 'running') ||
    operableContests.value.find((item) => item.status === 'frozen') ||
    operableContests.value[0] ||
    null
)

async function loadContests(): Promise<void> {
  loading.value = true
  loadError.value = ''

  try {
    const response = await getContests({
      page: 1,
      page_size: 100,
    })
    contests.value = response.list
  } catch (error) {
    contests.value = []
    loadError.value = error instanceof Error ? error.message : '赛事运维目录加载失败'
  } finally {
    loading.value = false
  }
}

async function handleEnterOperations(contestId: string): Promise<void> {
  await router.push({
    name: 'ContestOperations',
    params: { id: contestId },
  })
}

async function handleBackToContestDirectory(): Promise<void> {
  await router.push({
    name: 'ContestManage',
    query: { panel: 'list' },
  })
}

onMounted(() => {
  void loadContests()
})
</script>

<template>
  <section
    class="journal-shell journal-shell-admin journal-notes-card journal-hero workspace-shell flex min-h-full flex-1 flex-col"
  >
    <div class="workspace-grid">
      <main class="content-pane contest-ops-content">
        <ContestOperationsHubHeroPanel
          :operable-contest-count="operableContests.length"
          :running-contest-count="runningContestCount"
          :frozen-contest-count="frozenContestCount"
          :preferred-contest-title="preferredContest ? preferredContest.title : '暂无'"
          @back="void handleBackToContestDirectory()"
        />

        <ContestOperationsHubWorkspacePanel
          :loading="loading"
          :load-error="loadError"
          :operable-contests="operableContests"
          @retry="void loadContests()"
          @back="void handleBackToContestDirectory()"
          @enter-operations="void handleEnterOperations($event)"
        />
      </main>
    </div>
  </section>
</template>

<style scoped>
.contest-ops-content {
  display: flex;
  flex-direction: column;
  gap: var(--workspace-directory-page-block-gap, var(--space-5));
}
</style>
