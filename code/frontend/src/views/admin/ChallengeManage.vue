<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import AdminPaginationControls from '@/components/admin/AdminPaginationControls.vue'
import ChallengePackageImportEntry from '@/components/admin/challenge/ChallengePackageImportEntry.vue'
import ChallengePackageImportReview from '@/components/admin/challenge/ChallengePackageImportReview.vue'
import { useAdminChallenges } from '@/composables/useAdminChallenges'
import { useChallengeManagePresentation } from '@/composables/useChallengeManagePresentation'
import { useChallengePackageImport } from '@/composables/useChallengePackageImport'
import { useRouteQueryTabs } from '@/composables/useRouteQueryTabs'

type ChallengePanelKey = 'manage' | 'import' | 'queue'

const panelTabs: Array<{ key: ChallengePanelKey; label: string; panelId: string; tabId: string }> =
  [
    {
      key: 'manage',
      label: '题目管理',
      panelId: 'challenge-panel-manage',
      tabId: 'challenge-tab-manage',
    },
    {
      key: 'import',
      label: '导入题目包',
      panelId: 'challenge-panel-import',
      tabId: 'challenge-tab-import',
    },
    {
      key: 'queue',
      label: '待确认导入',
      panelId: 'challenge-panel-queue',
      tabId: 'challenge-tab-queue',
    },
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
  uploadResults,
  hasPreview,
  refreshQueue,
  selectPackages,
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

const panelTabOrder = panelTabs.map((tab) => tab.key) as ChallengePanelKey[]
const {
  activeTab: activePanel,
  setTabButtonRef,
  selectTab: switchPanel,
  handleTabKeydown,
} = useRouteQueryTabs<ChallengePanelKey>({
  route,
  router,
  orderedTabs: panelTabOrder,
  defaultTab: 'manage',
  routeName: 'ChallengeManage',
})

const queueCount = computed(() => queue.value.length)
const {
  openActionMenuId,
  getCategoryLabel,
  getCategoryColor,
  getDifficultyLabel,
  getDifficultyColor,
  getStatusLabel,
  getStatusColor,
  getPublishRequestLabel,
  getPublishRequestColor,
  formatDateTime,
  inspectImportTask,
  toggleActionMenu,
  closeActionMenu,
  openChallengeDetail,
  openChallengeTopology,
  openChallengeWriteup,
  submitPublishCheck,
  removeChallenge,
} = useChallengeManagePresentation({
  router,
  switchToImportPanel: () => switchPanel('import'),
  loadPreview,
  publish,
  remove,
})

async function handleSelectPackage(files: File[]) {
  await selectPackages(files, { parallel: files.length > 1 })
}

async function handleCommitPreview() {
  await commitPreview()
}

async function openPackageFormatGuide(): Promise<void> {
  await router.push({ name: 'AdminChallengePackageFormat' })
}

onMounted(() => {
  void refreshQueue()
})
</script>

<template>
  <section
    class="journal-shell journal-shell-admin journal-notes-card journal-hero flex min-h-full flex-1 flex-col rounded-[24px] border px-6 py-6 md:px-8"
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

    <nav class="top-tabs" role="tablist" aria-label="题目管理视图切换">
      <button
        v-for="(tab, index) in panelTabs"
        :id="tab.tabId"
        :key="tab.tabId"
        :ref="(element) => setTabButtonRef(tab.key, element as HTMLButtonElement | null)"
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
          <div class="manage-header__intro workspace-tab-heading__main">
            <div class="journal-note-label">Challenge Library</div>
            <h1 class="workspace-tab-heading__title">题目管理</h1>
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

          <div
            v-if="loading"
            class="workspace-directory-loading flex items-center justify-center py-12"
          >
            <div
              class="h-8 w-8 animate-spin rounded-full border-4 border-[var(--journal-border)] border-t-[var(--journal-accent)]"
            />
          </div>

          <template v-else>
            <div v-if="list.length === 0" class="admin-empty workspace-directory-empty">
              当前还没有题目，请先导入题目包。
            </div>

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
                        class="challenge-row__title"
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
                  <span class="admin-inline-chip admin-inline-chip-neutral"
                    >{{ row.points }} pts</span
                  >
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
              <AdminPaginationControls
                :page="page"
                :total-pages="Math.max(1, Math.ceil(total / pageSize))"
                :total="total"
                :total-label="`共 ${total} 条`"
                @change-page="void changePage($event)"
              />
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
        <div class="workspace-tab-heading">
          <div class="workspace-tab-heading__main">
            <div class="journal-note-label">Challenge Package</div>
            <h1 class="workspace-tab-heading__title">导入题目包</h1>
          </div>
        </div>

        <ChallengePackageImportEntry
          :hide-header="true"
          :uploading="uploading"
          :selected-file-name="selectedFileName"
          @select="handleSelectPackage"
        >
          <template #before-dropzone>
            <section class="sample-guide">
              <div class="sample-guide__header">
                <div>
                  <div class="sample-guide__eyebrow">Uploader Guide</div>
                  <h2 class="sample-guide__title">题目包示例</h2>
                </div>
                <p class="sample-guide__copy">
                  导入页只保留上传和预览流程，目录结构与 `challenge.yml`
                  示例统一放到独立说明页，避免同一份规则重复维护。
                </p>
              </div>

              <div class="sample-guide__actions">
                <a
                  class="sample-guide__link"
                  data-testid="challenge-package-download-link"
                  href="/downloads/challenge-package-sample-v1.zip"
                  download="challenge-package-sample-v1.zip"
                >
                  下载示例题目包
                </a>
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
          </template>
        </ChallengePackageImportEntry>

        <section v-if="uploadResults.length > 0" class="upload-result-panel" aria-live="polite">
          <div class="upload-result-panel__header">
            <div class="upload-result-panel__eyebrow">Upload Result</div>
            <h2 class="upload-result-panel__title">最近上传结果</h2>
          </div>

          <div class="upload-result-list">
            <article
              v-for="result in uploadResults"
              :key="result.id"
              class="upload-result-item"
              :class="{
                'upload-result-item--success': result.status === 'success',
                'upload-result-item--error': result.status === 'error',
              }"
            >
              <header class="upload-result-item__head">
                <span
                  class="upload-result-item__status"
                  :class="{
                    'upload-result-item__status--success': result.status === 'success',
                    'upload-result-item__status--error': result.status === 'error',
                  }"
                >
                  {{ result.status === 'success' ? '成功' : '失败' }}
                </span>
                <strong class="upload-result-item__name" :title="result.fileName">{{ result.fileName }}</strong>
              </header>

              <p class="upload-result-item__message">{{ result.message }}</p>

              <div class="upload-result-item__meta">
                <span>{{ formatDateTime(result.createdAt) }}</span>
                <span v-if="result.code !== undefined">错误码 {{ result.code }}</span>
                <span v-if="result.requestId">请求ID {{ result.requestId }}</span>
              </div>
            </article>
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
        <div class="workspace-tab-heading">
          <div class="workspace-tab-heading__main">
            <div class="journal-note-label">Import Review</div>
            <h1 class="workspace-tab-heading__title">待确认导入</h1>
          </div>
        </div>

        <p class="workspace-tab-copy">
          这里列出已生成预览、但还没正式导入题库的题目包。确认无误后，可继续查看预览并完成导入。
        </p>

        <div v-if="queueLoading" class="flex items-center justify-center py-12">
          <div
            class="h-8 w-8 animate-spin rounded-full border-4 border-[var(--journal-border)] border-t-[var(--journal-accent)]"
          />
        </div>

        <div v-else-if="queue.length === 0" class="admin-empty">当前没有待确认的导入任务。</div>

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
              <button
                class="admin-btn admin-btn-ghost admin-btn-compact"
                @click="inspectImportTask(item)"
              >
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
  --admin-summary-grid-columns: repeat(4, minmax(0, 1fr));
  --journal-topbar-padding-bottom: var(--space-3);
  --page-top-tabs-gap: var(--space-7);
  --page-top-tabs-margin: var(--space-2-5) calc(var(--space-6) * -1) 0;
  --page-top-tabs-padding: 0 var(--space-6);
  --page-top-tabs-border: color-mix(in srgb, var(--journal-ink) 10%, transparent);
  --page-top-tab-min-height: 52px;
  --page-top-tab-padding: var(--space-2-5) 0 var(--space-3-5);
  --page-top-tab-font-size: var(--font-size-15);
  --page-top-tab-active-color: color-mix(in srgb, var(--journal-accent) 74%, var(--journal-ink));
  --page-top-tab-active-border: color-mix(in srgb, var(--journal-accent) 86%, var(--journal-ink));
  --journal-note-value-weight: 700;
  --journal-shell-hero-radial-strength: 7%;
  --journal-shell-hero-radial-size: 22rem;
  --journal-shell-hero-end: var(--journal-surface);
  --journal-shell-hero-shadow: 0 22px 50px var(--color-shadow-soft);
}

