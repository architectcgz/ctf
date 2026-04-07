<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import type {
  AdminChallengeImportPreview,
  AdminChallengePublishRequestData,
  ChallengeCategory,
  ChallengeDifficulty,
  ChallengeStatus,
} from '@/api/contracts'
import ChallengePackageImportEntry from '@/components/admin/challenge/ChallengePackageImportEntry.vue'
import ChallengePackageImportReview from '@/components/admin/challenge/ChallengePackageImportReview.vue'
import { useAdminChallenges } from '@/composables/useAdminChallenges'
import { useChallengePackageImport } from '@/composables/useChallengePackageImport'

type ChallengePanelKey = 'manage' | 'import' | 'queue'

const validPanelKeys = new Set<ChallengePanelKey>(['manage', 'import', 'queue'])
const panelTabs: Array<{ key: ChallengePanelKey; label: string; panelId: string; tabId: string }> = [
  { key: 'manage', label: '题目管理', panelId: 'challenge-panel-manage', tabId: 'challenge-tab-manage' },
  { key: 'import', label: '导入题目包', panelId: 'challenge-panel-import', tabId: 'challenge-tab-import' },
  { key: 'queue', label: '待确认导入', panelId: 'challenge-panel-queue', tabId: 'challenge-tab-queue' },
]

const route = useRoute()
const router = useRouter()

const { list, total, page, pageSize, loading, changePage, refresh, publish, remove } =
  useAdminChallenges()

