<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ArrowRight, Filter, LayoutDashboard, Search, Target } from 'lucide-vue-next'
import { useRouter } from 'vue-router'

import { getChallenges } from '@/api/challenge'
import { ApiError } from '@/api/request'
import type { ChallengeCategory, ChallengeDifficulty } from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'
import PagePaginationControls from '@/components/common/PagePaginationControls.vue'
import { usePagination } from '@/composables/usePagination'

const router = useRouter()
const searchQuery = ref('')
const categoryFilter = ref<ChallengeCategory | ''>('')
const difficultyFilter = ref<ChallengeDifficulty | ''>('')

const { list, total, page, pageSize, loading, error, changePage, refresh } = usePagination(
  (params) => {
    const filters: Record<string, unknown> = { ...params }
    if (searchQuery.value) filters.keyword = searchQuery.value
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
  return '题目列表暂时无法加载，请稍后重试。'
})
const emptyTitle = computed(() => (hasActiveFilters.value ? '没有匹配的题目' : '目前还没有题目'))
const emptyDescription = computed(() =>
  hasActiveFilters.value
    ? '当前筛选条件下没有找到可训练的题目，建议放宽分类、难度或搜索词。'
    : '管理员还没有发布训练题目，稍后再来查看即可。'
)

const totalPages = computed(() => Math.max(1, Math.ceil(total.value / Math.max(pageSize.value, 1))))
const solvedCount = computed(() => list.value.filter((challenge) => challenge.is_solved).length)
const unsolvedCount = computed(() => list.value.filter((challenge) => !challenge.is_solved).length)
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

function onSearch(): void {
  page.value = 1
  void refresh()
}

function onFilterChange(): void {
  page.value = 1
  void refresh()
}

function resetFilters(): void {
  searchQuery.value = ''
  categoryFilter.value = ''
  difficultyFilter.value = ''
  page.value = 1
  void refresh()
}

function goToDashboard(): void {
  void router.push({ name: 'Dashboard' })
}

function openSkillProfile(): void {
  void router.push({ name: 'SkillProfile' })
}

function goToDetail(id: string): void {
  void router.push(`/challenges/${id}`)
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
    web: 'var(--challenge-tone-web)',
    pwn: 'var(--challenge-tone-pwn)',
    reverse: 'var(--challenge-tone-reverse)',
    crypto: 'var(--challenge-tone-crypto)',
    misc: 'var(--challenge-tone-misc)',
    forensics: 'var(--challenge-tone-forensics)',
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
    beginner: 'var(--challenge-diff-beginner)',
    easy: 'var(--challenge-diff-easy)',
    medium: 'var(--challenge-diff-medium)',
    hard: 'var(--challenge-diff-hard)',
    insane: 'var(--challenge-diff-insane)',
  }
  return map[difficulty]
}

onMounted(() => {
  void refresh()
})
</script>

