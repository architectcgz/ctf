<script setup lang="ts">
import { computed } from 'vue'

import type {
  AdminChallengePublishRequestData,
  ChallengeCategory,
  ChallengeDifficulty,
  ChallengeStatus,
} from '@/api/contracts'
import ChallengePackageImportEntry from '@/components/admin/challenge/ChallengePackageImportEntry.vue'
import ChallengePackageImportReview from '@/components/admin/challenge/ChallengePackageImportReview.vue'
import { useAdminChallenges } from '@/composables/useAdminChallenges'
import { useChallengePackageImport } from '@/composables/useChallengePackageImport'

const { list, total, page, pageSize, loading, changePage, refresh, publish, remove } =
  useAdminChallenges()

const {
  preview,
  uploading,
  committing,
  selectedFileName,
  hasPreview,
  selectPackage,
  resetPreview,
  commitPreview,
} = useChallengePackageImport({
  onCommitted: async () => {
    await refresh()
  },
})

const publishedCount = computed(
  () => list.value.filter((item) => item.status === 'published').length
)
const draftCount = computed(() => list.value.filter((item) => item.status === 'draft').length)

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
  return {
    web: '#2563eb',
    pwn: '#dc2626',
    reverse: '#7c3aed',
    crypto: '#d97706',
    misc: '#0f766e',
    forensics: '#0891b2',
  }[category]
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
  return {
    beginner: '#16a34a',
    easy: '#2563eb',
    medium: '#d97706',
    hard: '#dc2626',
    insane: '#6d28d9',
  }[difficulty]
}

function getStatusLabel(status: ChallengeStatus): string {
  return { draft: '草稿', published: '已发布', archived: '已归档' }[status]
}

function getStatusColor(status: ChallengeStatus): string {
  return { draft: '#64748b', published: '#059669', archived: '#6b7280' }[status]
}

function getPublishRequestLabel(request: AdminChallengePublishRequestData | null): string {
  if (!request) return '未提交检查'

  return {
    queued: '等待检查',
    running: '检查中',
    succeeded: '检查通过',
    failed: '检查失败',
  }[request.status]
}

function getPublishRequestColor(request: AdminChallengePublishRequestData | null): string {
  if (!request) return '#64748b'

  return {
    queued: '#d97706',
    running: '#2563eb',
    succeeded: '#059669',
    failed: '#dc2626',
  }[request.status]
}

async function handleSelectPackage(file: File) {
  await selectPackage(file)
}

async function handleCommitPreview() {
  await commitPreview()
}
</script>