.workspace-overline {
  font-size: var(--font-size-0-70);
  font-weight: 700;
  letter-spacing: 0.2em;
  text-transform: uppercase;
  color: var(--journal-accent);
}

.content-pane {
  display: flex;
  flex: 1 1 auto;
  flex-direction: column;
}

.manage-header {
  display: grid;
  gap: var(--space-6);
  padding-bottom: var(--space-6);
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
}

.list-heading {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-3);
  align-items: flex-end;
}

.list-heading__title {
  margin: var(--space-1) 0 0;
  font-size: var(--font-size-1-20);
  font-weight: 700;
  color: var(--journal-ink);
}

.challenge-list {
  --challenge-list-columns: minmax(16rem, 1.7fr) minmax(6.5rem, 0.68fr) minmax(6.5rem, 0.68fr)
    minmax(5.6rem, 0.58fr) minmax(7rem, 0.72fr) minmax(7rem, 0.76fr) minmax(9.5rem, 9.5rem);
  display: grid;
  gap: 0;
}

.manage-directory-head {
  display: grid;
  grid-template-columns: var(--challenge-list-columns);
  gap: var(--space-4);
  padding: 0 0 var(--space-3);
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
  font-size: var(--font-size-0-72);
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
  gap: var(--space-4);
  padding: var(--space-4) 0;
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
}