<template>
  <section
    class="journal-shell journal-shell-user journal-eyebrow-text journal-hero flex min-h-full flex-1 flex-col rounded-[30px] border px-6 py-6 md:px-8"
  >
    <div class="challenge-page">
      <header class="challenge-topbar">
        <div class="challenge-heading">
          <div class="journal-eyebrow">Challenges</div>
          <h1 class="challenge-title">靶场训练</h1>
          <p class="challenge-subtitle">按关键词、分类与难度筛选题目，直接进入训练。</p>
        </div>

        <div class="challenge-actions">
          <button type="button" class="challenge-btn challenge-btn-primary" @click="goToDashboard">
            <LayoutDashboard class="h-4 w-4" />
            返回仪表盘
          </button>
          <button type="button" class="challenge-btn challenge-btn-ghost" @click="openSkillProfile">
            能力画像
          </button>
        </div>
      </header>

      <section class="challenge-summary">
        <div class="challenge-summary-title">
          <Target class="h-4 w-4" />
          <span>当前题库概况</span>
        </div>
        <div class="challenge-summary-grid">
          <div v-for="stat in summaryStats" :key="stat.key" class="challenge-summary-item">
            <div class="challenge-summary-label">{{ stat.label }}</div>
            <div class="challenge-summary-value">{{ stat.value }}</div>
            <div class="challenge-summary-helper">{{ stat.helper }}</div>
          </div>
        </div>
      </section>

      <section class="challenge-controls">
        <div class="challenge-controls-bar">
          <div class="challenge-controls-heading">
            <h2 class="challenge-controls-title">筛选条件</h2>
            <p class="challenge-controls-copy">快速收束训练范围，定位当前要攻克的题目。</p>
          </div>
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

        <div class="challenge-filter-grid">
          <label class="challenge-input-wrap" for="challenge-search-input">
            <span class="sr-only">搜索题目</span>
            <Search class="challenge-search-icon h-4 w-4" />
            <input
              id="challenge-search-input"
              v-model="searchQuery"
              type="text"
              placeholder="搜索题目标题或描述..."
              class="challenge-input"
              aria-describedby="challenge-directory-meta"
              @input="onSearch"
            />
          </label>

          <label class="challenge-select-wrap" for="challenge-category-filter">
            <span class="sr-only">按分类筛选</span>
            <select
              id="challenge-category-filter"
              v-model="categoryFilter"
              class="challenge-select"
              @change="onFilterChange"
            >
              <option value="">全部分类</option>
              <option value="web">Web</option>
              <option value="pwn">Pwn</option>
              <option value="reverse">逆向</option>
              <option value="crypto">密码</option>
              <option value="misc">杂项</option>
              <option value="forensics">取证</option>
            </select>
          </label>

          <label class="challenge-select-wrap" for="challenge-difficulty-filter">
            <span class="sr-only">按难度筛选</span>
            <select
              id="challenge-difficulty-filter"
              v-model="difficultyFilter"
              class="challenge-select"
              @change="onFilterChange"
            >
              <option value="">全部难度</option>
              <option value="beginner">入门</option>
              <option value="easy">简单</option>
              <option value="medium">中等</option>
              <option value="hard">困难</option>
              <option value="insane">地狱</option>
            </select>
          </label>
        </div>
      </section>

      <div v-if="loading" class="challenge-loading">
        <div class="challenge-loading-spinner" />
      </div>

      <AppEmpty
        v-else-if="hasLoadError"
        class="challenge-empty-state"
        icon="AlertTriangle"
        title="题目列表加载失败"
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
        <section class="challenge-directory" aria-label="题目目录">
          <div class="challenge-directory-top">
            <h2 class="challenge-directory-title">题目列表</h2>
            <div id="challenge-directory-meta" class="challenge-directory-meta">
              共 {{ total }} 题
              <span v-if="hasActiveFilters">· 已按当前筛选收束结果</span>
            </div>
          </div>

          <div class="challenge-directory-head">
            <span>题目</span>
            <span>积分</span>
            <span>分类</span>
            <span>难度</span>
            <span>标签</span>
            <span>状态</span>
            <span>数据</span>
            <span>操作</span>
          </div>

          <button
            v-for="challenge in list"
            :key="challenge.id"
            type="button"
            class="challenge-row"
            :style="{ '--challenge-row-accent': getCategoryColor(challenge.category) }"
            :aria-label="`${challenge.title}，${getCategoryLabel(challenge.category)}，${getDifficultyLabel(challenge.difficulty)}，${challenge.is_solved ? '已解出' : '待攻克'}`"
            @click="goToDetail(challenge.id)"
          >
            <div class="challenge-row-main">
              <div class="challenge-row-title-group">
                <h2 class="challenge-row-title" :title="challenge.title">{{ challenge.title }}</h2>
              </div>
            </div>

            <div class="challenge-row-points">{{ challenge.points }} pts</div>

            <div class="challenge-row-category">
              <span
                class="challenge-chip"
                :style="{
                  '--challenge-chip-bg': `${getCategoryColor(challenge.category)}18`,
                  '--challenge-chip-color': getCategoryColor(challenge.category),
                }"
              >
                {{ getCategoryLabel(challenge.category) }}
              </span>
            </div>

            <div class="challenge-row-difficulty">
              <span
                class="challenge-chip"
                :style="{
                  '--challenge-chip-bg': `${getDifficultyColor(challenge.difficulty)}18`,
                  '--challenge-chip-color': getDifficultyColor(challenge.difficulty),
                }"
              >
                {{ getDifficultyLabel(challenge.difficulty) }}
              </span>
            </div>

            <div class="challenge-row-tags">
              <span
                v-for="tag in challenge.tags.slice(0, 2)"
                :key="tag"
                class="challenge-chip challenge-chip-muted"
              >
                {{ tag }}
              </span>
            </div>

            <div class="challenge-row-status">
              <span
                class="challenge-state-chip"
                :class="
                  challenge.is_solved ? 'challenge-state-chip-solved' : 'challenge-state-chip-ready'
                "
              >
                {{ challenge.is_solved ? '已解出' : '待攻克' }}
              </span>
            </div>

            <div class="challenge-row-metrics">
              <span>{{ challenge.solved_count }} 人解出</span>
              <span>尝试 {{ challenge.total_attempts }} 次</span>
            </div>

            <div class="challenge-row-cta">
              <span>{{ challenge.is_solved ? '继续查看' : '开始做题' }}</span>
              <ArrowRight class="h-4 w-4" />
            </div>
          </button>

          <div v-if="total > 0" class="challenge-pagination workspace-directory-pagination">
            <PagePaginationControls
              :page="page"
              :total-pages="totalPages"
              :total="total"
              :total-label="`共 ${total} 题`"
              @change-page="changePage"
            />
          </div>
        </section>
      </template>
    </div>
  </section>
