<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { BookOpen, Filter, Layers2, Search, Swords } from 'lucide-vue-next'
import { useRouter } from 'vue-router'

import { getChallenges } from '@/api/challenge'
import { ApiError } from '@/api/request'
import type { ChallengeCategory, ChallengeDifficulty } from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import { usePagination } from '@/composables/usePagination'

const router = useRouter()
const searchQuery = ref('')
const categoryFilter = ref<ChallengeCategory | ''>('')
const difficultyFilter = ref<ChallengeDifficulty | ''>('')

const { list, total, page, pageSize, loading, error, changePage, refresh } = usePagination((params) => {
  const filters: Record<string, unknown> = { ...params }
  if (searchQuery.value) filters.search = searchQuery.value
  if (categoryFilter.value) filters.category = categoryFilter.value
  if (difficultyFilter.value) filters.difficulty = difficultyFilter.value
  return getChallenges(filters)
})

const hasActiveFilters = computed(() => Boolean(searchQuery.value || categoryFilter.value || difficultyFilter.value))
const hasLoadError = computed(() => Boolean(error.value) && list.value.length === 0)
const errorMessage = computed(() => {
  if (error.value instanceof ApiError) return error.value.message
  if (error.value instanceof Error) return error.value.message
  return '挑战列表暂时无法加载，请稍后重试。'
})
const emptyTitle = computed(() => (hasActiveFilters.value ? '没有匹配的题目' : '目前还没有题目'))
const emptyDescription = computed(() =>
  hasActiveFilters.value ? '当前筛选条件下没有找到可训练的题目，建议放宽分类、难度或搜索词。' : '管理员还没有发布训练题目，稍后再来查看即可。'
)

const totalPages = computed(() => Math.ceil(total.value / pageSize.value))

