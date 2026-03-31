<script setup lang="ts">
import { computed } from 'vue'
import { CalendarClock, RefreshCw, Trophy, UserPlus } from 'lucide-vue-next'

import type { ContestDetailData, ContestStatus } from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'
import AppLoading from '@/components/common/AppLoading.vue'
import AdminContestTable from '@/components/admin/contest/AdminContestTable.vue'
import AWDOperationsPanel from '@/components/admin/contest/AWDOperationsPanel.vue'

type StatusFilter = 'all' | Extract<ContestStatus, 'draft' | 'registering' | 'running' | 'frozen' | 'ended'>

const props = defineProps<{
  list: ContestDetailData[]
  total: number
  page: number
  pageSize: number
  loading: boolean
  statusFilter: StatusFilter
  awdContests: ContestDetailData[]
  selectedAwdContestId: string | null
}>()

const emit = defineEmits<{
  refresh: []
  openCreateDialog: []
  updateStatusFilter: [value: StatusFilter]
  openEditDialog: [contest: ContestDetailData]
  changePage: [page: number]
  'update:selectedAwdContestId': [value: string]
}>()

const registeringCount = computed(() => props.list.filter((item) => item.status === 'registering').length)
const runningCount = computed(() => props.list.filter((item) => item.status === 'running').length)
const awdCount = computed(() => props.awdContests.length)
</script>