</template>

<style scoped>
.journal-shell {
  --journal-shell-surface-subtle: color-mix(in srgb, var(--color-bg-surface) 78%, var(--color-bg-base));
  --journal-shell-accent: color-mix(in srgb, var(--color-primary) 86%, var(--journal-ink));
  --journal-shell-accent-strong: color-mix(in srgb, var(--color-primary) 74%, var(--journal-ink));
  --challenge-tone-web: color-mix(in srgb, var(--color-primary) 82%, var(--journal-ink));
  --challenge-tone-pwn: color-mix(in srgb, var(--color-danger) 72%, var(--journal-ink));
  --challenge-tone-reverse: color-mix(in srgb, var(--color-success) 74%, var(--journal-ink));
  --challenge-tone-crypto: color-mix(in srgb, #0f766e 76%, var(--journal-ink));
  --challenge-tone-misc: color-mix(in srgb, #7c3aed 78%, var(--journal-ink));
  --challenge-tone-forensics: color-mix(in srgb, #ea580c 78%, var(--journal-ink));
  --challenge-diff-beginner: color-mix(in srgb, var(--color-success) 76%, var(--journal-ink));
  --challenge-diff-easy: color-mix(in srgb, #0891b2 78%, var(--journal-ink));
  --challenge-diff-medium: color-mix(in srgb, #2563eb 80%, var(--journal-ink));
  --challenge-diff-hard: color-mix(in srgb, #d97706 80%, var(--journal-ink));
  --challenge-diff-insane: color-mix(in srgb, var(--color-danger) 84%, var(--journal-ink));
}

.challenge-page {
  display: flex;
  min-height: 100%;
  flex: 1 1 auto;
  flex-direction: column;
}

.challenge-heading {
  min-width: 0;
}

.challenge-title {
  margin-top: var(--space-3);
  font-size: clamp(32px, 4vw, 46px);
  line-height: 1.02;
  letter-spacing: -0.04em;
  color: var(--journal-ink);
}

.challenge-subtitle {
  margin-top: var(--space-3);
  max-width: 680px;
  font-size: var(--font-size-14);
  line-height: 1.7;
  color: var(--journal-muted);
}

.challenge-actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--space-3);
}

.challenge-controls {
  padding: var(--space-6) 0 0;
}

.challenge-controls-bar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3) var(--space-4);
}

.challenge-controls-heading {
  min-width: 0;
}

.challenge-controls-title {
  font-size: var(--font-size-17);
  font-weight: 700;
  letter-spacing: -0.02em;
  color: var(--journal-ink);
}

.challenge-controls-copy {
  margin-top: var(--space-1-5);
  font-size: var(--font-size-13);
  line-height: 1.6;
  color: var(--journal-muted);
}

.challenge-filter-pill {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
  min-height: 32px;
  padding: 0 var(--space-2-5);
  border: 1px solid color-mix(in srgb, var(--journal-accent) 22%, transparent);
  border-radius: 8px;
  background: color-mix(in srgb, var(--journal-accent) 8%, transparent);
  font-size: var(--font-size-12);
  font-weight: 600;
  color: var(--journal-accent-strong);
}

.challenge-filter-grid {
  display: grid;
  grid-template-columns: minmax(0, 1.4fr) repeat(2, minmax(0, 220px));
  gap: var(--space-3);
  margin-top: var(--space-4-5);
}

.challenge-input-wrap {
  position: relative;
  display: block;
}

.challenge-select-wrap {
  display: block;
}

.challenge-search-icon {
  position: absolute;
  top: 50%;
  left: var(--space-3-5);
  transform: translateY(-50%);
  color: var(--journal-muted);
  pointer-events: none;
}

.challenge-input,
.challenge-select {
  width: 100%;
  min-height: 44px;
  border: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
  border-radius: 12px;
  background: color-mix(in srgb, var(--journal-surface) 88%, transparent);
  font-size: var(--font-size-14);
  color: var(--journal-ink);
  outline: none;
  transition:
    border-color 160ms ease,
    background 160ms ease,
    box-shadow 160ms ease;
}

.challenge-input {
  padding: 0 var(--space-3-5) 0 calc(var(--space-10) + var(--space-0-5));
}

.challenge-select {
  padding: 0 var(--space-3-5);
  cursor: pointer;
}

.challenge-input::placeholder {
  color: var(--journal-muted);
}

.challenge-input:focus,
.challenge-select:focus {
  border-color: color-mix(in srgb, var(--journal-accent) 54%, transparent);
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--journal-accent) 12%, transparent);
}

