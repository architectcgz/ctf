<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ArrowLeft, FileSearch, Plus } from 'lucide-vue-next'

import ChallengePackageImportEntry from '@/components/platform/challenge/ChallengePackageImportEntry.vue'
import { useChallengePackageImport } from '@/composables/useChallengePackageImport'

const router = useRouter()

const {
  uploading,
  queueLoading,
  selectedFileName,
  queue,
  uploadResults,
  refreshQueue,
  selectPackages,
} = useChallengePackageImport()

const queueCount = computed(() => queue.value.length)

const categoryLabels = {
  web: 'Web',
  pwn: 'Pwn',
  reverse: '逆向',
  crypto: '密码',
  misc: '杂项',
  forensics: '取证',
} as const

const difficultyLabels = {
  beginner: '入门',
  easy: '简单',
  medium: '中等',
  hard: '困难',
  insane: '地狱',
} as const

onMounted(() => {
  void refreshQueue()
})

async function handleSelectPackage(files: File[]) {
  const selectedPreview = await selectPackages(files, { parallel: files.length > 1 })
  if (!selectedPreview?.id) {
    return
  }

  await router.push({
    name: 'PlatformChallengeImportPreview',
    params: { importId: selectedPreview.id },
  })
}

async function openPackageFormatGuide(): Promise<void> {
  await router.push({ name: 'PlatformChallengePackageFormat' })
}

async function backToChallenges(): Promise<void> {
  await router.push({ name: 'ChallengeManage' })
}

async function inspectImportTask(importId: string): Promise<void> {
  await router.push({
    name: 'PlatformChallengeImportPreview',
    params: { importId },
  })
}

function getCategoryLabel(value: keyof typeof categoryLabels): string {
  return categoryLabels[value] ?? '杂项'
}

function getDifficultyLabel(value: keyof typeof difficultyLabels): string {
  return difficultyLabels[value] ?? '简单'
}

function formatDateTime(value: string): string {
  return new Date(value).toLocaleString('zh-CN')
}
</script>

