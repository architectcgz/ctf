<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ArrowRight, Filter, LayoutDashboard, Search, Target } from 'lucide-vue-next'
import { useRouter } from 'vue-router'

import { getChallenges } from '@/api/challenge'
import { ApiError } from '@/api/request'
import type { ChallengeCategory, ChallengeDifficulty } from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'
import { usePagination } from '@/composables/usePagination'

const router = useRouter()
const searchQuery = ref('')
const categoryFilter = ref<ChallengeCategory | ''>('')
const difficultyFilter = ref<ChallengeDifficulty | ''>('')

const { list, total, page, pageSize, loading, error, changePage, refresh } = usePagination(
  (params) => {
    const filters: Record<string, unknown> = { ...params }
    if (searchQuery.value) filters.search = searchQuery.value
    if (categoryFilter.value) filters.category = categoryFilter.value
    if (difficultyFilter.value) filters.difficulty = difficultyFilter.value
    return getChallenges(filters)
  }
)

const hasActiveFilters = computed(() =>
  Boolean(searchQuery.value || categoryFilter.value || difficultyFilter.value)
)
const hasLoadError = computed(() => Boolean(error.value) && list.value.length === 0)
const errorMessage = computed(() => {
  if (error.value instanceof ApiError) return error.value.message
  if (error.value instanceof Error) return error.value.message
  return '挑战列表暂时无法加载，请稍后重试。'
})
const emptyTitle = computed(() => (hasActiveFilters.value ? '没有匹配的题目' : '目前还没有题目'))
const emptyDescription = computed(() =>
  hasActiveFilters.value
    ? '当前筛选条件下没有找到可训练的题目，建议放宽分类、难度或搜索词。'
    : '管理员还没有发布训练题目，稍后再来查看即可。'
)

const totalPages = computed(() => Math.ceil(total.value / pageSize.value))

const solvedCount = computed(() => list.value.filter((c) => c.is_solved).length)
const unsolvedCount = computed(() => list.value.filter((c) => !c.is_solved).length)
const activeFilterCount = computed(
  () => [searchQuery.value, categoryFilter.value, difficultyFilter.value].filter(Boolean).length
)
const summaryStats = computed(() => [
  { key: 'total', label: '题目总数', value: total.value, helper: '当前可训练题库规模' },
  {
    key: 'visible',
    label: '当前结果',
    value: list.value.length,
    helper: hasActiveFilters.value ? '已按筛选条件收束范围' : '当前页展示结果数',
  },
  { key: 'solved', label: '已解出', value: solvedCount.value, helper: '当前列表内已完成题目' },
  { key: 'unsolved', label: '待攻克', value: unsolvedCount.value, helper: '仍可直接进入训练' },
])

function onSearch() {
  page.value = 1
  void refresh()
}

function onFilterChange() {
  page.value = 1
  void refresh()
}

function resetFilters() {
  searchQuery.value = ''
  categoryFilter.value = ''
  difficultyFilter.value = ''
  page.value = 1
  void refresh()
}

function goToDashboard() {
  router.push({ name: 'Dashboard' })
}

function openSkillProfile() {
  router.push({ name: 'SkillProfile' })
}

function goToDetail(id: string) {
  router.push(`/challenges/${id}`)
}

function challengeIndex(index: number): number {
  return (page.value - 1) * pageSize.value + index + 1
}

function getCategoryLabel(category: ChallengeCategory): string {
  const labels: Record<ChallengeCategory, string> = {
    web: 'Web',
    pwn: 'Pwn',
    reverse: '逆向',
    crypto: '密码',
    misc: '杂项',
    forensics: '取证',
  }
  return labels[category]
}

function getCategoryColor(category: ChallengeCategory): string {
  const map: Record<ChallengeCategory, string> = {
    web: 'var(--color-cat-web)',
    pwn: 'var(--color-cat-pwn)',
    reverse: 'var(--color-cat-reverse)',
    crypto: 'var(--color-cat-crypto)',
    misc: 'var(--color-cat-misc)',
    forensics: 'var(--color-cat-forensics)',
  }
  return map[category]
}

