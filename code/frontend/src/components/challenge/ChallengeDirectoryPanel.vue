<script setup lang="ts">
import { Search } from 'lucide-vue-next'

import type { ChallengeCategory, ChallengeDifficulty, ChallengeListItem } from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'
import PagePaginationControls from '@/components/common/PagePaginationControls.vue'
import {
  ChallengeDirectoryRow,
} from '@/entities/challenge'

interface Props {
  list: ChallengeListItem[]
  total: number
  page: number
  totalPages: number
  searchQuery: string
  categoryFilter: ChallengeCategory | ''
  difficultyFilter: ChallengeDifficulty | ''
  loading: boolean
  hasActiveFilters: boolean
  hasLoadError: boolean
  errorMessage: string
  emptyTitle: string
  emptyDescription: string
}

const props = defineProps<Props>()

const emit = defineEmits<{
  'update:searchQuery': [value: string]
  'update:categoryFilter': [value: ChallengeCategory | '']
  'update:difficultyFilter': [value: ChallengeDifficulty | '']
  search: []
  'filter-change': []
  'reset-filters': []
  refresh: []
  'open-detail': [challengeId: string]
  'change-page': [page: number]
}>()

function updateSearchQuery(event: Event): void {
  emit('update:searchQuery', (event.target as HTMLInputElement).value)
  emit('search')
}

function updateCategoryFilter(event: Event): void {
  emit('update:categoryFilter', (event.target as HTMLSelectElement).value as ChallengeCategory | '')
  emit('filter-change')
}

function updateDifficultyFilter(event: Event): void {
  emit(
    'update:difficultyFilter',
    (event.target as HTMLSelectElement).value as ChallengeDifficulty | ''
  )
  emit('filter-change')
}

</script>

<template>
  <section
    class="workspace-directory-section challenge-directory-section"
    aria-label="题目目录"
  >
    <header class="list-heading">
      <div>
        <div class="journal-note-label">
          Challenge Directory
        </div>
        <h2 class="list-heading__title">
          题目列表
        </h2>
      </div>
      <div
        id="challenge-directory-meta"
        class="challenge-directory-meta"
      >
        共 {{ total }} 题
        <span v-if="hasActiveFilters">· 已按当前筛选收束结果</span>
      </div>
    </header>

    <section
      class="challenge-directory-filters"
      aria-label="题目筛选"
    >
      <div class="challenge-directory-filter-grid">
        <label
          class="challenge-directory-filter-field"
          for="challenge-category-filter"
        >
          <span class="challenge-directory-filter-label">分类</span>
          <div class="ui-control-wrap">
            <select
              id="challenge-category-filter"
              :value="categoryFilter"
              class="ui-control"
              @change="updateCategoryFilter"
            >
              <option value="">全部分类</option>
              <option value="web">Web</option>
              <option value="pwn">Pwn</option>
              <option value="reverse">逆向</option>
              <option value="crypto">密码</option>
              <option value="misc">杂项</option>
              <option value="forensics">取证</option>
            </select>
          </div>
        </label>

        <label
          class="challenge-directory-filter-field"
          for="challenge-difficulty-filter"
        >
          <span class="challenge-directory-filter-label">难度</span>
          <div class="ui-control-wrap">
            <select
              id="challenge-difficulty-filter"
              :value="difficultyFilter"
              class="ui-control"
              @change="updateDifficultyFilter"
            >
              <option value="">全部难度</option>
              <option value="beginner">入门</option>
              <option value="easy">简单</option>
              <option value="medium">中等</option>
              <option value="hard">困难</option>
              <option value="insane">地狱</option>
            </select>
          </div>
        </label>

        <div class="challenge-directory-filter-search">
          <label
            class="challenge-directory-filter-search__label"
            for="challenge-search-input"
          >
            <span
              class="challenge-directory-filter-label challenge-directory-filter-label--ghost"
              aria-hidden="true"
            >
              搜索
            </span>
            <span class="ui-control-wrap challenge-directory-filter-search__control">
              <span class="ui-control-prefix">
                <Search class="h-4 w-4" />
              </span>
              <input
                id="challenge-search-input"
                :value="searchQuery"
                type="text"
                placeholder="搜索题目标题或描述..."
                class="ui-control"
                aria-describedby="challenge-directory-meta"
                @input="updateSearchQuery"
              >
            </span>
          </label>
        </div>

        <div class="challenge-directory-filter-actions">
          <span
            class="challenge-directory-filter-label challenge-directory-filter-label--ghost"
            aria-hidden="true"
          >
            操作
          </span>
          <div class="challenge-directory-filter-action-row">
            <button
              type="button"
              class="ui-btn ui-btn--ghost challenge-filter-clear"
              :disabled="!hasActiveFilters"
              @click="emit('reset-filters')"
            >
              清空筛选
            </button>
          </div>
        </div>
      </div>
    </section>

    <div
      v-if="loading"
      class="challenge-loading workspace-directory-loading"
    >
      <div class="challenge-loading-spinner" />
    </div>

    <AppEmpty
      v-else-if="hasLoadError"
      class="challenge-empty-state workspace-directory-empty"
      icon="AlertTriangle"
      title="题目列表加载失败"
      :description="errorMessage"
    >
      <template #action>
        <button
          type="button"
          class="ui-btn ui-btn--secondary"
          @click="emit('refresh')"
        >
          重新加载
        </button>
      </template>
    </AppEmpty>

    <AppEmpty
      v-else-if="list.length === 0"
      class="challenge-empty-state workspace-directory-empty"
      icon="Flag"
      :title="emptyTitle"
      :description="emptyDescription"
    >
      <template #action>
        <button
          v-if="hasActiveFilters"
          type="button"
          class="ui-btn ui-btn--secondary"
          @click="emit('reset-filters')"
        >
          清空筛选
        </button>
      </template>
    </AppEmpty>

    <template v-else>
      <section class="challenge-directory">
        <div class="challenge-directory-head">
          <span>题目</span>
          <span>积分</span>
          <span>分类</span>
          <span>难度</span>
          <span>标签</span>
          <span>状态</span>
          <span>解出</span>
          <span>尝试</span>
          <span>操作</span>
        </div>

        <ChallengeDirectoryRow
          v-for="challenge in list"
          :key="challenge.id"
          :challenge="challenge"
          @open="emit('open-detail', $event)"
        />

        <div
          v-if="total > 0"
          class="challenge-pagination workspace-directory-pagination"
        >
          <PagePaginationControls
            :page="page"
            :total-pages="totalPages"
            :total="total"
            :total-label="`共 ${total} 题`"
            @change-page="emit('change-page', $event)"
          />
        </div>
      </section>
    </template>
  </section>