<template>
  <div class="workspace-shell challenge-import-shell journal-shell journal-shell-admin journal-notes-card">
    <div class="workspace-grid">
      <main class="content-pane challenge-import-content">
        <section class="challenge-import-panel">
          <div class="workspace-tab-heading challenge-import-actions">
            <div class="workspace-tab-heading__main">
              <div class="workspace-overline">
                Challenge Import
              </div>
              <h1 class="workspace-page-title">
                导入资源包
              </h1>
              <p class="workspace-page-copy">
                在独立导入页完成上传、核对题目包规范，并处理待确认导入任务。
              </p>
            </div>
            <div class="challenge-import-hero-actions">
              <button
                type="button"
                class="challenge-import-action"
                @click="backToChallenges"
              >
                <ArrowLeft class="mr-1.5 h-3.5 w-3.5" />
                返回题目目录
              </button>
              <button
                type="button"
                class="challenge-import-action"
                @click="openPackageFormatGuide"
              >
                <FileSearch class="mr-1.5 h-3.5 w-3.5" />
                题目包规范
              </button>
              <a
                class="challenge-import-action challenge-import-action--primary"
                href="/downloads/challenge-package-sample-v1.zip"
                download="challenge-package-sample-v1.zip"
              >
                <Plus class="mr-1.5 h-4 w-4" />
                下载示例题目包
              </a>
            </div>
          </div>

          <section
            id="challenge-import-workspace"
            class="workspace-directory-section challenge-import-directory challenge-workspace-section"
          >
            <header class="list-heading challenge-section-heading">
              <div>
                <div class="workspace-overline">
                  Challenge Package
                </div>
                <h2 class="list-heading__title">
                  导入题目包
                </h2>
                <p class="challenge-section-copy">
                  上传压缩包后先进入预览，再确认是否写入题库。格式规范与示例已收敛到当前导入页。
                </p>
              </div>
              <div class="challenge-directory-meta">
                共 {{ queueCount }} 个待处理任务
              </div>
            </header>

            <ChallengePackageImportEntry
              :hide-header="true"
              :uploading="uploading"
              :selected-file-name="selectedFileName"
              @select="handleSelectPackage"
            />
          </section>

          <section class="workspace-directory-section challenge-import-directory challenge-plain-section">
            <div class="list-heading">
              <div>
                <div class="workspace-overline">
                  Upload Receipt
                </div>
                <h2 class="list-heading__title">
                  最近上传结果
                </h2>
              </div>
            </div>

            <div
              v-if="uploadResults.length === 0"
              class="challenge-directory-state"
            >
              还没有新的上传回执，选择题目包后会在这里显示解析结果。
            </div>

            <div
              v-else
              class="challenge-panel-stack"
            >
              <article
                v-for="result in uploadResults"
                :key="result.id"
                class="challenge-upload-result"
                :class="
                  result.status === 'success'
                    ? 'challenge-upload-result--success'
                    : 'challenge-upload-result--error'
                "
              >
                <div class="challenge-upload-result__head">
                  <span
                    class="challenge-upload-result__status"
                    :class="
                      result.status === 'success'
                        ? 'challenge-upload-result__status--success'
                        : 'challenge-upload-result__status--error'
                    "
                  >
                    {{ result.status === 'success' ? '成功' : '失败' }}
                  </span>
                  <strong
                    class="challenge-upload-result__title"
                    :title="result.fileName"
                  >
                    {{ result.fileName }}
                  </strong>
                </div>
                <p class="challenge-upload-result__copy">
                  {{ result.message }}
                </p>
                <div class="challenge-upload-result__meta">
                  <span>{{ formatDateTime(result.createdAt) }}</span>
                  <span v-if="result.code !== undefined">错误码 {{ result.code }}</span>
                  <span v-if="result.requestId">请求ID {{ result.requestId }}</span>
                </div>
              </article>
            </div>
          </section>

          <section
            id="challenge-queue-workspace"
            class="workspace-directory-section challenge-import-directory challenge-workspace-section"
          >
            <div class="list-heading challenge-directory-head challenge-section-heading">
              <div>
                <div class="workspace-overline">
                  Import Review
                </div>
                <h2 class="list-heading__title">
                  待确认导入
                </h2>
                <p class="challenge-section-copy">
                  这里列出已生成预览、但还没正式导入题库的题目包。确认无误后，再继续写入题库。
                </p>
              </div>
              <div class="challenge-directory-meta">
                共 {{ queueCount }} 个待处理任务
              </div>
            </div>

            <div
              v-if="queueLoading"
              class="challenge-directory-state"
            >
              正在同步导入队列...
            </div>
            <div
              v-else-if="queue.length === 0"
              class="challenge-directory-state"
            >
              当前没有待确认的导入任务。
            </div>

            <div
              v-else
              class="challenge-panel-stack"
            >
              <article
                v-for="item in queue"
                :key="item.id"
                class="challenge-plain-section challenge-queue-item"
              >
                <div class="flex min-w-0 items-start gap-4">
                  <div class="challenge-queue-id">
                    IMP-{{ item.id.slice(0, 6).toUpperCase() }}
                  </div>
                  <div class="min-w-0 flex-1">
                    <h2
                      class="challenge-queue-title"
                      :title="item.title"
                    >
                      {{ item.title }}
                    </h2>
                    <p
                      class="challenge-queue-file"
                      :title="item.file_name"
                    >
                      {{ item.file_name }}
                    </p>
                    <div class="mt-3 flex flex-wrap gap-2">
                      <span class="challenge-table-pill challenge-table-pill--category">
                        {{ getCategoryLabel(item.category) }}
                      </span>
                      <span class="challenge-table-pill challenge-table-pill--neutral">
                        {{ getDifficultyLabel(item.difficulty) }}
                      </span>
                      <span class="challenge-queue-points">{{ item.points }} pts</span>
                    </div>
                  </div>
                </div>

                <div class="flex flex-col items-start gap-2 md:items-end">
                  <div class="challenge-queue-time">
                    {{ formatDateTime(item.created_at) }}
                  </div>
                  <button
                    type="button"
                    class="challenge-import-action challenge-import-action--primary"
                    @click="inspectImportTask(item.id)"
                  >
                    继续查看预览
                  </button>
                </div>
              </article>
            </div>
          </section>
        </section>
      </main>
    </div>
  </div>
</template>