const {
  preview,
  uploading,
  committing,
  queueLoading,
  selectedFileName,
  queue,
  hasPreview,
  refreshQueue,
  selectPackage,
  loadPreview,
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

const activePanel = computed<ChallengePanelKey>(() => {
  const panel = route.query.panel
  if (typeof panel === 'string' && validPanelKeys.has(panel as ChallengePanelKey)) {
    return panel as ChallengePanelKey
  }
  return 'manage'
})

const queueCount = computed(() => queue.value.length)
const openActionMenuId = ref<string | null>(null)

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

function formatDateTime(value: string): string {
  return new Date(value).toLocaleString('zh-CN')
}

async function handleSelectPackage(file: File) {
  await selectPackage(file)
}

async function handleCommitPreview() {
  await commitPreview()
}

async function openPackageFormatGuide(): Promise<void> {
  await router.push({ name: 'AdminChallengePackageFormat' })
}

async function switchPanel(panelKey: ChallengePanelKey): Promise<void> {
  const nextQuery =
    panelKey === 'manage'
      ? (({ panel: _panel, ...restQuery }) => restQuery)(route.query)
      : { ...route.query, panel: panelKey }
  await router.replace({ name: 'ChallengeManage', query: nextQuery })
}

function focusTabByIndex(index: number): void {
  const safeIndex = Math.max(0, Math.min(index, panelTabs.length - 1))
  const targetTab = panelTabs[safeIndex]
  if (!targetTab) return
  document.getElementById(targetTab.tabId)?.focus()
}

function handleTabKeydown(event: KeyboardEvent, index: number): void {
  if (event.key !== 'ArrowRight' && event.key !== 'ArrowLeft' && event.key !== 'Home' && event.key !== 'End') {
    return
  }

  event.preventDefault()

  if (event.key === 'Home') {
    void switchPanel(panelTabs[0].key)
    focusTabByIndex(0)
    return
  }

  if (event.key === 'End') {
    const endIndex = panelTabs.length - 1
    void switchPanel(panelTabs[endIndex].key)
    focusTabByIndex(endIndex)
    return
  }

  const direction = event.key === 'ArrowRight' ? 1 : -1
  const nextIndex = (index + direction + panelTabs.length) % panelTabs.length
  void switchPanel(panelTabs[nextIndex].key)
  focusTabByIndex(nextIndex)
}

async function inspectImportTask(item: AdminChallengeImportPreview) {
  await loadPreview(item.id)
  await switchPanel('import')
}

function toggleActionMenu(challengeId: string): void {
  openActionMenuId.value = openActionMenuId.value === challengeId ? null : challengeId
}

function closeActionMenu(): void {
  openActionMenuId.value = null
}

function openChallengeDetail(challengeId: string): void {
  closeActionMenu()
  void router.push(`/platform/challenges/${challengeId}`)
}

function openChallengeTopology(challengeId: string): void {
  closeActionMenu()
  void router.push(`/platform/challenges/${challengeId}/topology`)
}

function openChallengeWriteup(challengeId: string): void {
  closeActionMenu()
  void router.push(`/platform/challenges/${challengeId}/writeup`)
}

async function submitPublishCheck(row: (typeof list.value)[number]): Promise<void> {
  closeActionMenu()
  await publish(row)
}

async function removeChallenge(challengeId: string): Promise<void> {
  closeActionMenu()
  await remove(challengeId)
}

onMounted(() => {
  void refreshQueue()
})
</script>

<template>
  <section
    class="journal-shell journal-hero flex min-h-full flex-1 flex-col rounded-[24px] border px-6 py-6 md:px-8"
  >
    <header class="workspace-topbar">
      <div class="topbar-leading">
        <span class="workspace-overline">Challenge Workspace</span>
        <span class="class-chip">题库管理</span>
      </div>
      <div class="top-note">
        <span>当前题目: {{ total }}</span>
        <span>待确认导入: {{ queueCount }}</span>
      </div>
    </header>

    <nav class="top-tabs" role="tablist" aria-label="靶场管理视图切换">
      <button
        v-for="(tab, index) in panelTabs"
        :id="tab.tabId"
        :key="tab.tabId"
        type="button"
        role="tab"
        class="top-tab"
        :class="{ active: activePanel === tab.key }"
        :aria-selected="activePanel === tab.key ? 'true' : 'false'"
        :aria-controls="tab.panelId"
        :tabindex="activePanel === tab.key ? 0 : -1"
        @click="switchPanel(tab.key)"
        @keydown="handleTabKeydown($event, index)"
      >
        {{ tab.label }}
      </button>
    </nav>

    <main class="content-pane">
      <section
        id="challenge-panel-manage"
        class="tab-panel"
        role="tabpanel"
        aria-labelledby="challenge-tab-manage"
        :aria-hidden="activePanel === 'manage' ? 'false' : 'true'"
        v-show="activePanel === 'manage'"
      >
        <header class="manage-header">
          <div class="manage-header__intro">
            <div class="journal-eyebrow">Challenge Library</div>
            <h1 class="manage-title">靶场管理</h1>
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

        <section class="workspace-directory-section">
          <header class="list-heading">
            <div>
              <div class="journal-note-label">Imported Challenges</div>
              <h2 class="list-heading__title">已导入题目</h2>
            </div>
          </header>

          <div v-if="loading" class="workspace-directory-loading flex items-center justify-center py-12">
            <div
              class="h-8 w-8 animate-spin rounded-full border-4 border-[var(--journal-border)] border-t-[var(--journal-accent)]"
            />
          </div>

          <template v-else>
            <div v-if="list.length === 0" class="admin-empty workspace-directory-empty">当前还没有题目，请先导入题目包。</div>

            <div v-else class="challenge-list workspace-directory-list">
              <div class="manage-directory-head" aria-hidden="true">
                <span>题目</span>
                <span>分类</span>
                <span>难度</span>
                <span>分值</span>
                <span>发布状态</span>
                <span>发布检查</span>
                <span class="manage-directory-head__actions">操作</span>
              </div>

              <article v-for="row in list" :key="row.id" class="challenge-row">
                <div class="challenge-row__identity">
                  <div class="min-w-0">
                    <div class="flex flex-wrap items-center gap-2">
                      <h2
                        class="challenge-row__title text-base font-semibold text-[var(--journal-ink)]"
                        :title="row.title"
                      >
                        {{ row.title }}
                      </h2>
                    </div>

                    <p
                      v-if="row.latestPublishRequest?.failure_summary"
                      class="challenge-row__failure"
                      :title="row.latestPublishRequest.failure_summary"
                    >
                      {{ row.latestPublishRequest.failure_summary }}
                    </p>
                  </div>
                </div>

                <div class="challenge-row__category">
                  <span
                    class="admin-inline-chip"
                    :style="{
                      backgroundColor: getCategoryColor(row.category) + '16',
                      color: getCategoryColor(row.category),
                    }"
                  >
                    {{ getCategoryLabel(row.category) }}
                  </span>
                </div>

                <div class="challenge-row__difficulty">
                  <span
                    class="admin-inline-chip"
                    :style="{
                      backgroundColor: getDifficultyColor(row.difficulty) + '16',
                      color: getDifficultyColor(row.difficulty),
                    }"
                  >
                    {{ getDifficultyLabel(row.difficulty) }}
                  </span>
                </div>

                <div class="challenge-row__points">
                  <span class="admin-inline-chip admin-inline-chip-neutral">{{ row.points }} pts</span>
                </div>

                <div class="challenge-row__status">
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

                <div class="challenge-row__review">
                  <div
                    class="challenge-row__review-status"
                    :style="{ color: getPublishRequestColor(row.latestPublishRequest) }"
                  >
                    {{ getPublishRequestLabel(row.latestPublishRequest) }}
                  </div>
                </div>

                <div class="challenge-row__actions" role="group" aria-label="题目操作">
                  <button
                    class="admin-btn admin-btn-primary admin-btn-compact"
                    @click="openChallengeDetail(row.id)"
                  >
                    查看
                  </button>
                  <button
                    class="admin-btn admin-btn-ghost admin-btn-compact"
                    data-testid="challenge-more-actions"
                    :aria-expanded="openActionMenuId === row.id ? 'true' : 'false'"
                    @click="toggleActionMenu(row.id)"
                  >
                    更多
                  </button>

                  <div
                    v-if="openActionMenuId === row.id"
                    class="challenge-row__actions-menu"
                    role="menu"
                    aria-label="更多题目操作"
                  >
                    <button
                      class="admin-btn admin-btn-ghost admin-btn-compact challenge-row__menu-button"
                      role="menuitem"
                      @click="openChallengeTopology(row.id)"
                    >
                      编排
                    </button>
                    <button
                      class="admin-btn admin-btn-ghost admin-btn-compact challenge-row__menu-button"
                      role="menuitem"
                      @click="openChallengeWriteup(row.id)"
                    >
                      题解
                    </button>
                    <button
                      v-if="row.status !== 'published'"
                      class="admin-btn admin-btn-success admin-btn-compact challenge-row__menu-button"
                      role="menuitem"
                      @click="void submitPublishCheck(row)"
                    >
                      提交发布检查
                    </button>
                    <button
                      class="admin-btn admin-btn-danger admin-btn-compact challenge-row__menu-button"
                      role="menuitem"
                      @click="void removeChallenge(row.id)"
                    >
                      删除
                    </button>
                  </div>
                </div>
              </article>
            </div>

            <div v-if="total > 0" class="admin-pagination workspace-directory-pagination">
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

      <section
        id="challenge-panel-import"
        class="tab-panel space-y-4"
        role="tabpanel"
        aria-labelledby="challenge-tab-import"
        :aria-hidden="activePanel === 'import' ? 'false' : 'true'"
        v-show="activePanel === 'import'"
      >
        <ChallengePackageImportEntry
          :uploading="uploading"
          :selected-file-name="selectedFileName"
          @select="handleSelectPackage"
        />

        <section class="sample-guide">
          <div class="sample-guide__header">
            <div>
              <div class="sample-guide__eyebrow">Uploader Guide</div>
              <h2 class="sample-guide__title">题目包示例</h2>
            </div>
            <p class="sample-guide__copy">
              导入页只保留上传和预览流程，目录结构与 `challenge.yml` 示例统一放到独立说明页，避免同一份规则重复维护。
            </p>
          </div>

          <div class="sample-guide__actions">
            <button
              type="button"
              class="sample-guide__link"
              data-testid="challenge-package-format-link"
              @click="void openPackageFormatGuide()"
            >
              查看题目包示例
            </button>
          </div>
        </section>

        <ChallengePackageImportReview
          v-if="hasPreview && preview"
          :preview="preview"
          :committing="committing"
          @confirm="handleCommitPreview"
          @reset="resetPreview"
        />
      </section>

      <section
        id="challenge-panel-queue"
        class="tab-panel space-y-3"
        role="tabpanel"
        aria-labelledby="challenge-tab-queue"
        :aria-hidden="activePanel === 'queue' ? 'false' : 'true'"
        v-show="activePanel === 'queue'"
      >
        <div class="admin-section-head">
          <div>
            <div class="journal-note-label">Import Review</div>
            <h1 class="manage-title manage-title--compact">待确认导入</h1>
          </div>
        </div>

        <p class="panel-copy">
          这里列出已生成预览、但还没正式导入题库的题目包。确认无误后，可继续查看预览并完成导入。
        </p>

        <div v-if="queueLoading" class="flex items-center justify-center py-12">
          <div
            class="h-8 w-8 animate-spin rounded-full border-4 border-[var(--journal-border)] border-t-[var(--journal-accent)]"
          />
        </div>

        <div v-else-if="queue.length === 0" class="admin-empty">
          当前没有待确认的导入任务。
        </div>

        <div v-else class="queue-list">
          <article v-for="item in queue" :key="item.id" class="queue-row">
            <div class="queue-row__identity">
              <div class="challenge-row__index">IMP-{{ item.id.slice(0, 6).toUpperCase() }}</div>
              <div class="min-w-0">
                <h2 class="queue-row__title" :title="item.title">{{ item.title }}</h2>
                <p class="queue-row__meta-text" :title="item.file_name">{{ item.file_name }}</p>
              </div>
            </div>

            <div class="queue-row__summary">
              <span
                class="admin-inline-chip"
                :style="{
                  backgroundColor: getCategoryColor(item.category) + '16',
                  color: getCategoryColor(item.category),
                }"
              >
                {{ getCategoryLabel(item.category) }}
              </span>
              <span
                class="admin-inline-chip"
                :style="{
                  backgroundColor: getDifficultyColor(item.difficulty) + '16',
                  color: getDifficultyColor(item.difficulty),
                }"
              >
                {{ getDifficultyLabel(item.difficulty) }}
              </span>
              <span class="admin-inline-chip admin-inline-chip-neutral">{{ item.points }} pts</span>
            </div>

            <div class="queue-row__details">
              <div class="queue-row__detail-label">创建时间</div>
              <div class="queue-row__detail-value">{{ formatDateTime(item.created_at) }}</div>
            </div>

            <div class="queue-row__actions" role="group" aria-label="导入任务操作">
              <button class="admin-btn admin-btn-ghost admin-btn-compact" @click="inspectImportTask(item)">
                继续查看预览
              </button>
            </div>
          </article>
        </div>
      </section>
    </main>
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

