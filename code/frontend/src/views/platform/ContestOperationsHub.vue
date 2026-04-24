<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ArrowRight, Settings, Trophy } from 'lucide-vue-next'
import { useRouter } from 'vue-router'

import { getContests } from '@/api/admin'
import type { ContestDetailData } from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'
import AppLoading from '@/components/common/AppLoading.vue'
import ContestOperationsHubHeroPanel from '@/components/platform/contest/ContestOperationsHubHeroPanel.vue'
import { getModeLabel, getStatusLabel } from '@/utils/contest'

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

function formatDateTime(value: string): string {
  return new Date(value).toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

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

async function handleOpenStudio(contestId: string): Promise<void> {
  await router.push({
    name: 'ContestEdit',
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
    <ContestOperationsHubHeroPanel
      :operable-contest-count="operableContests.length"
      :running-contest-count="runningContestCount"
      :frozen-contest-count="frozenContestCount"
      :preferred-contest-title="preferredContest ? preferredContest.title : '暂无'"
      @back="void handleBackToContestDirectory()"
    />

    <section
      v-if="loading"
      class="workspace-directory-section contest-ops-section"
    >
      <AppLoading>正在同步赛事运维目录...</AppLoading>
    </section>

    <AppEmpty
      v-else-if="loadError"
      class="workspace-directory-section contest-ops-section"
      title="赛事运维目录暂时不可用"
      :description="loadError"
      icon="AlertTriangle"
    >
      <template #action>
        <button
          type="button"
          class="ui-btn ui-btn--ghost"
          @click="loadContests"
        >
          重试加载
        </button>
      </template>
    </AppEmpty>

    <AppEmpty
      v-else-if="operableContests.length === 0"
      class="workspace-directory-section contest-ops-section"
      title="当前还没有可进入运维台的 AWD 赛事"
      description="请先在竞赛目录中创建 AWD 赛事，或将赛事推进到可运维状态。"
      icon="Trophy"
    >
      <template #action>
        <button
          type="button"
          class="ui-btn ui-btn--ghost"
          @click="handleBackToContestDirectory"
        >
          返回竞赛目录
        </button>
      </template>
    </AppEmpty>

    <section
      v-else
      class="workspace-directory-section contest-ops-section contest-ops-directory"
    >
      <header class="list-heading">
        <div>
          <div class="journal-note-label">
            Contest Ops Directory
          </div>
          <h2 class="list-heading__title">
            竞赛列表
          </h2>
        </div>
        <div class="contest-section-meta">
          进入具体赛事后查看轮次、流量、大屏和实时榜单
        </div>
      </header>

      <div class="contest-ops-directory__list">
        <article
          v-for="contest in operableContests"
          :key="contest.id"
          class="contest-ops-card"
        >
          <div class="contest-ops-card__head">
            <div class="contest-ops-card__title-wrap">
              <h3 class="contest-ops-card__title">
                {{ contest.title }}
              </h3>
              <p class="contest-ops-card__copy">
                {{ contest.description || '当前未填写赛事描述。' }}
              </p>
            </div>

            <div class="contest-ops-card__badges">
              <span class="contest-ops-card__badge">
                {{ getStatusLabel(contest.status) }}
              </span>
              <span class="contest-ops-card__badge contest-ops-card__badge--muted">
                {{ getModeLabel(contest.mode) }}
              </span>
            </div>
          </div>

          <div class="contest-ops-card__meta">
            <span>开始：{{ formatDateTime(contest.starts_at) }}</span>
            <span>结束：{{ formatDateTime(contest.ends_at) }}</span>
          </div>

          <div class="contest-ops-actions">
            <button
              :id="`contest-ops-enter-${contest.id}`"
              type="button"
              class="ui-btn ui-btn--primary"
              @click="handleEnterOperations(contest.id)"
            >
              <ArrowRight class="h-4 w-4" />
              进入运维台
            </button>
            <button
              type="button"
              class="ui-btn ui-btn--ghost"
              @click="handleOpenStudio(contest.id)"
            >
              <Settings class="h-4 w-4" />
              返回编辑
            </button>
          </div>
        </article>
      </div>
    </section>
  </section>
</template>

<style scoped>
.contest-ops-hero,
.contest-ops-section {
  padding: 1.5rem;
}

.contest-ops-hero {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
}

.contest-ops-hero__main {
  display: grid;
  gap: 0.75rem;
  max-width: 52rem;
}

.contest-ops-hero__actions,
.contest-ops-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
}

.contest-ops-summary {
  margin-top: 1.5rem;
}

.contest-ops-directory {
  display: grid;
  gap: 1rem;
}

.contest-ops-directory__list {
  display: grid;
  gap: 1rem;
}

.contest-ops-card {
  display: grid;
  gap: 1rem;
  padding: 1.2rem 1.25rem;
  border: 1px solid color-mix(in srgb, var(--journal-border) 72%, transparent);
  border-radius: 1.1rem;
  background: color-mix(in srgb, var(--journal-surface) 95%, transparent);
}

.contest-ops-card__head {
  display: flex;
  flex-wrap: wrap;
  justify-content: space-between;
  gap: 0.9rem;
}

.contest-ops-card__title-wrap {
  display: grid;
  gap: 0.45rem;
}

.contest-ops-card__title {
  margin: 0;
  color: var(--journal-ink);
  font-size: 1.05rem;
  font-weight: 700;
}

.contest-ops-card__copy {
  margin: 0;
  color: var(--color-text-secondary);
  line-height: 1.7;
}

.contest-ops-card__badges,
.contest-ops-card__meta {
  display: flex;
  flex-wrap: wrap;
  gap: 0.6rem 0.9rem;
}

.contest-ops-card__badge {
  display: inline-flex;
  align-items: center;
  border-radius: 999px;
  padding: 0.3rem 0.78rem;
  background: color-mix(in srgb, var(--journal-accent) 10%, transparent);
  color: var(--journal-accent);
  font-size: 0.8rem;
  font-weight: 600;
}

.contest-ops-card__badge--muted {
  background: color-mix(in srgb, var(--journal-border) 14%, transparent);
  color: var(--color-text-secondary);
}

.contest-ops-card__meta {
  color: var(--color-text-secondary);
  font-size: 0.9rem;
}

@media (max-width: 768px) {
  .contest-ops-hero,
  .contest-ops-section {
    padding: 1.15rem;
  }

  .contest-ops-card {
    padding: 1rem;
  }

  .contest-ops-card__actions {
    align-items: stretch;
  }
}
</style>