<style scoped>
.challenge-import-shell {
  --workspace-brand: #2563eb;
  --workspace-brand-ink: #1e40af;
  --challenge-page-bg: color-mix(in srgb, var(--journal-surface-subtle) 90%, var(--color-bg-base));
  --challenge-page-surface: color-mix(in srgb, var(--journal-surface) 96%, var(--color-bg-surface));
  --challenge-page-surface-subtle: color-mix(
    in srgb,
    var(--color-bg-elevated) 82%,
    var(--color-bg-surface)
  );
  --challenge-page-surface-elevated: color-mix(
    in srgb,
    var(--color-bg-elevated) 90%,
    var(--color-bg-surface)
  );
  --challenge-page-line: color-mix(in srgb, var(--journal-border) 84%, transparent);
  --challenge-page-line-strong: color-mix(in srgb, var(--journal-border) 92%, transparent);
  --challenge-page-text: color-mix(in srgb, var(--journal-ink) 94%, transparent);
  --challenge-page-muted: color-mix(in srgb, var(--journal-muted) 92%, transparent);
  --challenge-page-faint: color-mix(in srgb, var(--color-text-muted) 90%, transparent);
  --challenge-page-accent: color-mix(in srgb, var(--workspace-brand) 88%, var(--challenge-page-text));
  border: none;
  background: var(--challenge-page-bg);
}

.challenge-import-content {
  display: grid;
  gap: 2rem;
  background: transparent;
}

.challenge-import-panel {
  display: grid;
  gap: 2rem;
  min-width: 0;
}

.challenge-import-actions {
  align-items: flex-end;
}

.challenge-import-hero-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.challenge-workspace-section {
  scroll-margin-top: 6rem;
}

.challenge-import-directory {
  display: grid;
  gap: 1.5rem;
}

.challenge-section-heading {
  align-items: flex-start;
  gap: 1rem;
}

.challenge-section-copy {
  margin: 0.5rem 0 0;
  max-width: 44rem;
  font-size: 0.92rem;
  line-height: 1.6;
  color: var(--challenge-page-muted);
}

.challenge-panel-stack {
  display: grid;
  gap: 1rem;
}

.challenge-import-action {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 2.5rem;
  padding: 0 1.25rem;
  border: 1px solid var(--challenge-page-line);
  border-radius: 12px;
  background: var(--challenge-page-surface);
  font-size: 12px;
  font-weight: 700;
  color: var(--challenge-page-muted);
  transition:
    border-color 0.2s ease,
    background-color 0.2s ease,
    color 0.2s ease,
    transform 0.2s ease,
    box-shadow 0.2s ease;
  box-shadow: 0 1px 2px color-mix(in srgb, var(--color-shadow-soft) 42%, transparent);
}

.challenge-import-action:hover {
  border-color: var(--challenge-page-line-strong);
  background: var(--challenge-page-surface-elevated);
  color: var(--challenge-page-text);
  transform: translateY(-1px);
}

.challenge-import-action--primary {
  border-color: color-mix(in srgb, var(--workspace-brand) 42%, transparent);
  background: color-mix(in srgb, var(--workspace-brand) 88%, var(--challenge-page-text));
  color: white;
  box-shadow: 0 10px 24px color-mix(in srgb, var(--workspace-brand) 18%, transparent);
}

.challenge-import-action--primary:hover {
  color: white;
  background: color-mix(in srgb, var(--workspace-brand-ink) 92%, var(--challenge-page-text));
  border-color: color-mix(in srgb, var(--workspace-brand-ink) 62%, transparent);
}

.challenge-directory-state,
.challenge-directory-meta,
.challenge-queue-file,
.challenge-queue-time,
.challenge-upload-result__copy,
.challenge-upload-result__meta {
  color: var(--challenge-page-muted);
}

.challenge-table-pill {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  height: 1.4rem;
  padding: 0 0.5rem;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 800;
  letter-spacing: 0.02em;
  text-transform: uppercase;
}

.challenge-table-pill--category {
  background: color-mix(in srgb, var(--workspace-brand) 10%, var(--challenge-page-surface));
  color: var(--challenge-page-accent);
  border: 1px solid color-mix(in srgb, var(--workspace-brand) 18%, transparent);
}

.challenge-table-pill--neutral {
  background: color-mix(in srgb, var(--challenge-page-line) 18%, var(--challenge-page-surface));
  color: var(--challenge-page-muted);
  border: 1px solid color-mix(in srgb, var(--challenge-page-line-strong) 78%, transparent);
}