.workspace-topbar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem 1rem;
  padding-bottom: 0.85rem;
}

.topbar-leading {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.65rem;
}

.workspace-overline,
.journal-note-label,
.journal-eyebrow {
  font-size: 0.7rem;
  font-weight: 700;
  letter-spacing: 0.2em;
  text-transform: uppercase;
  color: var(--journal-accent);
}

.class-chip {
  display: inline-flex;
  align-items: center;
  min-height: 30px;
  border-radius: 999px;
  border: 1px solid color-mix(in srgb, var(--journal-accent) 26%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 10%, transparent);
  padding: 0.25rem 0.7rem;
  font-size: 0.76rem;
  font-weight: 600;
  color: var(--journal-accent);
}

.top-note {
  display: flex;
  flex-wrap: wrap;
  gap: 0.55rem 1rem;
  font-size: 0.82rem;
  color: var(--journal-muted);
}

.top-tabs {
  display: flex;
  gap: 28px;
  margin-top: 10px;
  margin-left: -1.5rem;
  margin-right: -1.5rem;
  padding: 0 1.5rem;
  border-bottom: 1px solid color-mix(in srgb, var(--journal-ink) 10%, transparent);
  overflow-x: auto;
  scrollbar-width: none;
}

.top-tabs::-webkit-scrollbar {
  display: none;
}