.challenge-loading {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: calc(var(--space-10) * 2) 0;
}

.challenge-loading-spinner {
  width: 32px;
  height: 32px;
  border: 4px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
  border-top-color: var(--journal-accent);
  border-radius: 999px;
  animation: challengeSpin 900ms linear infinite;
}

:deep(.challenge-empty-state) {
  margin-top: var(--space-6);
  border-top-style: solid;
  border-bottom-style: solid;
  border-top-color: color-mix(in srgb, var(--journal-border) 88%, transparent);
  border-bottom-color: color-mix(in srgb, var(--journal-border) 88%, transparent);
}

.challenge-directory {
  --challenge-directory-columns: minmax(0, 1.25fr) minmax(88px, 0.32fr) minmax(96px, 0.38fr)
    minmax(96px, 0.38fr) minmax(160px, 0.82fr) 120px 180px 120px;
  display: flex;
  flex: 1 1 auto;
  flex-direction: column;
  margin-top: var(--space-6);
}

.challenge-directory-head {
  display: grid;
  grid-template-columns: var(--challenge-directory-columns);
  gap: var(--space-4);
  padding: 0 0 var(--space-3);
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
  font-size: var(--font-size-11);
  font-weight: 700;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.challenge-row {
  display: grid;
  grid-template-columns: var(--challenge-directory-columns);
  gap: var(--space-4);
  align-items: center;
  width: 100%;
  padding: var(--space-4-5) 0;
  border: 0;
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
  background: transparent;
  text-align: left;
  cursor: pointer;
  transition:
    background 160ms ease,
    border-color 160ms ease;
}

.challenge-row:hover,
.challenge-row:focus-visible {
  background: color-mix(
    in srgb,
    var(--challenge-row-accent, var(--journal-accent)) 5%,
    transparent
  );
  box-shadow: inset 2px 0 0
    color-mix(in srgb, var(--challenge-row-accent, var(--journal-accent)) 64%, transparent);
  outline: none;
}

.challenge-row-main {
  display: grid;
  gap: var(--space-2-5);
  min-width: 0;
}

.challenge-row-title-group {
  display: flex;
  align-items: center;
}

.challenge-row-title {
  min-width: 0;
  font-family: var(--font-family-mono);
  font-size: var(--font-size-16);
  font-weight: 600;
  line-height: 1.35;
  color: var(--journal-ink);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.challenge-row-points {
  display: flex;
  align-items: center;
  font-family: var(--font-family-mono);
  font-size: var(--font-size-13);
  font-weight: 700;
  color: var(--challenge-row-accent, var(--journal-accent));
}

.challenge-row-category,
.challenge-row-difficulty {
  display: flex;
  align-items: center;
  min-width: 0;
}

.challenge-row-tags {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
  min-width: 0;
}

.challenge-chip {
  display: inline-flex;
  align-items: center;
  min-height: 26px;
  padding: 0 var(--space-2-5);
  border-radius: 8px;
  background: var(--challenge-chip-bg, color-mix(in srgb, var(--journal-accent) 10%, transparent));
  color: var(--challenge-chip-color, var(--journal-accent-strong));
  font-size: var(--font-size-12);
  font-weight: 600;
}

.challenge-chip-muted {
  background: color-mix(in srgb, var(--journal-muted) 10%, transparent);
  color: var(--journal-muted);
}

.challenge-row-fallback {
  font-size: var(--font-size-13);
  color: var(--journal-muted);
}

.challenge-row-status {
  display: flex;
  justify-content: flex-start;
}

.challenge-state-chip {
  display: inline-flex;
  align-items: center;
  min-height: 28px;
  padding: 0 var(--space-2-5);
  border-radius: 8px;
  font-size: var(--font-size-12);
  font-weight: 600;
}

.challenge-state-chip-solved {
  background: color-mix(in srgb, var(--color-success) 12%, transparent);
  color: color-mix(in srgb, var(--color-success) 84%, var(--journal-ink));
}

.challenge-state-chip-ready {
  background: color-mix(in srgb, var(--journal-accent) 10%, transparent);
  color: var(--journal-accent-strong);
}

.challenge-row-metrics {
  display: grid;
  gap: var(--space-1);
  font-size: var(--font-size-13);
  line-height: 1.5;
  color: var(--journal-muted);
}

.challenge-row-cta {
  display: inline-flex;
  align-items: center;
  justify-content: flex-start;
  gap: var(--space-1-5);
  font-size: var(--font-size-13);
  font-weight: 700;
  color: var(--journal-accent-strong);
}

.challenge-pagination {
  margin-top: var(--space-6);
  padding-top: var(--space-6);
  border-top: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
}

.challenge-btn-ghost {
  background: color-mix(in srgb, var(--journal-surface) 84%, transparent);
}

@keyframes challengeSpin {
  from {
    transform: rotate(0deg);
  }

  to {
    transform: rotate(360deg);
  }
}

@media (max-width: 1180px) {
  .challenge-directory-head {
    display: none;
  }

  .challenge-row {
    grid-template-columns: 1fr;
    gap: var(--space-3-5);
    padding: var(--space-4-5) 0;
  }

  .challenge-row-status,
  .challenge-row-cta {
    justify-content: flex-start;
  }
}

@media (max-width: 960px) {
  .challenge-topbar,
  .challenge-controls-bar {
    align-items: flex-start;
    flex-direction: column;
  }

  .challenge-filter-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 640px) {
  .challenge-title {
    font-size: var(--font-size-34);
  }

  .challenge-directory-top {
    align-items: flex-start;
  }
}
</style>