<template>
  <section
    class="journal-shell journal-hero flex min-h-full flex-1 flex-col rounded-[24px] border px-6 py-6 md:px-8"
  >
    <header class="manage-header">
      <div class="manage-header__intro">
        <div class="journal-eyebrow">Challenge Library</div>
        <h1 class="manage-title">挑战管理</h1>
      </div>

      <div class="manage-summary-grid">
        <article class="journal-note">
          <div class="journal-note-label">题目总量</div>
          <div class="journal-note-value">{{ total }}</div>
          <div class="journal-note-helper">当前题库中可管理的题目</div>
        </article>
        <article class="journal-note">
          <div class="journal-note-label">当前页</div>
          <div class="journal-note-value">{{ list.length }}</div>
          <div class="journal-note-helper">当前分页中的题目数量</div>
        </article>
        <article class="journal-note">
          <div class="journal-note-label">已发布</div>
          <div class="journal-note-value">{{ publishedCount }}</div>
          <div class="journal-note-helper">当前页已开放训练的题目</div>
        </article>
        <article class="journal-note">
          <div class="journal-note-label">草稿</div>
          <div class="journal-note-value">{{ draftCount }}</div>
          <div class="journal-note-helper">导入后仍待完善或发布的题目</div>
        </article>
      </div>
    </header>

    <div class="journal-divider" />

    <ChallengePackageImportEntry
      :uploading="uploading"
      :selected-file-name="selectedFileName"
      @select="handleSelectPackage"
    />

    <ChallengePackageImportReview
      v-if="hasPreview && preview"
      :preview="preview"
      :committing="committing"
      @confirm="handleCommitPreview"
      @reset="resetPreview"
    />

    <div class="journal-divider" />

    <section class="space-y-3">
      <header class="list-heading">
        <div>
          <div class="journal-note-label">Imported Challenges</div>
          <h2 class="list-heading__title">已导入题目</h2>
        </div>
      </header>

      <div v-if="loading" class="flex items-center justify-center py-12">
        <div
          class="h-8 w-8 animate-spin rounded-full border-4 border-[var(--journal-border)] border-t-[var(--journal-accent)]"
        />
      </div>

      <template v-else>
        <div v-if="list.length === 0" class="admin-empty">当前还没有题目，请先导入题目包。</div>

        <div v-else class="space-y-3">
          <div class="manage-directory-head" aria-hidden="true">
            <span>题目</span>
            <span>标签</span>
            <span>发布检查</span>
            <span>操作</span>
          </div>

          <article v-for="row in list" :key="row.id" class="challenge-row">
            <div class="challenge-row__identity">
              <div class="challenge-row__index">CH-{{ String(row.id).slice(0, 6).toUpperCase() }}</div>
              <div class="min-w-0">
                <div class="flex flex-wrap items-center gap-2">
                  <h2 class="text-base font-semibold text-[var(--journal-ink)]">{{ row.title }}</h2>
                  <span
                    class="admin-status-chip"
                    :style="{
                      backgroundColor: getStatusColor(row.status) + '18',
                      color: getStatusColor(row.status),
                    }"
                  >
                    {{ getStatusLabel(row.status) }}
                  </span>
                </div>

                <p
                  v-if="row.latestPublishRequest?.failure_summary"
                  class="challenge-row__failure"
                >
                  {{ row.latestPublishRequest.failure_summary }}
                </p>
              </div>
            </div>

            <div class="challenge-row__meta">
              <span
                class="admin-inline-chip"
                :style="{
                  backgroundColor: getCategoryColor(row.category) + '16',
                  color: getCategoryColor(row.category),
                }"
              >
                {{ getCategoryLabel(row.category) }}
              </span>
              <span
                class="admin-inline-chip"
                :style="{
                  backgroundColor: getDifficultyColor(row.difficulty) + '16',
                  color: getDifficultyColor(row.difficulty),
                }"
              >
                {{ getDifficultyLabel(row.difficulty) }}
              </span>
              <span class="admin-inline-chip admin-inline-chip-neutral">{{ row.points }} pts</span>
            </div>

            <div class="challenge-row__review">
              <div class="journal-note-label">Publish Check</div>
              <div
                class="challenge-row__review-status"
                :style="{ color: getPublishRequestColor(row.latestPublishRequest) }"
              >
                {{ getPublishRequestLabel(row.latestPublishRequest) }}
              </div>
            </div>

            <div class="challenge-row__actions">
              <button
                class="admin-btn admin-btn-ghost admin-btn-compact"
                @click="$router.push(`/admin/challenges/${row.id}`)"
              >
                查看
              </button>
              <button
                class="admin-btn admin-btn-ghost admin-btn-compact"
                @click="$router.push(`/admin/challenges/${row.id}/topology`)"
              >
                编排
              </button>
              <button
                class="admin-btn admin-btn-ghost admin-btn-compact"
                @click="$router.push(`/admin/challenges/${row.id}/writeup`)"
              >
                题解
              </button>
              <button
                v-if="row.status !== 'published'"
                class="admin-btn admin-btn-success admin-btn-compact"
                @click="void publish(row)"
              >
                提交发布检查
              </button>
              <button
                class="admin-btn admin-btn-danger admin-btn-compact"
                @click="void remove(row.id)"
              >
                删除
              </button>
            </div>
          </article>
        </div>

        <div v-if="total > 0" class="admin-pagination mt-4">
          <span>共 {{ total }} 条</span>
          <div class="flex items-center gap-2">
            <button
              :disabled="page === 1"
              class="admin-btn admin-btn-ghost admin-btn-compact disabled:cursor-not-allowed disabled:opacity-50"
              @click="void changePage(page - 1)"
            >
              上一页
            </button>
            <span>{{ page }} / {{ Math.ceil(total / pageSize) }}</span>
            <button
              :disabled="page >= Math.ceil(total / pageSize)"
              class="admin-btn admin-btn-ghost admin-btn-compact disabled:cursor-not-allowed disabled:opacity-50"
              @click="void changePage(page + 1)"
            >
              下一页
            </button>
          </div>
        </div>
      </template>
    </section>
  </section>
