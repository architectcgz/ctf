<script setup lang="ts">
import type { AdminAwdChallengeImportPreview } from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'
import ChallengePackageImportEntry from '@/components/platform/challenge/ChallengePackageImportEntry.vue'
import type { PlatformAwdChallengeImportUploadResult } from '@/features/platform-awd-challenges'

const props = defineProps<{
  uploading: boolean
  queueLoading: boolean
  importQueue: AdminAwdChallengeImportPreview[]
  uploadResults: PlatformAwdChallengeImportUploadResult[]
  selectedFileName?: string
}>()

const emit = defineEmits<{
  selectImportPackages: [files: File[]]
  commitImport: [preview: AdminAwdChallengeImportPreview]
}>()

function handleSelectImportPackages(files: File[]) {
  emit('selectImportPackages', files)
}

function getImageSourceLabel(
  value?: NonNullable<AdminAwdChallengeImportPreview['image_delivery']>['source_type']
): string {
  switch (value) {
    case 'platform_build':
      return '平台构建'
    case 'external_ref':
      return '外部镜像'
    case 'manual':
      return '手工登记'
    default:
      return '未声明'
  }
}

function getImageStatusLabel(
  value?: NonNullable<AdminAwdChallengeImportPreview['image_delivery']>['build_status']
): string {
  switch (value) {
    case 'available':
      return '可用'
    case 'building':
      return '构建中'
    case 'pushed':
      return '已推送'
    case 'verifying':
      return '校验中'
    case 'failed':
      return '失败'
    case 'pending':
      return '等待中'
    default:
      return '待处理'
  }
}

function getImportTargetImageRef(item: AdminAwdChallengeImportPreview): string {
  const runtimeImageRef = item.runtime_config?.image_ref
  return (
    item.image_delivery?.target_image_ref ||
    (typeof runtimeImageRef === 'string' ? runtimeImageRef : '未生成')
  )
}

function formatStructuredJSON(value?: Record<string, unknown>): string {
  if (!value || Object.keys(value).length === 0) {
    return '{}'
  }
  return JSON.stringify(value, null, 2)
}
</script>

<template>
  <div class="awd-import-pane">
    <section class="workspace-directory-section awd-import-tool-section">
      <header class="list-heading awd-challenge-import__head">
        <div>
          <div class="workspace-overline">Ingestion</div>
          <h2 class="list-heading__title">导入 AWD 题目包</h2>
          <p class="hero-summary awd-challenge-import__copy">
            教师按统一题目包规范写好 `challenge.yml` 后，从这里导入 AWD 题目。
          </p>
        </div>
        <div class="awd-challenge-import__head-actions">
          <div class="quick-actions">
            <a
              class="ui-btn ui-btn--ghost"
              href="/downloads/awd-challenge-package-sample-v1.zip"
              download="awd-challenge-package-sample-v1.zip"
            >
              下载示例题包
            </a>
          </div>
        </div>
      </header>

      <div class="awd-challenge-import__entry">
        <ChallengePackageImportEntry
          :hide-header="true"
          :uploading="uploading"
          :selected-file-name="selectedFileName"
          @select="handleSelectImportPackages"
        />
      </div>

      <div v-if="uploadResults.length > 0" class="awd-challenge-import__uploads">
        <article
          v-for="item in uploadResults"
          :key="item.id"
          class="awd-challenge-import__upload"
          :class="item.status === 'success' ? 'is-success' : 'is-error'"
        >
          <div class="awd-challenge-import__upload-head">
            <strong>{{ item.fileName }}</strong>
            <span>{{ item.status === 'success' ? '成功' : '失败' }}</span>
          </div>
          <p>{{ item.message }}</p>
        </article>
      </div>
    </section>

    <section class="workspace-directory-section awd-import-queue-section">
      <header class="list-heading awd-challenge-import__queue-head">
        <div>
          <div class="workspace-overline">Review Queue</div>
          <h2 class="list-heading__title">待确认题目包</h2>
        </div>
        <span class="awd-challenge-import__queue-count">共 {{ importQueue.length }} 个待确认包</span>
      </header>

      <div v-if="queueLoading" class="awd-challenge-import__state">正在同步导入队列...</div>
      <AppEmpty
        v-else-if="importQueue.length === 0"
        class="awd-challenge-import__empty"
        icon="Box"
        title="队列为空"
        description="上传题目包后，待确认的项将出现在此处。"
      />
      <div v-else class="workspace-directory-list awd-challenge-import__queue">
        <article
          v-for="item in importQueue"
          :key="item.id"
          class="awd-challenge-import__card"
        >
          <div class="awd-challenge-import__card-head">
            <div>
              <h3 class="awd-challenge-import__card-title">
                {{ item.title }}
              </h3>
              <p class="awd-challenge-import__card-file">
                {{ item.file_name }}
              </p>
            </div>
            <button
              type="button"
              class="ui-btn ui-btn--primary"
              @click="emit('commitImport', item)"
            >
              确认导入
            </button>
          </div>

          <div class="awd-challenge-import__chips">
            <span class="awd-status-pill awd-status-pill--primary">{{ item.service_type }}</span>
            <span class="awd-status-pill awd-status-pill--warning">{{ item.deployment_mode }}</span>
            <span class="awd-status-pill awd-status-pill--muted">{{
              item.flag_mode || '未定义 flag_mode'
            }}</span>
            <span class="awd-status-pill awd-status-pill--success">{{
              item.defense_entry_mode || '未定义入口'
            }}</span>
          </div>

          <div class="awd-challenge-import__image">
            <span>{{ getImageSourceLabel(item.image_delivery?.source_type) }}</span>
            <strong :title="getImportTargetImageRef(item)">{{ getImportTargetImageRef(item) }}</strong>
            <span>{{ getImageStatusLabel(item.image_delivery?.build_status) }}</span>
          </div>
          <p
            v-if="item.image_delivery?.last_error"
            class="awd-challenge-import__image-error"
          >
            {{ item.image_delivery.last_error }}
          </p>

          <div class="awd-challenge-import__grid">
            <pre class="awd-challenge-import__json">{{ formatStructuredJSON(item.access_config) }}</pre>
            <pre class="awd-challenge-import__json">{{
              formatStructuredJSON(item.runtime_config)
            }}</pre>
          </div>
        </article>
      </div>
    </section>
  </div>