function getDifficultyLabel(difficulty: ChallengeDifficulty): string {
  const labels: Record<ChallengeDifficulty, string> = {
    beginner: '入门',
    easy: '简单',
    medium: '中等',
    hard: '困难',
    insane: '地狱',
  }
  return labels[difficulty]
}

function getDifficultyColor(difficulty: ChallengeDifficulty): string {
  const map: Record<ChallengeDifficulty, string> = {
    beginner: 'var(--color-diff-beginner)',
    easy: 'var(--color-diff-easy)',
    medium: 'var(--color-diff-medium)',
    hard: 'var(--color-diff-hard)',
    insane: 'var(--color-diff-insane)',
  }
  return map[difficulty]
}

onMounted(() => {
  void refresh()
})
</script>

<template>
  <div class="journal-shell space-y-6">
    <section class="journal-hero rounded-[30px] border px-6 py-6 md:px-8">
      <div class="grid gap-6 xl:grid-cols-[1.06fr_0.94fr]">
        <div>
          <div class="journal-eyebrow">Training Range</div>
          <h2
            class="mt-3 text-3xl font-semibold tracking-tight text-[var(--journal-ink)] md:text-[2.45rem]"
          >
            靶场训练
          </h2>
          <p class="mt-3 max-w-2xl text-sm leading-7 text-[var(--journal-muted)]">
            在这里搜索、筛选并进入题目。
          </p>

          <div class="mt-6 flex flex-wrap gap-3">
            <button
              type="button"
              class="challenge-btn challenge-btn-primary"
              @click="goToDashboard"
            >
              <LayoutDashboard class="h-4 w-4" />
              返回仪表盘
            </button>
            <button
              type="button"
              class="challenge-btn challenge-btn-ghost"
              @click="openSkillProfile"
            >
              能力画像
            </button>
          </div>
        </div>

        <article class="journal-brief rounded-[24px] border px-5 py-5">
          <div class="flex items-center gap-3 text-sm font-medium text-[var(--journal-ink)]">
            <Target class="h-5 w-5 text-[var(--journal-accent)]" />
            当前题库概况
          </div>
          <div class="mt-5 grid gap-3 sm:grid-cols-2">
            <div v-for="stat in summaryStats" :key="stat.key" class="journal-note">
              <div class="journal-note-label">{{ stat.label }}</div>
              <div class="journal-note-value">{{ stat.value }}</div>
              <div class="journal-note-helper">{{ stat.helper }}</div>
            </div>
          </div>
        </article>
      </div>
      <div class="challenge-filter-panel mt-6">
        <div class="challenge-filter-head gap-4">
          <div>
            <div class="journal-eyebrow">Challenge Filters</div>
            <h3 class="mt-3 text-xl font-semibold text-[var(--journal-ink)]">
              按关键词、分类和难度收束训练范围
            </h3>
            <p class="mt-2 max-w-3xl text-sm leading-7 text-[var(--journal-muted)]">
              用关键词、分类和难度快速筛选题目。
            </p>
          </div>

          <div class="flex flex-wrap items-center gap-3">
            <div class="challenge-filter-pill">
              <Filter class="h-4 w-4" />
              激活筛选 {{ activeFilterCount }} 项
            </div>
            <button
              v-if="hasActiveFilters"
              type="button"
              class="challenge-btn challenge-btn-ghost"
              @click="resetFilters"
            >
              清空筛选
            </button>
          </div>
        </div>

        <div class="mt-5 grid gap-3 lg:grid-cols-[minmax(0,1.35fr)_repeat(2,minmax(0,220px))]">
          <div class="relative">
            <Search
              class="challenge-search-icon absolute left-4 top-1/2 h-4 w-4 -translate-y-1/2 text-[var(--journal-muted)]"
            />
            <input
              v-model="searchQuery"
              type="text"
              placeholder="搜索挑战标题或标签..."
              class="challenge-input"
              @input="onSearch"
            />
          </div>
          <select v-model="categoryFilter" class="challenge-select" @change="onFilterChange">
            <option value="">全部分类</option>
            <option value="web">Web</option>
            <option value="pwn">Pwn</option>
            <option value="reverse">逆向</option>
            <option value="crypto">密码</option>
            <option value="misc">杂项</option>
            <option value="forensics">取证</option>
          </select>
          <select v-model="difficultyFilter" class="challenge-select" @change="onFilterChange">
            <option value="">全部难度</option>
            <option value="beginner">入门</option>
            <option value="easy">简单</option>
            <option value="medium">中等</option>
            <option value="hard">困难</option>
            <option value="insane">地狱</option>
          </select>
        </div>
      </div>

      <div class="challenge-panel-divider" />

      <div v-if="loading" class="flex items-center justify-center py-16">
        <div
          class="h-8 w-8 animate-spin rounded-full border-4 border-[var(--journal-border)] border-t-[var(--journal-accent)]"
        />
      </div>

      <AppEmpty
        v-else-if="hasLoadError"
        class="challenge-empty-state"
        icon="AlertTriangle"
        title="挑战列表加载失败"
        :description="errorMessage"
      >
        <template #action>
          <button type="button" class="challenge-btn" @click="refresh">重新加载</button>
        </template>
      </AppEmpty>

      <AppEmpty
        v-else-if="list.length === 0"
        class="challenge-empty-state"
        icon="Flag"
        :title="emptyTitle"
        :description="emptyDescription"
      >
        <template #action>
          <button v-if="hasActiveFilters" type="button" class="challenge-btn" @click="resetFilters">
            清空筛选
          </button>
        </template>
      </AppEmpty>

      <template v-else>
        <div class="mt-5 grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
          <button
            v-for="(challenge, index) in list"
            :key="challenge.id"
            type="button"
            class="challenge-card rounded-[22px] border px-5 py-5 text-left transition"
            :style="{ borderTopWidth: '3px', borderTopColor: getCategoryColor(challenge.category) }"
            @click="goToDetail(challenge.id)"
          >
            <div class="flex items-center justify-between gap-3">
              <div class="flex items-center gap-2">
                <span
                  class="status-dot"
                  :class="challenge.is_solved ? 'status-dot-solved' : 'status-dot-ready'"
                />
                <span class="challenge-card-code">CH-{{ challengeIndex(index) }}</span>
              </div>
              <span
                class="challenge-card-points"
                :style="{ color: getCategoryColor(challenge.category) }"
              >
                {{ challenge.points }} pts
              </span>
            </div>

            <div class="mt-4">
              <h3 class="challenge-card-title">{{ challenge.title }}</h3>
              <p class="challenge-card-subtitle">
                {{
                  challenge.tags.length > 0
                    ? challenge.tags.join(' / ')
                    : '暂无标签，直接进入题目查看详情。'
                }}
              </p>
            </div>

            <div class="mt-4 flex flex-wrap gap-2">
              <span
                class="challenge-tag"
                :style="{
                  background: getCategoryColor(challenge.category) + '20',
                  color: getCategoryColor(challenge.category),
                }"
              >
                {{ getCategoryLabel(challenge.category) }}
              </span>
              <span
                class="challenge-tag"
                :style="{
                  background: getDifficultyColor(challenge.difficulty) + '20',
                  color: getDifficultyColor(challenge.difficulty),
                }"
              >
                {{ getDifficultyLabel(challenge.difficulty) }}
              </span>
              <span
                class="challenge-state-chip"
                :class="
                  challenge.is_solved ? 'challenge-state-chip-solved' : 'challenge-state-chip-ready'
                "
              >
                {{ challenge.is_solved ? '已解出' : '待攻克' }}
              </span>
            </div>

            <div class="challenge-card-footer mt-5">
              <div class="challenge-card-meta">
                <span>{{ challenge.solved_count }} 人解出</span>
                <span>尝试 {{ challenge.total_attempts }} 次</span>
              </div>
              <span class="challenge-card-cta">
                {{ challenge.is_solved ? '继续查看' : '开始挑战' }}
                <ArrowRight class="h-4 w-4" />
              </span>
            </div>
          </button>
        </div>

        <div v-if="totalPages > 1" class="challenge-pagination">
          <div>
            <div class="journal-note-label">Page Control</div>
            <div class="mt-2 text-sm text-[var(--journal-muted)]">
              共 {{ total }} 题 · 第 {{ page }} / {{ totalPages }} 页
            </div>
          </div>
          <div class="flex items-center gap-2">
            <button
              :disabled="page === 1"
              class="challenge-btn disabled:opacity-40 disabled:cursor-not-allowed"
              @click="changePage(page - 1)"
            >
              上一页
            </button>
            <button
              :disabled="page >= totalPages"
              class="challenge-btn disabled:opacity-40 disabled:cursor-not-allowed"
              @click="changePage(page + 1)"
            >
              下一页
            </button>
          </div>
        </div>
      </template>
    </section>
  </div>