<template>
  <div class="journal-shell">
    <section class="journal-hero rounded-[30px] border px-6 py-6 md:px-8">
      <div class="grid gap-6 xl:grid-cols-[1.06fr_0.94fr]">
        <div>
          <div class="journal-eyebrow">Contest Orchestration</div>
          <h1 class="mt-3 text-3xl font-semibold tracking-tight text-[var(--journal-ink)] md:text-[2.45rem]">
            赛事编排台
          </h1>
          <p class="mt-3 max-w-2xl text-sm leading-7 text-[var(--journal-muted)]">
            在这里查看赛事窗口、状态流转和 AWD 运维入口。
          </p>

          <div class="mt-6 flex flex-wrap gap-3">
            <button type="button" class="admin-btn admin-btn-ghost" @click="emit('refresh')">
              <RefreshCw class="h-4 w-4" />
              刷新列表
            </button>
            <button type="button" class="admin-btn admin-btn-primary" @click="emit('openCreateDialog')">
              <UserPlus class="h-4 w-4" />
              创建竞赛
            </button>
          </div>
        </div>

        <article class="journal-brief rounded-[24px] border px-5 py-5">
          <div class="flex items-center gap-3 text-sm font-medium text-[var(--journal-ink)]">
            <Trophy class="h-5 w-5 text-[var(--journal-accent)]" />
            当前赛事概况
          </div>
          <div class="mt-5 grid gap-3 sm:grid-cols-2">
            <div class="journal-note">
              <div class="journal-note-label">赛事总量</div>
              <div class="journal-note-value">{{ total }}</div>
              <div class="journal-note-helper">当前筛选条件下的赛事总数</div>
            </div>
            <div class="journal-note">
              <div class="journal-note-label">报名中</div>
              <div class="journal-note-value">{{ registeringCount }}</div>
              <div class="journal-note-helper">当前页开放报名的赛事</div>
            </div>
            <div class="journal-note">
              <div class="journal-note-label">进行中</div>
              <div class="journal-note-value">{{ runningCount }}</div>
              <div class="journal-note-helper">当前页正在进行的赛事</div>
            </div>
            <div class="journal-note">
              <div class="journal-note-label">AWD</div>
              <div class="journal-note-value">{{ awdCount }}</div>
              <div class="journal-note-helper">当前页可直接进入运维视图</div>
            </div>
          </div>
        </article>
      </div>
      <div class="journal-divider mt-6" />

      <div class="admin-section-head admin-section-head-intro">
        <div>
          <div class="journal-note-label">Status Window</div>
          <h2 class="mt-2 text-xl font-semibold text-[var(--journal-ink)]">状态窗口</h2>
        </div>
        <div class="admin-pill">
          <CalendarClock class="h-4 w-4" />
          {{ statusFilter === 'all' ? '全部状态' : statusFilter }}
        </div>
      </div>

      <div class="mt-5 grid gap-4 md:grid-cols-[minmax(0,320px)_1fr]">
        <label class="space-y-2">
          <span class="text-sm text-[var(--journal-muted)]">状态筛选</span>
          <select
            :value="statusFilter"
            class="admin-input"
            @change="emit('updateStatusFilter', ($event.target as HTMLSelectElement).value as StatusFilter)"
          >
            <option value="all">全部状态</option>
            <option value="draft">草稿</option>
            <option value="registering">报名中</option>
            <option value="running">进行中</option>
            <option value="frozen">已冻结</option>
            <option value="ended">已结束</option>
          </select>
        </label>

        <div class="grid gap-3 md:grid-cols-3">
          <div class="journal-note">
            <div class="journal-note-label">列表状态</div>
            <div class="journal-note-helper">创建、编辑和分页都在当前主卡片内完成。</div>
          </div>
          <div class="journal-note">
            <div class="journal-note-label">时间窗口</div>
            <div class="journal-note-helper">筛选后直接判断哪些比赛需要调整。</div>
          </div>
          <div class="journal-note">
            <div class="journal-note-label">AWD 入口</div>
            <div class="journal-note-helper">有 AWD 赛事时会在下方直接展开运维视图。</div>
          </div>
        </div>
      </div>

      <div class="journal-divider mt-6" />

      <section class="space-y-4">
        <div class="admin-section-head">
          <div>
            <div class="journal-note-label">Contests</div>
            <h2 class="mt-2 text-xl font-semibold text-[var(--journal-ink)]">赛事列表</h2>
          </div>
        </div>

        <div v-if="loading && list.length === 0" class="flex justify-center py-10">
          <AppLoading>正在同步竞赛列表...</AppLoading>
        </div>

        <AppEmpty
          v-else-if="list.length === 0"
          title="暂无竞赛"
          description="当前筛选条件下没有竞赛数据。"
          icon="Trophy"
        >
          <template #action>
            <button type="button" class="admin-btn admin-btn-primary" @click="emit('openCreateDialog')">
              创建第一场竞赛
            </button>
          </template>
        </AppEmpty>

        <AdminContestTable
          v-else
          :contests="list"
          :page="page"
          :page-size="pageSize"
          :total="total"
          @edit="emit('openEditDialog', $event)"
          @change-page="emit('changePage', $event)"
        />
      </section>

      <div class="journal-divider mt-6" />

      <section class="space-y-4">
        <div class="admin-section-head">
          <div>
            <div class="journal-note-label">AWD Operations</div>
            <h2 class="mt-2 text-xl font-semibold text-[var(--journal-ink)]">AWD 运维视图</h2>
          </div>
        </div>

        <AWDOperationsPanel
          :contests="awdContests"
          :selected-contest-id="selectedAwdContestId"
          @update:selected-contest-id="emit('update:selectedAwdContestId', $event)"
        />
      </section>
    </section>
  </div>
</template>

<style scoped>
.journal-shell {
  --journal-ink: #0f172a;
  --journal-muted: #64748b;
  --journal-accent: #2563eb;
  --journal-border: rgba(226, 232, 240, 0.84);
  --journal-surface: rgba(248, 250, 252, 0.92);
  --journal-surface-subtle: rgba(241, 245, 249, 0.72);
}

