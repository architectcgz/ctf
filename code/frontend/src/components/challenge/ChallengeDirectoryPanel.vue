<script setup lang="ts">
import { Search } from 'lucide-vue-next'

import type { ChallengeCategory, ChallengeDifficulty, ChallengeListItem } from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'
import PagePaginationControls from '@/components/common/PagePaginationControls.vue'
import { ChallengeDirectoryRow } from '@/entities/challenge'

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

defineProps<Props>()

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
    class="student-directory-section workspace-directory-section challenge-directory-section"
    aria-label="题目目录"
  >
    <section class="student-directory-shell challenge-directory-shell workspace-directory-list">
      <header class="student-directory-shell__head challenge-directory-shell__head list-heading">
        <div class="student-directory-shell__heading challenge-directory-shell__heading">
          <div
            class="journal-note-label student-directory-shell__eyebrow challenge-directory-shell__eyebrow"
          >
            Challenge Directory
          </div>
          <h2 class="student-directory-shell__title challenge-directory-shell__title">题目列表</h2>
        </div>
      </header>

      <section class="student-directory-filters challenge-directory-filters" aria-label="题目筛选">
        <div class="student-directory-filter-grid challenge-directory-filter-grid">
          <label
            class="student-directory-filter-field challenge-directory-filter-field"
            for="challenge-category-filter"
          >
            <span class="student-directory-filter-label challenge-directory-filter-label"
              >分类</span
            >
            <div
              class="ui-control-wrap student-directory-filter-control challenge-directory-filter-control"
            >
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
            class="student-directory-filter-field challenge-directory-filter-field"
            for="challenge-difficulty-filter"
          >
            <span class="student-directory-filter-label challenge-directory-filter-label"
              >难度</span
            >
            <div
              class="ui-control-wrap student-directory-filter-control challenge-directory-filter-control"
            >
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

          <div class="student-directory-filter-search challenge-directory-filter-search">
            <label
              class="student-directory-filter-search__label challenge-directory-filter-search__label"
              for="challenge-search-input"
            >
              <span class="student-directory-filter-label challenge-directory-filter-label"
                >搜索</span
              >
              <span
                class="ui-control-wrap student-directory-filter-control challenge-directory-filter-search__control"
              >
                <span class="ui-control-prefix">
                  <Search class="h-4 w-4" />
                </span>
                <input
                  id="challenge-search-input"
                  :value="searchQuery"
                  type="text"
                  placeholder="搜索题目标题或描述..."
                  class="ui-control"
                  @input="updateSearchQuery"
                />
              </span>
            </label>
          </div>

          <div class="student-directory-filter-actions challenge-directory-filter-actions">
            <span
              class="student-directory-filter-label student-directory-filter-label--ghost challenge-directory-filter-label challenge-directory-filter-label--ghost"
              aria-hidden="true"
            >
              操作
            </span>
            <div class="student-directory-filter-action-row challenge-directory-filter-action-row">
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
        class="challenge-loading student-directory-state challenge-directory-state workspace-directory-loading"
      >
        <div class="student-directory-spinner" />
      </div>

      <AppEmpty
        v-else-if="hasLoadError"
        class="challenge-empty-state student-directory-state challenge-directory-state workspace-directory-empty"
        icon="AlertTriangle"
        title="题目列表加载失败"
        :description="errorMessage"
      >
        <template #action>
          <button type="button" class="ui-btn ui-btn--secondary" @click="emit('refresh')">
            重新加载
          </button>
        </template>
      </AppEmpty>

      <AppEmpty
        v-else-if="list.length === 0"
        class="challenge-empty-state student-directory-state challenge-directory-state workspace-directory-empty"
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

      <section v-else class="challenge-directory">
        <div class="challenge-directory-head">
          <span>题目</span>
          <span>积分</span>
          <span>分类</span>
          <span>难度</span>
          <span>标签</span>
          <span>状态</span>
          <span>解出</span>
          <span>尝试</span>
          <span class="challenge-directory-head__action">操作</span>
        </div>

        <ChallengeDirectoryRow
          v-for="challenge in list"
          :key="challenge.id"
          :challenge="challenge"
          @open="emit('open-detail', $event)"
        />

        <div v-if="total > 0" class="challenge-pagination workspace-directory-pagination">
          <PagePaginationControls
            :page="page"
            :total-pages="totalPages"
            :total="total"
            :total-label="`共 ${total} 题`"
            @change-page="emit('change-page', $event)"
          />
        </div>
      </section>
    </section>
  </section>
</template>

<style scoped>
.challenge-directory-shell__eyebrow {
  color: color-mix(in srgb, var(--color-primary) 72%, var(--journal-muted));
}

.challenge-loading {
  display: flex;
  align-items: center;
  justify-content: center;
}

:deep(.challenge-empty-state.challenge-directory-state) {
  margin-top: 0;
  border-color: color-mix(in srgb, var(--journal-border) 80%, transparent);
  background: color-mix(in srgb, var(--journal-surface) 92%, transparent);
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
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 84%, transparent);
  font-size: var(--font-size-11);
  font-weight: 700;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.challenge-directory-head__action {
  text-align: right;
}

.challenge-pagination {
  margin-top: var(--workspace-directory-gap-pagination);
}

@media (max-width: 1180px) {
  .challenge-directory-head {
    display: none;
  }
}
</style>