</template>

<style scoped>
.journal-shell {
  --journal-ink: var(--color-text-primary);
  --journal-muted: var(--color-text-secondary);
  --journal-accent: var(--color-primary);
  --journal-border: color-mix(in srgb, var(--color-border-default) 84%, transparent);
  --journal-surface: color-mix(in srgb, var(--color-bg-surface) 92%, var(--color-bg-base));
  --journal-surface-subtle: color-mix(in srgb, var(--color-bg-surface) 78%, var(--color-bg-base));
}

.journal-hero {
  border-color: var(--journal-border);
  background:
    radial-gradient(circle at top right, color-mix(in srgb, var(--journal-accent) 7%, transparent), transparent 22rem),
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--journal-surface) 96%, var(--color-bg-base)),
      var(--journal-surface)
    );
  box-shadow: 0 22px 50px var(--color-shadow-soft);
}

.journal-brief {
  border: 1px solid var(--journal-border);
  background: linear-gradient(
    180deg,
    color-mix(in srgb, var(--journal-surface) 96%, var(--color-bg-base)),
    color-mix(in srgb, var(--journal-surface-subtle) 94%, var(--color-bg-base))
  );
  border-radius: 16px;
  box-shadow: 0 8px 18px var(--color-shadow-soft);
}

.journal-eyebrow {
  font-size: 0.7rem;
  font-weight: 700;
  letter-spacing: 0.2em;
  text-transform: uppercase;
  color: var(--journal-accent);
}

.journal-note {
  padding: 0 0 0 1rem;
  border-left: 2px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
  background: transparent;
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
  font-weight: 700;
  color: var(--journal-ink);
}

.journal-note-helper {
  margin-top: 0.55rem;
  font-size: 0.78rem;
  line-height: 1.5;
  color: var(--journal-muted);
}

.journal-divider {
  margin-block: 1rem;
  border-top: 1px dashed color-mix(in srgb, var(--journal-border) 88%, transparent);
}

.manage-header {
  display: grid;
  gap: 1.5rem;
  padding-bottom: 1.5rem;
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
}

.manage-title {
  margin-top: 0.75rem;
  font-size: clamp(32px, 4vw, 46px);
  line-height: 1.02;
  letter-spacing: -0.04em;
  color: var(--journal-ink);
}

.manage-summary-grid {
  display: grid;
  gap: 0.85rem;
  grid-template-columns: repeat(4, minmax(0, 1fr));
}

.list-heading {
  display: flex;
  flex-wrap: wrap;
  gap: 0.8rem;
  align-items: flex-end;
}

.list-heading__title {
  margin: 0.3rem 0 0;
  font-size: 1.2rem;
  font-weight: 700;
  color: var(--journal-ink);
}

.manage-directory-head {
  display: grid;
  grid-template-columns: minmax(0, 1.5fr) minmax(0, 1fr) minmax(8rem, 0.8fr) auto;
  gap: 1rem;
  padding: 0 0 0.8rem;
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.16em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.challenge-row {
  display: grid;
  grid-template-columns: minmax(0, 1.5fr) minmax(0, 1fr) minmax(8rem, 0.8fr) auto;
  align-items: start;
  gap: 1rem;
  padding: 1rem 0;
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
}

.challenge-row__identity {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr);
  align-items: start;
  gap: 0.85rem;
  min-width: 0;
}