</template>

<style scoped>
.journal-shell {
  --journal-ink: #0f172a;
  --journal-muted: #64748b;
  --journal-accent: #4f46e5;
  --journal-accent-strong: #4338ca;
  --journal-border: rgba(226, 232, 240, 0.8);
  --journal-surface: rgba(248, 250, 252, 0.9);
  --journal-surface-subtle: rgba(241, 245, 249, 0.7);
  font-family: 'Inter', 'Noto Sans SC', system-ui, sans-serif;
}

.journal-hero {
  border-color: var(--journal-border);
  background:
    radial-gradient(circle at top right, rgba(79, 70, 229, 0.08), transparent 18rem),
    linear-gradient(180deg, #ffffff, #f8fafc);
  border-radius: 16px !important;
  overflow: hidden;
  box-shadow: 0 18px 40px rgba(15, 23, 42, 0.06);
}

.journal-brief {
  background: var(--journal-surface-subtle);
  border-color: var(--journal-border);
  border-radius: 16px !important;
  overflow: hidden;
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
  padding: 0.625rem 0.875rem;
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
  font-size: 0.95rem;
  font-weight: 600;
  color: var(--journal-ink);
}

.journal-note-helper {
  margin-top: 0.55rem;
  font-size: 0.78rem;
  line-height: 1.45;
  color: var(--journal-muted);
}

.challenge-filter-head {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-start;
  justify-content: space-between;
}

.challenge-filter-panel {
  border-top: 1px dashed rgba(148, 163, 184, 0.72);
  padding-top: 1.25rem;
}

.challenge-panel-divider {
  margin-top: 1.5rem;
  border-top: 1px dashed rgba(148, 163, 184, 0.62);
}

:deep(.challenge-empty-state) {
  border-top-style: dashed;
  border-bottom-style: dashed;
  border-top-color: rgba(148, 163, 184, 0.58);
  border-bottom-color: rgba(148, 163, 184, 0.58);
}

.challenge-filter-pill {
  display: inline-flex;
  align-items: center;
  gap: 0.45rem;
  border-radius: 999px;
  border: 1px solid rgba(99, 102, 241, 0.16);
  background: rgba(99, 102, 241, 0.06);
  padding: 0.48rem 0.9rem;
  font-size: 0.8rem;
  font-weight: 600;
  color: var(--journal-accent-strong);
}

.challenge-tag {
  border-radius: 6px;
  padding: 0.18rem 0.5rem;
  font-size: 0.72rem;
  font-weight: 600;
}

.challenge-input {
  width: 100%;
  min-height: 2.75rem;
  border-radius: 1rem;
  border: 1px solid var(--journal-border);
  background: var(--journal-surface);
  padding: 0.6rem 1rem 0.6rem 2.85rem;
  font-size: 0.875rem;
  color: var(--journal-ink);
  outline: none;
  transition: border-color 150ms ease;
}

.challenge-search-icon {
  pointer-events: none;
}

.challenge-input::placeholder {
  color: var(--journal-muted);
}

.challenge-input:focus {
  border-color: var(--journal-accent);
  box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.12);
}