const summaryStats = computed(() => [
  { key: 'total', label: '题目总数', value: total.value },
  { key: 'solved', label: '已解出', value: list.value.filter((c) => c.is_solved).length },
  { key: 'unsolved', label: '未解出', value: list.value.filter((c) => !c.is_solved).length },
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

function goToDetail(id: string) {
  router.push(`/challenges/${id}`)
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
    <!-- Hero -->
    <section class="journal-hero rounded-[30px] border px-6 py-6 md:px-8">
      <div class="grid gap-6 xl:grid-cols-[1.1fr_0.9fr]">
        <div>
          <div class="journal-eyebrow">Training Range</div>
          <h2 class="mt-3 text-3xl font-semibold tracking-tight text-[var(--journal-ink)] md:text-[2.45rem]">
            靶场训练
          </h2>
          <p class="mt-3 max-w-2xl text-sm leading-7 text-[var(--journal-muted)]">
            按分类和难度浏览训练题目，逐步突破技能边界。
          </p>
        </div>

        <article class="journal-brief rounded-[24px] border px-5 py-5">
          <div class="flex items-center gap-3 text-sm font-medium text-[var(--journal-ink)]">
            <Swords class="h-5 w-5 text-[var(--journal-accent)]" />
            当前题库概况
          </div>
          <div class="mt-4 grid grid-cols-3 gap-3">
            <div v-for="stat in summaryStats" :key="stat.key" class="journal-note">
              <div class="journal-note-label">{{ stat.label }}</div>
              <div class="journal-note-value">{{ stat.value }}</div>
            </div>
          </div>
        </article>
      </div>
    </section>

    <!-- 筛选栏 -->
    <section class="journal-panel rounded-[24px] border px-5 py-5">
      <div class="flex items-center gap-2 text-sm font-medium text-[var(--journal-ink)] mb-4">
        <Filter class="h-4 w-4 text-[var(--journal-accent)]" />
        筛选条件
      </div>
      <div class="grid gap-3 lg:grid-cols-[minmax(0,1.3fr)_repeat(2,minmax(0,220px))]">
        <div class="relative">
          <Search class="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-[var(--journal-muted)]" />
          <input
            v-model="searchQuery"
            type="text"
            placeholder="搜索挑战标题或标签..."
            class="challenge-input pl-9"
            @input="onSearch"
          />
        </div>
        <select
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
        <select
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
      </div>
    </section>

    <!-- 加载中 -->
    <div v-if="loading" class="flex items-center justify-center py-16">
      <div class="h-8 w-8 animate-spin rounded-full border-4 border-[var(--journal-border)] border-t-[var(--journal-accent)]" />
    </div>

    <!-- 加载失败 -->
    <AppEmpty
      v-else-if="hasLoadError"
      icon="AlertTriangle"
      title="挑战列表加载失败"
      :description="errorMessage"
    >
      <template #action>
        <button type="button" class="challenge-btn" @click="refresh">重新加载</button>
      </template>
    </AppEmpty>

    <!-- 空状态 -->
    <AppEmpty
      v-else-if="list.length === 0"
      icon="Flag"
      :title="emptyTitle"
      :description="emptyDescription"
    >
      <template #action>
        <button v-if="hasActiveFilters" type="button" class="challenge-btn" @click="resetFilters">清空筛选</button>
      </template>
    </AppEmpty>

    <!-- 题目卡片 -->
    <template v-else>
      <section class="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
        <article
          v-for="challenge in list"
          :key="challenge.id"
          class="journal-log rounded-[22px] border px-5 py-5 cursor-pointer transition-colors hover:border-[var(--journal-accent)]"
          :style="{ borderTopWidth: '3px', borderTopColor: getCategoryColor(challenge.category) }"
          @click="goToDetail(challenge.id)"
        >
          <div class="flex items-start justify-between gap-2">
            <h3 class="font-mono text-base font-semibold text-[var(--journal-ink)] leading-snug">{{ challenge.title }}</h3>
            <span class="shrink-0 font-mono text-sm font-semibold" :style="{ color: getCategoryColor(challenge.category) }">{{ challenge.points }}pts</span>
          </div>

          <div class="mt-3 flex flex-wrap gap-2">
            <span
              class="challenge-tag"
              :style="{ background: getCategoryColor(challenge.category) + '20', color: getCategoryColor(challenge.category) }"
            >
              {{ getCategoryLabel(challenge.category) }}
            </span>
            <span
              class="challenge-tag"
              :style="{ background: getDifficultyColor(challenge.difficulty) + '20', color: getDifficultyColor(challenge.difficulty) }"
            >
              {{ getDifficultyLabel(challenge.difficulty) }}
            </span>
          </div>

          <p class="mt-3 text-xs text-[var(--journal-muted)] line-clamp-1">
            {{ challenge.tags.length > 0 ? challenge.tags.join(' / ') : '暂无标签' }}
          </p>

          <div class="mt-4 flex items-center justify-between">
            <span class="text-xs text-[var(--journal-muted)]">{{ challenge.solved_count }} 人解出</span>
            <span
              class="rounded-lg px-3 py-1 text-xs font-semibold transition-colors"
              :class="challenge.is_solved
                ? 'bg-[rgba(16,185,129,0.12)] text-[#10b981]'
                : 'bg-[var(--journal-accent)] text-white'"
            >
              {{ challenge.is_solved ? '已解出' : '开始挑战' }}
            </span>
          </div>
        </article>
      </section>

      <!-- 分页 -->
      <div v-if="totalPages > 1" class="journal-panel rounded-[24px] border px-5 py-4 flex items-center justify-between">
        <span class="text-sm text-[var(--journal-muted)]">共 {{ total }} 题 · 第 {{ page }} / {{ totalPages }} 页</span>
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
  </div>
</template>

<style scoped>
.journal-shell {
  --journal-ink: #0f172a;
  --journal-muted: #64748b;
  --journal-border: rgba(226, 232, 240, 0.8);
  --journal-surface: #ffffff;
  --journal-surface-subtle: rgba(248, 250, 252, 0.9);
  --journal-accent: var(--color-primary);
}

.journal-hero {
  border-color: var(--journal-border);
  background:
    radial-gradient(circle at top right, rgba(191, 219, 254, 0.75), transparent 15rem),
    linear-gradient(180deg, #ffffff, #f8fafc);
  box-shadow: 0 18px 40px rgba(15, 23, 42, 0.06);
}

.journal-brief {
  background: rgba(255, 255, 255, 0.75);
  border-color: rgba(99, 102, 241, 0.14);
  box-shadow: 0 4px 16px rgba(99, 102, 241, 0.06);
}

.journal-panel {
  background: var(--journal-surface-subtle);
  border-color: var(--journal-border);
}

.journal-log {
  background: var(--journal-surface);
  border-color: var(--journal-border);
}

.journal-eyebrow {
  display: inline-flex;
  align-items: center;
  border-radius: 999px;
  border: 1px solid rgba(99, 102, 241, 0.22);
  background: rgba(99, 102, 241, 0.07);
  padding: 0.2rem 0.75rem;
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--journal-accent);
}

.journal-note {
  border-radius: 18px;
  border: 1px solid var(--journal-border);
  background: var(--journal-surface-subtle);
  padding: 0.95rem 1rem;
}

.journal-note-label {
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.26em;
  text-transform: uppercase;
  color: #64748b;
}

.journal-note-value {
  margin-top: 0.65rem;
  font-size: 1.05rem;
  font-weight: 600;
  color: var(--journal-ink);
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
  padding: 0.6rem 1rem;
  font-size: 0.875rem;
  color: var(--journal-ink);
  outline: none;
  transition: border-color 150ms ease;
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
  min-height: 2.25rem;
  cursor: pointer;
  border-radius: 0.75rem;
  border: 1px solid var(--journal-border);
  background: var(--journal-surface);
  padding: 0.4rem 1rem;
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--journal-ink);
  transition: border-color 150ms ease, background 150ms ease;
}

.challenge-btn:hover {
  border-color: var(--journal-accent);
  background: rgba(99, 102, 241, 0.06);
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
</style>