.challenge-row__index {
  padding-top: 0.1rem;
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.16em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.challenge-row__meta {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  align-content: start;
}

.challenge-row__review {
  display: grid;
  gap: 0.35rem;
  min-width: 0;
}

.challenge-row__review-status {
  font-size: 0.92rem;
  font-weight: 600;
  line-height: 1.5;
}

.challenge-row__failure {
  margin-top: 0.7rem;
  font-size: 0.82rem;
  line-height: 1.6;
  color: var(--color-danger);
}

.challenge-row__actions {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 0.5rem;
}

.admin-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  min-height: 2.45rem;
  border-radius: 0.75rem;
  padding: 0.55rem 0.95rem;
  font-size: 0.875rem;
  font-weight: 600;
  transition:
    border-color 150ms ease,
    background 150ms ease,
    color 150ms ease;
}

.admin-btn-compact {
  min-height: 2.25rem;
  padding: 0.48rem 0.82rem;
}

.admin-btn-ghost {
  border: 1px solid var(--journal-border);
  background: var(--journal-surface);
  color: var(--journal-ink);
}

.admin-btn-success {
  border: 1px solid color-mix(in srgb, var(--color-success) 28%, transparent);
  background: color-mix(in srgb, var(--color-success) 10%, var(--journal-surface));
  color: color-mix(in srgb, var(--color-success) 88%, var(--journal-ink));
}

.admin-btn-danger {
  border: 1px solid color-mix(in srgb, var(--color-danger) 28%, transparent);
  background: color-mix(in srgb, var(--color-danger) 10%, var(--journal-surface));
  color: color-mix(in srgb, var(--color-danger) 88%, var(--journal-ink));
}

.admin-status-chip,
.admin-inline-chip {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  border-radius: 0.5rem;
  padding: 0.32rem 0.65rem;
  font-size: 0.72rem;
  font-weight: 600;
}

.admin-inline-chip-neutral {
  background: color-mix(in srgb, var(--journal-surface-subtle) 90%, var(--color-bg-base));
  color: var(--journal-muted);
}

.admin-empty {
  padding: 1rem 0 0;
  font-size: 0.875rem;
  color: var(--journal-muted);
}

.admin-pagination {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  padding-top: 1rem;
  border-top: 1px dashed color-mix(in srgb, var(--journal-border) 88%, transparent);
  font-size: 0.82rem;
  color: var(--journal-muted);
}

:global([data-theme='dark']) .journal-shell {
  --journal-ink: var(--color-text-primary);
  --journal-muted: var(--color-text-secondary);
  --journal-accent: var(--color-primary-hover);
  --journal-border: color-mix(in srgb, var(--color-border-default) 84%, transparent);
  --journal-surface: color-mix(in srgb, var(--color-bg-surface) 92%, var(--color-bg-base));
  --journal-surface-subtle: color-mix(in srgb, var(--color-bg-surface) 78%, var(--color-bg-base));
}

:global([data-theme='dark']) .journal-hero {
  background:
    radial-gradient(circle at top right, rgba(96, 165, 250, 0.1), transparent 18rem),
    linear-gradient(180deg, rgba(15, 23, 42, 0.96), rgba(15, 23, 42, 0.9));
}

@media (max-width: 960px) {
  .manage-summary-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .manage-directory-head {
    display: none;
  }

  .challenge-row {
    grid-template-columns: minmax(0, 1fr);
  }

  .challenge-row__actions {
    justify-content: flex-start;
  }

  .journal-shell {
    padding-inline: 1.1rem;
  }
}

@media (max-width: 640px) {
  .manage-summary-grid {
    grid-template-columns: minmax(0, 1fr);
  }
}
</style>