.journal-hero {
  border-color: var(--journal-border);
  background:
    radial-gradient(circle at top right, rgba(37, 99, 235, 0.08), transparent 18rem),
    linear-gradient(180deg, #ffffff, #f8fafc);
  border-radius: 16px !important;
  box-shadow: 0 18px 40px rgba(15, 23, 42, 0.06);
}

.journal-brief {
  background: var(--journal-surface-subtle);
  border-color: var(--journal-border);
  border-radius: 16px !important;
  box-shadow: 0 8px 18px rgba(15, 23, 42, 0.035);
}

.journal-eyebrow {
  font-size: 0.7rem;
  font-weight: 700;
  letter-spacing: 0.2em;
  text-transform: uppercase;
  color: var(--journal-accent);
}

.journal-note {
  border-radius: 14px;
  border: 1px solid var(--journal-border);
  background: var(--journal-surface);
  padding: 0.75rem 0.875rem;
}

.journal-note-label {
  font-size: 0.7rem;
  font-weight: 600;
  letter-spacing: 0.15em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.journal-note-value {
  margin-top: 0.35rem;
  font-size: 1rem;
  font-weight: 600;
  color: var(--journal-ink);
}

.journal-note-helper {
  margin-top: 0.55rem;
  font-size: 0.78rem;
  line-height: 1.5;
  color: var(--journal-muted);
}

.journal-divider {
  border-top: 1px dashed rgba(148, 163, 184, 0.7);
}

.admin-section-head {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
}

.admin-section-head-intro {
  position: relative;
  padding: 1rem 1.1rem 1rem 1.35rem;
  border: 1px dashed rgba(148, 163, 184, 0.42);
  border-radius: 18px;
  background: linear-gradient(90deg, rgba(37, 99, 235, 0.08), rgba(255, 255, 255, 0) 72%);
}

.admin-section-head-intro::before {
  content: '';
  position: absolute;
  left: 0.82rem;
  top: 0.95rem;
  bottom: 0.95rem;
  width: 3px;
  border-radius: 999px;
  background: linear-gradient(180deg, rgba(37, 99, 235, 0.92), rgba(59, 130, 246, 0.2));
}

.admin-section-head-intro .journal-note-label {
  color: var(--journal-accent);
}

.admin-pill {
  display: inline-flex;
  align-items: center;
  gap: 0.45rem;
  border-radius: 999px;
  border: 1px solid rgba(37, 99, 235, 0.16);
  background: rgba(37, 99, 235, 0.06);
  padding: 0.48rem 0.9rem;
  font-size: 0.8rem;
  font-weight: 600;
  color: var(--journal-accent);
}

.admin-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  min-height: 2.75rem;
  border-radius: 1rem;
  padding: 0.65rem 1rem;
  font-size: 0.875rem;
  font-weight: 600;
  transition: all 150ms ease;
}

.admin-btn-primary {
  background: var(--journal-accent);
  color: #fff;
}

.admin-btn-primary:hover {
  background: #1d4ed8;
}

.admin-btn-ghost {
  border: 1px solid var(--journal-border);
  background: rgba(255, 255, 255, 0.75);
  color: var(--journal-ink);
}

.admin-btn-ghost:hover {
  border-color: rgba(37, 99, 235, 0.28);
  color: var(--journal-accent);
}

.admin-input {
  width: 100%;
  min-height: 2.75rem;
  border-radius: 1rem;
  border: 1px solid var(--journal-border);
  background: var(--journal-surface);
  padding: 0.7rem 1rem;
  font-size: 0.875rem;
  color: var(--journal-ink);
  outline: none;
  transition: border-color 150ms ease;
}

.admin-input:focus {
  border-color: rgba(37, 99, 235, 0.42);
}

:global([data-theme='dark']) .journal-shell {
  --journal-ink: #e2e8f0;
  --journal-muted: #94a3b8;
  --journal-accent: #60a5fa;
  --journal-border: rgba(71, 85, 105, 0.78);
  --journal-surface: rgba(15, 23, 42, 0.7);
  --journal-surface-subtle: rgba(15, 23, 42, 0.78);
}

:global([data-theme='dark']) .journal-hero {
  background:
    radial-gradient(circle at top right, rgba(96, 165, 250, 0.1), transparent 18rem),
    linear-gradient(180deg, rgba(15, 23, 42, 0.96), rgba(15, 23, 42, 0.9));
}

:global([data-theme='dark']) .admin-section-head-intro {
  border-color: rgba(96, 165, 250, 0.24);
  background: linear-gradient(90deg, rgba(96, 165, 250, 0.14), rgba(15, 23, 42, 0) 72%);
}

@media (max-width: 767px) {
  .journal-hero {
    padding-left: 1rem;
    padding-right: 1rem;
  }
}
</style>