.challenge-select {
  width: 100%;
  min-height: 2.75rem;
  border-radius: 1rem;
  border: 1px solid var(--journal-border);
  background: var(--journal-surface);
  padding: 0.6rem 1rem;
  font-size: 0.875rem;
  color: var(--journal-ink);
  cursor: pointer;
  outline: none;
  transition: border-color 150ms ease;
}

.challenge-select:focus {
  border-color: var(--journal-accent);
  box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.12);
}

.challenge-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  min-height: 2.25rem;
  cursor: pointer;
  border-radius: 0.75rem;
  border: 1px solid var(--journal-border);
  background: var(--journal-surface);
  padding: 0.4rem 1rem;
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--journal-ink);
  transition:
    border-color 150ms ease,
    background 150ms ease;
}

.challenge-btn:hover {
  border-color: var(--journal-accent);
  background: rgba(99, 102, 241, 0.06);
}

.challenge-btn-primary {
  border-color: transparent;
  background: var(--journal-accent);
  color: #ffffff;
  box-shadow: 0 12px 24px rgba(79, 70, 229, 0.18);
}

.challenge-btn-primary:hover {
  border-color: transparent;
  background: var(--journal-accent-strong);
}

.challenge-btn-ghost {
  background: rgba(255, 255, 255, 0.66);
}

