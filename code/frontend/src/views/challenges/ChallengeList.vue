<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold text-[#c9d1d9]">靶场训练</h1>
      <div class="flex gap-3">
        <input
          v-model="searchQuery"
          type="text"
          placeholder="搜索挑战..."
          class="w-64 rounded-lg border border-[#30363d] bg-[#0d1117] px-4 py-2 text-[#c9d1d9] placeholder-[#6e7681] focus:border-[#0891b2] focus:outline-none"
          @input="onSearch"
        />
        <select
          v-model="categoryFilter"
          class="rounded-lg border border-[#30363d] bg-[#0d1117] px-4 py-2 text-[#c9d1d9] focus:border-[#0891b2] focus:outline-none"
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
          class="rounded-lg border border-[#30363d] bg-[#0d1117] px-4 py-2 text-[#c9d1d9] focus:border-[#0891b2] focus:outline-none"
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
    </div>

    <div v-if="loading" class="flex items-center justify-center py-12">
      <div class="h-8 w-8 animate-spin rounded-full border-4 border-[#30363d] border-t-[#0891b2]"></div>
    </div>

    <div v-else class="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
      <div
        v-for="challenge in list"
        :key="challenge.id"
        class="cursor-pointer rounded-lg border bg-[#161b22] transition-all hover:border-[#0891b2]/50"
        :class="challenge.is_solved ? 'border-green-500/30' : 'border-[#30363d]'"
        :style="{ borderTopWidth: '2px', borderTopColor: getCategoryBorderColor(challenge.category) }"
        @click="goToDetail(challenge.id)"
      >
        <div class="space-y-3 p-4">
          <div class="flex items-start justify-between">
            <h3 class="font-mono text-lg font-medium text-[#c9d1d9]">{{ challenge.title }}</h3>
            <span class="font-mono text-sm text-[#0891b2]">{{ challenge.points }}pts</span>
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

          <p class="line-clamp-2 text-sm text-[#8b949e]">
            {{ challenge.description || '暂无描述' }}
          </p>

          <div class="flex items-center justify-between">
            <span class="text-xs text-[#6e7681]">{{ challenge.solved_count }} 人解出</span>
            <button
              class="rounded-lg px-4 py-1.5 text-sm font-medium transition-colors"
              :class="
                challenge.is_solved
                  ? 'bg-[#21262d] text-[#8b949e] hover:bg-[#30363d]'
                  : 'bg-[#0891b2] text-white hover:bg-[#0891b2]/90'
              "
            >
              {{ challenge.is_solved ? '查看详情' : '开始挑战' }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <div v-if="!loading && total > 0" class="flex items-center justify-between border-t border-[#30363d] pt-4">
      <span class="text-sm text-[#8b949e]">共 {{ total }} 个挑战</span>
      <div class="flex items-center gap-2">
        <button
          :disabled="page === 1"
          class="rounded-lg border border-[#30363d] px-3 py-1.5 text-sm text-[#c9d1d9] transition-colors hover:border-[#0891b2] disabled:cursor-not-allowed disabled:opacity-50"
          @click="changePage(page - 1)"
        >
          上一页
        </button>
        <span class="text-sm text-[#8b949e]">{{ page }} / {{ Math.ceil(total / pageSize) }}</span>
        <button
          :disabled="page >= Math.ceil(total / pageSize)"
          class="rounded-lg border border-[#30363d] px-3 py-1.5 text-sm text-[#c9d1d9] transition-colors hover:border-[#0891b2] disabled:cursor-not-allowed disabled:opacity-50"
          @click="changePage(page + 1)"
        >
          下一页
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { getChallenges } from '@/api/challenge'
import { usePagination } from '@/composables/usePagination'
import type { ChallengeCategory, ChallengeDifficulty } from '@/api/contracts'

const router = useRouter()
const searchQuery = ref('')
const categoryFilter = ref<ChallengeCategory | ''>('')
const difficultyFilter = ref<ChallengeDifficulty | ''>('')

const { list, total, page, pageSize, loading, changePage, changePageSize, refresh } = usePagination((params) => {
  const filters: Record<string, unknown> = { ...params }
  if (searchQuery.value) filters.search = searchQuery.value
  if (categoryFilter.value) filters.category = categoryFilter.value
  if (difficultyFilter.value) filters.difficulty = difficultyFilter.value
  return getChallenges(filters)
})

function onSearch() {
  page.value = 1
  refresh()
}

function onFilterChange() {
  page.value = 1
  refresh()
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
  refresh()
})
</script>
