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
  <section class="workspace-directory-section challenge-directory-section" aria-label="题目目录">
    <section class="challenge-directory-shell workspace-directory-list">
      <header class="challenge-directory-shell__head list-heading">
        <div class="challenge-directory-shell__heading">
          <div class="journal-note-label challenge-directory-shell__eyebrow">
            Challenge Directory
          </div>
          <h2 class="challenge-directory-shell__title">题目列表</h2>
        </div>
      </header>

      <section class="challenge-directory-filters" aria-label="题目筛选">
        <div class="challenge-directory-filter-grid">
          <label class="challenge-directory-filter-field" for="challenge-category-filter">
            <span class="challenge-directory-filter-label">分类</span>
            <div class="ui-control-wrap challenge-directory-filter-control">
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

          <label class="challenge-directory-filter-field" for="challenge-difficulty-filter">
            <span class="challenge-directory-filter-label">难度</span>
            <div class="ui-control-wrap challenge-directory-filter-control">
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
            <label class="challenge-directory-filter-search__label" for="challenge-search-input">
              <span class="challenge-directory-filter-label">搜索</span>
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
                  @input="updateSearchQuery"
                />
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
        class="challenge-loading challenge-directory-state workspace-directory-loading"
      >
        <div class="challenge-loading-spinner" />
      </div>

      <AppEmpty
        v-else-if="hasLoadError"
        class="challenge-empty-state challenge-directory-state workspace-directory-empty"
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
        class="challenge-empty-state challenge-directory-state workspace-directory-empty"
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
.challenge-directory-section {
  margin-top: var(--space-6);
}

.challenge-directory-shell {
  --workspace-directory-shell-padding: var(--space-5);
  --workspace-directory-shell-radius: var(--radius-2xl);
  --workspace-directory-shell-border: color-mix(in srgb, var(--journal-border) 84%, transparent);
  --workspace-directory-shell-background:
    radial-gradient(
      circle at top right,
      color-mix(in srgb, var(--color-primary) 6%, transparent),
      transparent 38%
    ),
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--journal-surface) 98%, var(--color-bg-base)),
      color-mix(in srgb, var(--journal-surface-subtle) 74%, var(--color-bg-base))
    );
  display: grid;
  gap: var(--space-4);
  box-shadow: 0 18px 34px color-mix(in srgb, var(--color-shadow-soft) 20%, transparent);
}

.challenge-directory-shell__head {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: var(--space-4);
  align-items: end;
  padding-bottom: var(--space-4);
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 84%, transparent);
}

.challenge-directory-shell__heading {
  min-width: 0;
}

.challenge-directory-shell__eyebrow {
  color: color-mix(in srgb, var(--color-primary) 72%, var(--journal-muted));
}

.challenge-directory-shell__title {
  margin: var(--space-1) 0 0;
  font-size: var(--workspace-directory-title-size, clamp(1.2rem, 1rem + 0.5vw, 1.45rem));
  font-weight: 700;
  line-height: 1.15;
  color: var(--journal-ink);
}

.challenge-directory-filters {
  display: grid;
  gap: var(--space-4);
}

.challenge-directory-filter-grid {
  display: grid;
  gap: var(--space-4);
  grid-template-columns:
    minmax(12rem, 14rem)
    minmax(12rem, 14rem)
    minmax(16rem, 1fr)
    auto;
  align-items: end;
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

.challenge-directory-filter-control,
.challenge-directory-filter-search__control {
  min-width: 0;
  border-color: color-mix(in srgb, var(--journal-border) 84%, transparent);
  background: color-mix(in srgb, var(--journal-surface) 94%, transparent);
  box-shadow: inset 0 1px 0 color-mix(in srgb, white 30%, transparent);
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
}

.challenge-directory-state {
  min-height: calc(var(--space-12) * 2);
  margin-top: 0;
  border: 0;
  border-radius: var(--radius-xl);
  background: color-mix(in srgb, var(--journal-surface) 86%, transparent);
}

.challenge-loading-spinner {
  width: var(--space-8);
  height: var(--space-8);
  border: var(--space-1) solid color-mix(in srgb, var(--journal-border) 88%, transparent);
  border-top-color: var(--journal-accent);
  border-radius: 999px;
  animation: challengeSpin 900ms linear infinite;
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
  .challenge-directory-shell__head {
    grid-template-columns: minmax(0, 1fr);
    align-items: flex-start;
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