.challenge-queue-title,
.challenge-upload-result__title {
  margin: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 1rem;
  font-weight: 700;
  color: var(--challenge-page-text);
}

.challenge-queue-file {
  margin: 0.25rem 0 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 0.875rem;
}

.challenge-queue-id {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 5.5rem;
  height: 2rem;
  padding: 0 0.75rem;
  border: 1px solid color-mix(in srgb, var(--workspace-brand) 18%, var(--challenge-page-line));
  border-radius: 999px;
  background: color-mix(in srgb, var(--workspace-brand) 9%, var(--challenge-page-surface));
  font-family: var(--font-family-mono, ui-monospace, SFMono-Regular, monospace);
  font-size: 0.72rem;
  font-weight: 800;
  letter-spacing: 0.08em;
  color: var(--challenge-page-accent);
}

.challenge-queue-points {
  font-family: var(--font-family-mono, ui-monospace, SFMono-Regular, monospace);
  font-size: 0.72rem;
  font-weight: 700;
  color: var(--challenge-page-muted);
}

.challenge-upload-result {
  display: grid;
  gap: 0.5rem;
  padding: 1rem;
  border: 1px solid var(--challenge-page-line);
  border-radius: 1rem;
  background: var(--challenge-page-surface);
}

.challenge-upload-result--success {
  border-color: color-mix(in srgb, var(--color-success) 24%, var(--challenge-page-line));
  background: color-mix(in srgb, var(--color-success) 9%, var(--challenge-page-surface));
}

.challenge-upload-result--error {
  border-color: color-mix(in srgb, var(--color-danger) 24%, var(--challenge-page-line));
  background: color-mix(in srgb, var(--color-danger) 8%, var(--challenge-page-surface));
}

.challenge-upload-result__head {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.challenge-upload-result__status {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: 999px;
  padding: 0.125rem 0.5rem;
  font-size: 0.75rem;
  font-weight: 700;
}

.challenge-upload-result__status--success {
  background: color-mix(in srgb, var(--color-success) 16%, transparent);
  color: color-mix(in srgb, var(--color-success) 92%, var(--challenge-page-text));
}

.challenge-upload-result__status--error {
  background: color-mix(in srgb, var(--color-danger) 14%, transparent);
  color: color-mix(in srgb, var(--color-danger) 92%, var(--challenge-page-text));
}

.challenge-upload-result__copy {
  margin: 0;
  font-size: 0.8rem;
}

.challenge-upload-result__meta {
  display: flex;
  flex-wrap: wrap;
  gap: 1rem;
  font-size: 0.65rem;
  font-weight: 600;
}

:global([data-theme='light']) .challenge-import-shell {
  --challenge-page-bg: #f8fafc;
  --challenge-page-surface: white;
  --challenge-page-surface-subtle: #f8fafc;
  --challenge-page-surface-elevated: white;
  --challenge-page-line: color-mix(in srgb, #e2e8f0 90%, transparent);
  --challenge-page-line-strong: color-mix(in srgb, #d4dde8 94%, transparent);
  --challenge-page-text: #0f172a;
  --challenge-page-muted: #64748b;
  --challenge-page-faint: #94a3b8;
}

:global([data-theme='dark']) .challenge-import-shell {
  --challenge-page-bg: color-mix(in srgb, var(--color-bg-base) 92%, var(--color-bg-surface));
  --challenge-page-surface: color-mix(in srgb, var(--color-bg-surface) 92%, var(--color-bg-base));
  --challenge-page-surface-subtle: color-mix(in srgb, var(--color-bg-elevated) 84%, var(--color-bg-surface));
  --challenge-page-surface-elevated: color-mix(in srgb, var(--color-bg-elevated) 92%, var(--color-bg-surface));
  --challenge-page-line: color-mix(in srgb, var(--color-border-default) 88%, transparent);
  --challenge-page-line-strong: color-mix(in srgb, var(--color-border-default) 94%, transparent);
  --challenge-page-text: color-mix(in srgb, var(--color-text-primary) 94%, transparent);
  --challenge-page-muted: color-mix(in srgb, var(--color-text-secondary) 90%, transparent);
  --challenge-page-faint: color-mix(in srgb, var(--color-text-muted) 90%, transparent);
}
</style>