.challenge-row > div {
  min-width: 0;
}

.queue-row__identity {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr);
  align-items: start;
  gap: var(--space-3);
  min-width: 0;
}

.challenge-row__identity {
  min-width: 0;
}

.challenge-row__title {
  min-width: 0;
  margin: 0;
  font-size: var(--font-size-0-90);
  font-weight: 600;
  line-height: 1.35;
  color: var(--journal-ink);
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
  gap: var(--space-2);
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
  gap: var(--space-1-5);
  min-width: 0;
}

.challenge-row__review-status {
  font-size: var(--font-size-0-92);
  font-weight: 600;
  line-height: 1.5;
}

.challenge-row__failure {
  margin-top: var(--space-3);
  display: -webkit-box;
  font-size: var(--font-size-0-82);
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
  gap: var(--space-2);
  position: relative;
  justify-self: end;
}

.queue-row__actions {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  align-items: flex-start;
  gap: var(--space-2);
  position: relative;
}

.challenge-row__actions-menu {
  position: absolute;
  top: calc(100% + var(--space-1-5));
  right: 0;
  z-index: 10;
  display: grid;
  gap: var(--space-2);
  min-width: 10rem;
  padding: var(--space-2-5);
  border: 1px solid color-mix(in srgb, var(--journal-border) 92%, transparent);
  border-radius: 0.9rem;
  background: linear-gradient(
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
  gap: var(--space-2);
  min-height: 2.45rem;
  border-radius: 0.75rem;
  padding: var(--space-2) var(--space-4);
  font-size: var(--font-size-0-875);
  font-weight: 600;
  transition:
    border-color 150ms ease,
    background 150ms ease,
    color 150ms ease;
}

.admin-btn-compact {
  min-height: 2.25rem;
  padding: var(--space-2) var(--space-3);
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
  gap: var(--space-1-5);
  border-radius: 999px;
  padding: var(--space-1) var(--space-3);
  font-size: var(--font-size-0-78);
  font-weight: 700;
}

.admin-inline-chip-neutral {
  background: color-mix(in srgb, var(--journal-border) 34%, transparent);
  color: var(--journal-muted);
}

.admin-empty {
  padding: var(--space-4) 0;
  color: var(--journal-muted);
}

.sample-guide {
  display: grid;
  gap: var(--space-4);
  padding: var(--space-4) var(--space-4-5) var(--space-4-5);
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
  gap: var(--space-3);
}

.sample-guide__eyebrow,
.sample-guide__link {
  font-size: var(--font-size-0-72);
  font-weight: 700;
}

.sample-guide__eyebrow {
  letter-spacing: 0.16em;
  text-transform: uppercase;
  color: var(--journal-accent);
}

.sample-guide__title {
  margin: 0;
  font-size: var(--font-size-1-05);
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
  flex-wrap: wrap;
  gap: var(--space-2-5);
}

.sample-guide__link {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 2.75rem;
  padding: var(--space-3) var(--space-4);
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

.upload-result-panel {
  display: grid;
  gap: var(--space-3);
  padding: var(--space-1-5) 0 var(--space-1);
}

.upload-result-panel__header {
  display: grid;
  gap: var(--space-2);
}

.upload-result-panel__eyebrow {
  font-size: var(--font-size-0-72);
  font-weight: 700;
  letter-spacing: 0.16em;
  text-transform: uppercase;
  color: var(--journal-accent);
}

.upload-result-panel__title {
  margin: 0;
  font-size: var(--font-size-1-00);
  font-weight: 700;
  color: var(--journal-ink);
}

.upload-result-list {
  display: grid;
  gap: var(--space-2-5);
}

.upload-result-item {
  display: grid;
  gap: var(--space-2);
  padding: var(--space-3) var(--space-3-5);
  border: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
  border-radius: 0.85rem;
  background: color-mix(in srgb, var(--journal-surface) 96%, var(--color-bg-base));
}

.upload-result-item--success {
  border-color: color-mix(in srgb, #059669 30%, var(--journal-border));
}

.upload-result-item--error {
  border-color: color-mix(in srgb, #dc2626 34%, var(--journal-border));
}

.upload-result-item__head {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--space-2);
}

.upload-result-item__status {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 2.7rem;
  min-height: 1.5rem;
  padding: var(--space-0-5) var(--space-2);
  border-radius: 999px;
  font-size: var(--font-size-0-72);
  font-weight: 700;
  letter-spacing: 0.06em;
}

.upload-result-item__status--success {
  color: #047857;
  background: color-mix(in srgb, #059669 12%, transparent);
}

.upload-result-item__status--error {
  color: #b91c1c;
  background: color-mix(in srgb, #dc2626 12%, transparent);
}

.upload-result-item__name {
  min-width: 0;
  color: var(--journal-ink);
  font-size: var(--font-size-0-92);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.upload-result-item__message {
  margin: 0;
  color: var(--journal-ink);
  line-height: 1.55;
  font-size: var(--font-size-0-86);
}

.upload-result-item__meta {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-3);
  color: var(--journal-muted);
  font-size: var(--font-size-0-78);
}

.queue-list {
  display: grid;
  gap: var(--space-3);
}

.queue-row {
  display: grid;
  grid-template-columns: minmax(0, 1.4fr) minmax(0, 1fr) minmax(8rem, 0.8fr) auto;
  gap: var(--space-4);
  align-items: start;
  padding: var(--space-4) 0;
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
}

.queue-row__title {
  margin: 0;
  font-size: var(--font-size-1-00);
  font-weight: 700;
  color: var(--journal-ink);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.queue-row__meta-text,
.queue-row__detail-label {
  margin: var(--space-1) 0 0;
  font-size: var(--font-size-0-82);
  color: var(--journal-muted);
}

.queue-row__meta-text {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.queue-row__details {
  display: grid;
  gap: var(--space-1-5);
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
    --challenge-list-columns: minmax(13rem, 1.45fr) minmax(5.4rem, 0.6fr) minmax(5.4rem, 0.6fr)
      minmax(5rem, 0.54fr) minmax(6.2rem, 0.66fr) minmax(6.2rem, 0.7fr) minmax(8.6rem, 8.6rem);
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
    gap: var(--space-4-5);
    margin-left: calc(var(--space-4) * -1);
    margin-right: calc(var(--space-4) * -1);
    padding: 0 var(--space-4);
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