</template>

<style scoped>
.awd-import-pane {
  display: flex;
  flex-direction: column;
  gap: var(--workspace-directory-page-block-gap, var(--space-5));
}

.hero-summary {
  max-width: 760px;
  margin-top: var(--space-3-5);
  font-size: var(--font-size-15);
  line-height: 1.9;
  color: var(--journal-muted);
}

.quick-actions {
  gap: var(--space-3);
}

.awd-challenge-import__uploads {
  display: grid;
  gap: var(--space-3);
  margin-top: var(--space-4);
}

.awd-challenge-import__upload {
  padding: var(--space-4);
  border-radius: 1rem;
  border: 1px solid var(--color-border-default);
  background: var(--color-bg-surface);
}

.awd-challenge-import__upload.is-success {
  border-color: color-mix(in srgb, var(--color-success) 24%, transparent);
}

.awd-challenge-import__upload-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
  margin-bottom: var(--space-2);
}

.awd-challenge-import__entry {
  margin-top: var(--workspace-directory-gap-top);
}

.awd-challenge-import__queue-head {
  margin-bottom: 0;
}

.awd-challenge-import__queue-count {
  color: var(--color-text-muted);
  font-size: var(--font-size-13);
  font-weight: 700;
}

.awd-challenge-import__queue {
  display: grid;
  gap: 0;
  padding: 0;
}

.awd-challenge-import__card {
  display: grid;
  gap: var(--space-4);
  padding: var(--space-4-5) var(--space-5);
  border-bottom: 1px solid var(--workspace-directory-row-divider);
}

.awd-challenge-import__card:last-child {
  border-bottom: 0;
}

.awd-challenge-import__card-title {
  margin: 0;
  font-size: var(--font-size-17);
  font-weight: 800;
  color: var(--color-text-primary);
}

.awd-challenge-import__card-file {
  margin: var(--space-1) 0 0;
  color: var(--color-text-muted);
  font-family: var(--font-family-mono);
  font-size: var(--font-size-12);
}

.awd-challenge-import__chips {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
}

.awd-challenge-import__image {
  display: grid;
  grid-template-columns: max-content minmax(0, 1fr) max-content;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-3);
  border: 1px solid var(--color-border-subtle);
  border-radius: 0.75rem;
  background: color-mix(in srgb, var(--color-bg-elevated) 84%, transparent);
  font-size: var(--font-size-13);
  color: var(--color-text-muted);
}

.awd-challenge-import__image strong {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-family: var(--font-family-mono);
  color: var(--color-text-primary);
}

.awd-challenge-import__image-error {
  margin: calc(var(--space-2) * -1) 0 0;
  color: var(--color-danger);
  font-size: var(--font-size-12);
}

.awd-challenge-import__grid {
  display: grid;
  gap: var(--space-3);
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.awd-challenge-import__json {
  margin: 0;
  min-height: 10rem;
  padding: var(--space-4);
  border-radius: 1rem;
  background: var(--color-bg-elevated);
  border: 1px solid var(--color-border-subtle);
  color: var(--color-text-secondary);
  font-family: var(--font-family-mono);
  font-size: var(--font-size-12);
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-word;
}

@media (max-width: 1024px) {
  .awd-challenge-import__grid {
    grid-template-columns: 1fr;
  }
}
</style>
