<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold">挑战列表</h1>
      <div class="flex gap-3">
        <ElInput v-model="searchQuery" placeholder="搜索挑战..." clearable class="w-64" @input="onSearch" />
        <ElSelect v-model="categoryFilter" placeholder="分类" clearable @change="onFilterChange">
          <ElOption label="Web" value="web" />
          <ElOption label="Pwn" value="pwn" />
          <ElOption label="Reverse" value="reverse" />
          <ElOption label="Crypto" value="crypto" />
          <ElOption label="Misc" value="misc" />
          <ElOption label="Forensics" value="forensics" />
        </ElSelect>
        <ElSelect v-model="difficultyFilter" placeholder="难度" clearable @change="onFilterChange">
          <ElOption label="入门" value="beginner" />
          <ElOption label="简单" value="easy" />
          <ElOption label="中等" value="medium" />
          <ElOption label="困难" value="hard" />
          <ElOption label="地狱" value="hell" />
        </ElSelect>
      </div>
    </div>

    <div v-loading="loading" class="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
      <ElCard v-for="challenge in list" :key="challenge.id" shadow="hover" class="cursor-pointer" @click="goToDetail(challenge.id)">
        <div class="space-y-3">
          <div class="flex items-start justify-between">
            <h3 class="text-lg font-semibold">{{ challenge.title }}</h3>
            <ElTag v-if="challenge.is_solved" type="success" size="small">已完成</ElTag>
          </div>

          <div class="flex flex-wrap gap-2">
            <ElTag :type="getCategoryColor(challenge.category)" size="small">{{ getCategoryLabel(challenge.category) }}</ElTag>
            <ElTag :type="getDifficultyColor(challenge.difficulty)" size="small">{{ getDifficultyLabel(challenge.difficulty) }}</ElTag>
            <ElTag v-for="tag in challenge.tags" :key="tag" size="small">{{ tag }}</ElTag>
          </div>

          <div class="flex items-center justify-between text-sm text-gray-600">
            <span>{{ challenge.points }} 分</span>
            <span>{{ challenge.solved_count }} 人解决</span>
          </div>
        </div>
      </ElCard>
    </div>

    <ElPagination
      v-model:current-page="page"
      v-model:page-size="pageSize"
      :total="total"
      :page-sizes="[12, 24, 48]"
      layout="total, sizes, prev, pager, next"
      @current-change="changePage"
      @size-change="changePageSize"
    />
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
    web: 'primary',
    pwn: 'danger',
    reverse: 'warning',
    crypto: 'success',
    misc: 'info',
    forensics: '',
  }
  return colors[category]
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
    beginner: 'info',
    easy: 'success',
    medium: 'warning',
    hard: 'danger',
    hell: 'danger',
  }
  return colors[difficulty]
}

onMounted(() => {
  refresh()
})
</script>