.top-tab {
  position: relative;
  display: inline-flex;
  align-items: center;
  min-height: 52px;
  padding: 10px 0 13px;
  border: 0;
  border-bottom: 2px solid transparent;
  background: transparent;
  font-size: 15px;
  font-weight: 600;
  line-height: 1;
  color: color-mix(in srgb, var(--journal-muted) 88%, var(--color-bg-base));
  white-space: nowrap;
  cursor: pointer;
  transition:
    border-color 0.16s ease,
    color 0.16s ease;
}

.top-tab:hover,
.top-tab.active,
.top-tab:focus-visible {
  color: color-mix(in srgb, var(--journal-accent) 74%, var(--journal-ink));
  border-bottom-color: color-mix(in srgb, var(--journal-accent) 86%, var(--journal-ink));
  outline: none;
}

.content-pane {
  display: flex;
  flex: 1 1 auto;
  flex-direction: column;
  padding-top: 1.5rem;
}

.tab-panel {
  min-width: 0;
}

.journal-note {
  padding: 0 0 0 1rem;
  border-left: 2px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
  background: transparent;
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

.manage-title--compact {
  margin-top: 0.35rem;
  font-size: clamp(24px, 3vw, 34px);
}

.manage-summary-grid {
  display: grid;
  gap: 0.85rem;
  grid-template-columns: repeat(4, minmax(0, 1fr));
}

.list-heading,
.admin-section-head {
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

.challenge-list {
  --challenge-list-columns:
    minmax(16rem, 1.7fr)
    minmax(6.5rem, 0.68fr)
    minmax(6.5rem, 0.68fr)
    minmax(5.6rem, 0.58fr)
    minmax(7rem, 0.72fr)
    minmax(7rem, 0.76fr)
    minmax(9.5rem, 9.5rem);
  display: grid;
  gap: 0;
}

.manage-directory-head {
  display: grid;
  grid-template-columns: var(--challenge-list-columns);
  gap: 1rem;
  padding: 0 0 0.8rem;
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.16em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.manage-directory-head > span {
  min-width: 0;
}

.manage-directory-head__actions {
  text-align: right;
}

.challenge-row {
  display: grid;
  grid-template-columns: var(--challenge-list-columns);
  align-items: start;
  gap: 1rem;
  padding: 1rem 0;
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
}

.challenge-row > div {
  min-width: 0;
}

.queue-row__identity {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr);
  align-items: start;
  gap: 0.85rem;
  min-width: 0;
}

.challenge-row__identity {
  min-width: 0;
}

.challenge-row__title {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.challenge-row__category,
.challenge-row__difficulty,
.challenge-row__points,
.challenge-row__status,
.queue-row__summary {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  align-content: start;
}

.challenge-row__points,
.challenge-row__status {
  min-width: 0;
}

.challenge-row__category,
.challenge-row__difficulty,
.challenge-row__points,
.challenge-row__status,
.challenge-row__review {
  justify-self: start;
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
  display: -webkit-box;
  font-size: 0.82rem;
  line-height: 1.6;
  color: var(--color-danger);
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.challenge-row__actions {
  display: flex;
  flex-wrap: nowrap;
  justify-content: flex-end;
  align-items: flex-start;
  gap: 0.5rem;
  position: relative;
  justify-self: end;
}

.queue-row__actions {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  align-items: flex-start;
  gap: 0.5rem;
  position: relative;
}

.challenge-row__actions-menu {
  position: absolute;
  top: calc(100% + 0.4rem);
  right: 0;
  z-index: 10;
  display: grid;
  gap: 0.45rem;
  min-width: 10rem;
  padding: 0.6rem;
  border: 1px solid color-mix(in srgb, var(--journal-border) 92%, transparent);
  border-radius: 0.9rem;
  background:
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--journal-surface) 98%, var(--color-bg-base)),
      color-mix(in srgb, var(--journal-surface-subtle) 96%, var(--color-bg-base))
    );
  box-shadow: 0 16px 32px var(--color-shadow-soft);
}

.challenge-row__menu-button {
  justify-content: flex-start;
  width: 100%;
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
  background: color-mix(in srgb, var(--journal-surface) 96%, var(--color-bg-base));
  color: var(--journal-ink);
}

.admin-btn-success {
  border: 1px solid color-mix(in srgb, #059669 30%, transparent);
  background: color-mix(in srgb, #059669 12%, var(--journal-surface));
  color: #047857;
}

.admin-btn-danger {
  border: 1px solid color-mix(in srgb, #dc2626 28%, transparent);
  background: color-mix(in srgb, #dc2626 10%, var(--journal-surface));
  color: #b91c1c;
}

.admin-btn:hover,
.admin-btn:focus-visible {
  outline: none;
  border-color: color-mix(in srgb, var(--journal-accent) 42%, transparent);
}

.admin-inline-chip,
.admin-status-chip {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  border-radius: 999px;
  padding: 0.28rem 0.7rem;
  font-size: 0.78rem;
  font-weight: 700;
}

.admin-inline-chip-neutral {
  background: color-mix(in srgb, var(--journal-border) 34%, transparent);
  color: var(--journal-muted);
}

.admin-empty {
  padding: 1rem 0;
  color: var(--journal-muted);
}

.panel-copy {
  margin: 0.5rem 0 0;
  max-width: 54rem;
  color: var(--journal-muted);
  line-height: 1.7;
}

.sample-guide {
  display: grid;
  gap: 1rem;
  padding: 1rem 1.1rem 1.1rem;
  border: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
  border-radius: 1rem;
  background: linear-gradient(
    180deg,
    color-mix(in srgb, var(--journal-surface) 96%, var(--color-bg-base)),
    color-mix(in srgb, var(--journal-surface-subtle) 94%, var(--color-bg-base))
  );
}

.sample-guide__header {
  display: grid;
  gap: 0.7rem;
}

.sample-guide__eyebrow,
.sample-guide__link {
  font-size: 0.72rem;
  font-weight: 700;
}

.sample-guide__eyebrow {
  letter-spacing: 0.16em;
  text-transform: uppercase;
  color: var(--journal-accent);
}

.sample-guide__title {
  margin: 0;
  font-size: 1.05rem;
  font-weight: 700;
  color: var(--journal-ink);
}

.sample-guide__copy {
  margin: 0;
  color: var(--journal-muted);
  line-height: 1.65;
}

.sample-guide__actions {
  display: flex;
  align-items: center;
}

.sample-guide__link {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 2.75rem;
  padding: 0.72rem 1rem;
  border: 1px solid color-mix(in srgb, var(--journal-accent) 40%, var(--journal-border));
  border-radius: 0.85rem;
  background: color-mix(in srgb, var(--journal-accent) 12%, transparent);
  color: var(--journal-ink);
  transition:
    border-color 0.18s ease,
    background 0.18s ease,
    transform 0.18s ease;
}

.sample-guide__link:hover {
  border-color: color-mix(in srgb, var(--journal-accent) 64%, var(--journal-border));
  background: color-mix(in srgb, var(--journal-accent) 18%, transparent);
  transform: translateY(-1px);
}

.queue-list {
  display: grid;
  gap: 0.75rem;
}

.queue-row {
  display: grid;
  grid-template-columns: minmax(0, 1.4fr) minmax(0, 1fr) minmax(8rem, 0.8fr) auto;
  gap: 1rem;
  align-items: start;
  padding: 1rem 0;
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
}

.queue-row__title {
  margin: 0;
  font-size: 1rem;
  font-weight: 700;
  color: var(--journal-ink);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.queue-row__meta-text,
.queue-row__detail-label {
  margin: 0.3rem 0 0;
  font-size: 0.82rem;
  color: var(--journal-muted);
}

.queue-row__meta-text {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.queue-row__details {
  display: grid;
  gap: 0.35rem;
}

.queue-row__detail-value {
  color: var(--journal-ink);
  line-height: 1.6;
}

@media (max-width: 1200px) {
  .manage-summary-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .challenge-list {
    --challenge-list-columns:
      minmax(13rem, 1.45fr)
      minmax(5.4rem, 0.6fr)
      minmax(5.4rem, 0.6fr)
      minmax(5rem, 0.54fr)
      minmax(6.2rem, 0.66fr)
      minmax(6.2rem, 0.7fr)
      minmax(8.6rem, 8.6rem);
  }

  .queue-row {
    grid-template-columns: minmax(0, 1fr);
  }

  .queue-row__actions {
    justify-content: flex-start;
  }
}

@media (max-width: 960px) {
  .manage-directory-head {
    display: none;
  }

  .challenge-row {
    grid-template-columns: minmax(0, 1fr);
  }

  .challenge-row__actions {
    justify-content: flex-start;
  }
}

@media (max-width: 720px) {
  .top-tabs {
    gap: 18px;
    margin-left: -1rem;
    margin-right: -1rem;
    padding: 0 1rem;
  }

  .manage-summary-grid {
    grid-template-columns: 1fr;
  }

  .admin-pagination {
    align-items: flex-start;
    flex-direction: column;
  }
}
</style>
