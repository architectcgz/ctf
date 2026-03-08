<template>
  <div class="space-y-6">
    <PageHeader eyebrow="Training Range" title="靶场训练" description="按分类和难度浏览训练题目，空状态和错误状态都会明确反馈，不再显示空白区域。">
      <button
        type="button"
        class="rounded-xl border border-[var(--color-border-default)] px-4 py-2 text-sm font-medium text-[var(--color-text-primary)] transition-colors hover:border-[var(--color-primary)]"
        @click="refresh"
      >
        刷新列表
      </button>
    </PageHeader>

    <section class="rounded-[28px] border border-border bg-surface/82 p-5 shadow-[0_20px_50px_var(--color-shadow-soft)]">
      <div class="grid gap-3 lg:grid-cols-[minmax(0,1.3fr)_repeat(2,minmax(0,220px))]">
        <input
          v-model="searchQuery"
          type="text"
          placeholder="搜索挑战标题或标签..."
          class="min-h-11 rounded-2xl border border-[var(--color-border-default)] bg-[var(--color-bg-base)] px-4 py-3 text-[var(--color-text-primary)] placeholder-[var(--color-text-muted)] focus:border-[var(--color-primary)] focus:outline-none"
          @input="onSearch"
        />
        <select
          v-model="categoryFilter"
          class="min-h-11 rounded-2xl border border-[var(--color-border-default)] bg-[var(--color-bg-base)] px-4 py-3 text-[var(--color-text-primary)] focus:border-[var(--color-primary)] focus:outline-none"
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
          class="min-h-11 rounded-2xl border border-[var(--color-border-default)] bg-[var(--color-bg-base)] px-4 py-3 text-[var(--color-text-primary)] focus:border-[var(--color-primary)] focus:outline-none"
          @change="onFilterChange"
        >
          <option value="">全部难度</option>
          <option value="beginner">入门</option>
          <option value="easy">简单</option>
          <option value="medium">中等</option>
          <option value="hard">困难</option>
          <option value="hell">地狱</option>
        </select>
      </div>
    </section>

    <div v-if="loading" class="flex items-center justify-center py-12">
      <div class="h-8 w-8 animate-spin rounded-full border-4 border-[var(--color-border-default)] border-t-[var(--color-primary)]"></div>
    </div>

    <AppEmpty
      v-else-if="hasLoadError"
      icon="AlertTriangle"
      title="挑战列表加载失败"
      :description="errorMessage"
    >
      <template #action>
        <button
          type="button"
          class="rounded-xl bg-[var(--color-primary)] px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-[var(--color-primary)]/90"
          @click="refresh"
        >
          重新加载
        </button>
      </template>
    </AppEmpty>

    <AppEmpty
      v-else-if="list.length === 0"
      icon="Flag"
      :title="emptyTitle"
      :description="emptyDescription"
    >
      <template #action>
        <button
          v-if="hasActiveFilters"
          type="button"
          class="rounded-xl border border-[var(--color-border-default)] px-4 py-2 text-sm font-medium text-[var(--color-text-primary)] transition-colors hover:border-[var(--color-primary)]"
          @click="resetFilters"
        >
          清空筛选
        </button>
      </template>
    </AppEmpty>

    <div v-else class="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
      <div
        v-for="challenge in list"
        :key="challenge.id"
        class="cursor-pointer rounded-lg border bg-[var(--color-bg-surface)] transition-all hover:border-[var(--color-primary)]/50"
        :class="challenge.is_solved ? 'border-green-500/30' : 'border-[var(--color-border-default)]'"
        :style="{ borderTopWidth: '2px', borderTopColor: getCategoryBorderColor(challenge.category) }"
        @click="goToDetail(challenge.id)"
      >
        <div class="space-y-3 p-4">
          <div class="flex items-start justify-between">
            <h3 class="font-mono text-lg font-medium text-[var(--color-text-primary)]">{{ challenge.title }}</h3>
            <span class="font-mono text-sm text-[var(--color-primary)]">{{ challenge.points }}pts</span>
          </div>

          <div class="flex flex-wrap gap-2">
            <span
              class="rounded px-2 py-1 text-xs font-medium"
              :style="{ backgroundColor: getCategoryColor(challenge.category) + '20', color: getCategoryColor(challenge.category) }"
            >
              {{ getCategoryLabel(challenge.category) }}
            </span>
            <span
              class="rounded px-2 py-1 text-xs font-medium"
              :style="{ backgroundColor: getDifficultyColor(challenge.difficulty) + '20', color: getDifficultyColor(challenge.difficulty) }"
            >
              {{ getDifficultyLabel(challenge.difficulty) }}
            </span>
          </div>

          <p class="line-clamp-2 text-sm text-[var(--color-text-secondary)]">
            标签：{{ challenge.tags.length > 0 ? challenge.tags.join(' / ') : '暂无' }}
          </p>

          <div class="flex items-center justify-between">
            <span class="text-xs text-[var(--color-text-muted)]">{{ challenge.solved_count }} 人解出</span>
            <button
              class="rounded-lg px-4 py-1.5 text-sm font-medium transition-colors"
              :class="
                challenge.is_solved
                  ? 'bg-[#21262d] text-[var(--color-text-secondary)] hover:bg-[#30363d]'
                  : 'bg-[var(--color-primary)] text-white hover:bg-[var(--color-primary)]/90'
              "
            >
              {{ challenge.is_solved ? '查看详情' : '开始挑战' }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <div v-if="!loading && total > 0" class="flex items-center justify-between border-t border-[var(--color-border-default)] pt-4">
      <span class="text-sm text-[var(--color-text-secondary)]">共 {{ total }} 个挑战</span>
      <div class="flex items-center gap-2">
        <button
          :disabled="page === 1"
          class="rounded-lg border border-[var(--color-border-default)] px-3 py-1.5 text-sm text-[var(--color-text-primary)] transition-colors hover:border-[var(--color-primary)] disabled:cursor-not-allowed disabled:opacity-50"
          @click="changePage(page - 1)"
        >
          上一页
        </button>
        <span class="text-sm text-[var(--color-text-secondary)]">{{ page }} / {{ Math.ceil(total / pageSize) }}</span>
        <button
          :disabled="page >= Math.ceil(total / pageSize)"
          class="rounded-lg border border-[var(--color-border-default)] px-3 py-1.5 text-sm text-[var(--color-text-primary)] transition-colors hover:border-[var(--color-primary)] disabled:cursor-not-allowed disabled:opacity-50"
          @click="changePage(page + 1)"
        >
          下一页
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { getChallenges } from '@/api/challenge'
import { ApiError } from '@/api/request'
import AppEmpty from '@/components/common/AppEmpty.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import { usePagination } from '@/composables/usePagination'
import type { ChallengeCategory, ChallengeDifficulty } from '@/api/contracts'

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
  const colors: Record<ChallengeCategory, string> = {
    web: '#3b82f6',
    pwn: '#ef4444',
    reverse: '#8b5cf6',
    crypto: '#f59e0b',
    misc: '#10b981',
    forensics: '#06b6d4',
  }
  return colors[category]
}

function getCategoryBorderColor(category: ChallengeCategory): string {
  return getCategoryColor(category)
}

function getDifficultyLabel(difficulty: ChallengeDifficulty): string {
  const labels: Record<ChallengeDifficulty, string> = {
    beginner: '入门',
    easy: '简单',
    medium: '中等',
    hard: '困难',
    hell: '地狱',
  }
  return labels[difficulty]
}

function getDifficultyColor(difficulty: ChallengeDifficulty): string {
  const colors: Record<ChallengeDifficulty, string> = {
    beginner: '#10b981',
    easy: '#3b82f6',
    medium: '#f59e0b',
    hard: '#ef4444',
    hell: '#7c3aed',
  }
  return colors[difficulty]
}

onMounted(() => {
  void refresh()
})
</script>