.challenge-card {
  border-color: var(--journal-border);
  border-radius: 16px !important;
  overflow: hidden;
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.92), rgba(248, 250, 252, 0.82));
  cursor: pointer;
}

.challenge-card:hover {
  border-color: rgba(99, 102, 241, 0.28);
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.98), rgba(248, 250, 252, 0.9));
}

.challenge-card-code {
  font-family: 'JetBrains Mono', 'Fira Code', monospace;
  font-size: 0.76rem;
  font-weight: 600;
  letter-spacing: 0.08em;
  color: var(--journal-muted);
}

.challenge-card-points {
  font-family: 'JetBrains Mono', 'Fira Code', monospace;
  font-size: 0.8rem;
  font-weight: 700;
}

.challenge-card-title {
  font-family: 'JetBrains Mono', 'Fira Code', monospace;
  font-size: 1rem;
  font-weight: 700;
  line-height: 1.5;
  color: var(--journal-ink);
}

.challenge-card-subtitle {
  margin-top: 0.7rem;
  min-height: 2.9rem;
  font-size: 0.83rem;
  line-height: 1.65;
  color: var(--journal-muted);
}

.challenge-state-chip {
  border-radius: 999px;
  padding: 0.32rem 0.75rem;
  font-size: 0.72rem;
  font-weight: 600;
}

.challenge-state-chip-solved {
  background: rgba(16, 185, 129, 0.12);
  color: #059669;
}

.challenge-state-chip-ready {
  background: rgba(79, 70, 229, 0.1);
  color: var(--journal-accent-strong);
}

.challenge-card-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  border-top: 1px dashed rgba(148, 163, 184, 0.56);
  padding-top: 1rem;
}

.challenge-card-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 0.4rem 0.85rem;
  font-size: 0.76rem;
  color: var(--journal-muted);
}

.challenge-card-cta {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  font-size: 0.82rem;
  font-weight: 600;
  color: var(--journal-accent-strong);
}

.challenge-pagination {
  margin-top: 1.5rem;
  border-top: 1px solid var(--journal-border);
  padding-top: 1.5rem;
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
}

.status-dot {
  display: inline-block;
  height: 0.5rem;
  width: 0.5rem;
  border-radius: 999px;
}

.status-dot-solved {
  background: #10b981;
  box-shadow: 0 0 0 4px rgba(16, 185, 129, 0.12);
}

.status-dot-ready {
  background: var(--journal-accent);
  box-shadow: 0 0 0 4px rgba(79, 70, 229, 0.12);
}

:global([data-theme='dark']) .journal-shell {
  --journal-ink: #f1f5f9;
  --journal-muted: #94a3b8;
  --journal-border: rgba(51, 65, 85, 0.72);
  --journal-surface: rgba(15, 23, 42, 0.85);
  --journal-surface-subtle: rgba(30, 41, 59, 0.6);
}

:global([data-theme='dark']) .journal-hero {
  background:
    radial-gradient(circle at top right, rgba(99, 102, 241, 0.14), transparent 18rem),
    linear-gradient(180deg, rgba(15, 23, 42, 0.95), rgba(2, 6, 23, 0.98));
}

:global([data-theme='dark']) .challenge-card {
  background: linear-gradient(180deg, rgba(15, 23, 42, 0.95), rgba(15, 23, 42, 0.82));
}

:global([data-theme='dark']) .challenge-btn-ghost {
  background: rgba(15, 23, 42, 0.72);
}

:global([data-theme='dark']) .challenge-card-footer {
  border-top-color: rgba(51, 65, 85, 0.72);
}
</style>