</template>

<style scoped>
.challenge-directory-section {
  margin-top: var(--space-6);
}

.list-heading {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-3);
  align-items: flex-end;
  justify-content: space-between;
}

.list-heading__title {
  margin: var(--space-1) 0 0;
  font-size: var(--font-size-1-20);
  font-weight: 700;
  color: var(--journal-ink);
}

.challenge-directory-meta {
  font-size: var(--font-size-0-82);
  color: var(--journal-muted);
}

.challenge-directory-filters {
  display: grid;
  gap: var(--space-4);
  padding: var(--space-5) 0;
}

.challenge-directory-filter-grid {
  display: grid;
  gap: var(--space-4);
  grid-template-columns: repeat(2, minmax(14rem, 16rem)) minmax(16rem, 1.2fr) auto;
}

.challenge-directory-filter-field,
.challenge-directory-filter-search,
.challenge-directory-filter-actions {
  display: grid;
  gap: var(--space-2);
}

.challenge-directory-filter-search__label {
  display: grid;
  gap: var(--space-2);
}

.challenge-directory-filter-label {
  font-size: var(--font-size-0-78);
  font-weight: 700;
  color: var(--journal-muted);
}

.challenge-directory-filter-label--ghost {
  opacity: 0;
  pointer-events: none;
}

.challenge-directory-filter-actions {
  justify-items: end;
}

.challenge-directory-filter-action-row {
  display: flex;
  justify-content: flex-end;
  gap: var(--space-2-5);
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
    minmax(96px, 0.38fr) minmax(160px, 0.82fr) 120px minmax(104px, 0.42fr) minmax(116px, 0.48fr)
    120px;
  display: flex;
  flex: 1 1 auto;
  flex-direction: column;
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

.challenge-pagination {
  margin-top: var(--space-6);
  padding-top: var(--space-6);
  border-top: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
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
}

@media (max-width: 960px) {
  .list-heading {
    align-items: flex-start;
    flex-direction: column;
  }

  .challenge-directory-filter-grid {
    grid-template-columns: 1fr;
  }

  .challenge-directory-filter-action-row {
    width: 100%;
    justify-content: stretch;
  }

  .challenge-directory-filter-action-row > * {
    flex: 1 1 0;
  }

  .challenge-directory-meta {
    width: 100%;
  }

  .challenge-directory-filter-actions {
    justify-items: stretch;
  }

  .challenge-directory-filter-search__control,
  .challenge-directory-filter-field {
    min-width: 0;
  }

  .challenge-directory-filter-actions,
  .challenge-directory-filter-search,
  .challenge-directory-filter-field {
    width: 100%;
  }

  .challenge-directory-filter-label--ghost {
    display: none;
  }
}
</style>
