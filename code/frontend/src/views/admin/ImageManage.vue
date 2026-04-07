<template>
  <section
    class="journal-shell journal-hero flex min-h-full flex-1 flex-col rounded-[30px] border px-6 py-6 md:px-8"
  >
    <header class="image-header">
      <div class="image-header__intro">
        <div class="journal-eyebrow">Image Registry</div>
        <h1 class="image-title">镜像管理</h1>
        <p class="image-copy">集中查看镜像构建状态、描述与创建时间。</p>
      </div>

      <div class="image-header__side">
        <div class="image-header__actions" role="group" aria-label="镜像列表操作">
          <button
            :disabled="loading"
            class="admin-btn admin-btn-ghost"
            data-testid="image-refresh-button"
            @click="handleManualRefresh"
          >
            立即刷新
          </button>
          <button class="admin-btn admin-btn-primary" @click="dialogVisible = true">
            创建镜像
          </button>
        </div>
        <div class="image-summary-grid">
          <article class="journal-note">
            <div class="journal-note-label">镜像总量</div>
            <div class="journal-note-value">{{ total }}</div>
            <div class="journal-note-helper">当前查询结果的镜像总数</div>
          </article>
          <article class="journal-note">
            <div class="journal-note-label">当前页</div>
            <div class="journal-note-value">{{ list.length }}</div>
            <div class="journal-note-helper">这一页已加载的镜像数量</div>
          </article>
        </div>
      </div>
    </header>
    <div class="journal-divider image-divider" />

    <section class="image-board workspace-directory-section">
      <div class="image-board__head">
        <div>
          <div class="journal-note-label">Images</div>
          <h2 class="image-section-title">镜像列表</h2>
        </div>
        <div class="image-board__hint">{{ refreshHint }}</div>
      </div>

      <div
        v-if="loading"
        class="workspace-directory-loading flex items-center justify-center py-12"
      >
        <div
          class="h-8 w-8 animate-spin rounded-full border-4 border-[var(--journal-border)] border-t-[var(--journal-accent)]"
        ></div>
      </div>

      <template v-else>
        <div v-if="list.length === 0" class="admin-empty workspace-directory-empty">
          当前还没有镜像。
        </div>

        <div v-else class="image-list workspace-directory-list">
          <div class="image-directory-head" aria-hidden="true">
            <span>镜像名称</span>
            <span>标签</span>
            <span>描述</span>
            <span>状态</span>
            <span>创建时间</span>
            <span class="image-directory-head__actions">操作</span>
          </div>

          <article v-for="row in list" :key="row.id" class="image-row">
            <div class="image-row__name" :title="row.name">{{ row.name }}</div>

            <div class="image-row__tag" :title="row.tag">{{ row.tag }}</div>

            <p class="image-row__description" :title="row.description || '未填写镜像说明'">
              {{ row.description || '未填写镜像说明' }}
            </p>

            <div class="image-row__status">
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

            <div class="image-row__time">{{ new Date(row.created_at).toLocaleString() }}</div>

            <div class="image-row__actions">
              <button
                class="admin-btn admin-btn-danger admin-btn-compact"
                @click="handleDelete(row.id)"
              >
                删除
              </button>
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

    <ElDialog v-model="dialogVisible" title="创建镜像" width="500px">
      <ElForm :model="form" label-width="100px">
        <ElFormItem label="镜像名称" required>
          <ElInput v-model="form.name" placeholder="例如：ubuntu" />
        </ElFormItem>
        <ElFormItem label="标签" required>
          <ElInput v-model="form.tag" placeholder="例如：22.04" />
        </ElFormItem>
        <ElFormItem label="描述">
          <ElInput
            v-model="form.description"
            type="textarea"
            :rows="3"
            placeholder="镜像说明（可选）"
          />
        </ElFormItem>
      </ElForm>
      <template #footer>
        <button class="admin-btn admin-btn-ghost admin-btn-compact" @click="dialogVisible = false">
          取消
        </button>
        <button
          :disabled="creating"
          class="admin-btn admin-btn-primary admin-btn-compact ml-2 disabled:cursor-not-allowed disabled:opacity-50"
          @click="handleCreate"
        >
          {{ creating ? '创建中...' : '创建' }}
        </button>
      </template>
    </ElDialog>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, reactive, ref, watch } from 'vue'
import { ElMessageBox } from 'element-plus'
import { getImages, createImage, deleteImage } from '@/api/admin'
import AdminPaginationControls from '@/components/admin/AdminPaginationControls.vue'
import { usePagination } from '@/composables/usePagination'
import { useToast } from '@/composables/useToast'
import type { ImageStatus } from '@/api/contracts'

const toast = useToast()
const dialogVisible = ref(false)
const creating = ref(false)
const form = reactive({
  name: '',
  tag: '',
  description: '',
})

const { list, total, page, pageSize, loading, changePage, refresh } = usePagination(getImages)

let pollTimer: number | null = null

const hasActiveImages = computed(() =>
  list.value.some((row) => row.status === 'pending' || row.status === 'building')
)

const refreshHint = computed(() =>
  hasActiveImages.value ? '构建中镜像会每 10 秒自动刷新' : '当前无进行中镜像，可手动刷新'
)

function stopPolling() {
  if (pollTimer !== null) {
    clearInterval(pollTimer)
    pollTimer = null
  }
}

function startPolling() {
  if (pollTimer !== null) return
  pollTimer = window.setInterval(() => {
    void refresh()
  }, 10000)
}

async function handleCreate() {
  if (!form.name || !form.tag) {
    toast.error('请填写完整信息')
    return
  }
  creating.value = true
  try {
    await createImage(form)
    toast.success('镜像创建成功')
    dialogVisible.value = false
    Object.assign(form, { name: '', tag: '', description: '' })
    refresh()
  } catch (error) {
    toast.error('创建失败')
  } finally {
    creating.value = false
  }
}

async function handleDelete(id: string) {
  try {
    await ElMessageBox.confirm('确定要删除此镜像吗？', '确认', { type: 'warning' })
    await deleteImage(id)
    toast.success('删除成功')
    refresh()
  } catch (error) {
    if (error !== 'cancel') {
      toast.error('删除失败')
    }
  }
}

async function handleManualRefresh() {
  await refresh()
}

function getStatusLabel(status: ImageStatus): string {
  return { pending: '等待中', building: '构建中', available: '可用', failed: '失败' }[status]
}

function getStatusColor(status: ImageStatus): string {
  return { pending: '#8b949e', building: '#f59e0b', available: '#10b981', failed: '#ef4444' }[
    status
  ]
}

watch(
  hasActiveImages,
  (active) => {
    if (active) {
      startPolling()
      return
    }
    stopPolling()
  },
  { immediate: true }
)

onMounted(() => {
  void refresh()
})

onUnmounted(() => {
  stopPolling()
})
</script>

<style scoped>
.journal-shell {
  --journal-ink: var(--color-text-primary);
  --journal-muted: var(--color-text-secondary);
  --journal-accent: var(--color-primary);
  --journal-border: color-mix(in srgb, var(--color-border-default) 84%, transparent);
  --journal-surface: color-mix(in srgb, var(--color-bg-surface) 92%, var(--color-bg-base));
  --journal-surface-subtle: color-mix(in srgb, var(--color-bg-surface) 78%, var(--color-bg-base));
  --admin-control-border: color-mix(in srgb, var(--journal-border) 78%, transparent);
  font-family:
    'IBM Plex Sans', 'Noto Sans SC', 'PingFang SC', 'Hiragino Sans GB', 'Microsoft YaHei',
    sans-serif;
}

.journal-hero {
  border-color: var(--journal-border);
  background:
    radial-gradient(
      circle at top right,
      color-mix(in srgb, var(--journal-accent) 7%, transparent),
      transparent 22rem
    ),
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--journal-surface) 96%, var(--color-bg-base)),
      var(--journal-surface)
    );
  box-shadow: 0 22px 50px var(--color-shadow-soft);
}

.journal-eyebrow,
.journal-note-label {
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

.journal-note-value {
  margin-top: 0.35rem;
  font-size: 1rem;
  font-weight: 600;
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
  border-top: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
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
  min-height: 2.35rem;
  padding: 0.5rem 0.85rem;
}

.admin-btn-primary {
  background: var(--journal-accent);
  color: #fff;
  box-shadow: 0 10px 24px color-mix(in srgb, var(--journal-accent) 18%, transparent);
}

.admin-btn-danger {
  border: 1px solid color-mix(in srgb, var(--color-danger) 28%, transparent);
  background: color-mix(in srgb, var(--color-danger) 10%, var(--journal-surface));
  color: color-mix(in srgb, var(--color-danger) 88%, var(--journal-ink));
}

.admin-btn-ghost {
  border: 1px solid var(--journal-border);
  background: color-mix(in srgb, var(--journal-surface) 92%, var(--color-bg-base));
  color: var(--journal-ink);
}

.admin-status-chip {
  display: inline-flex;
  align-items: center;
  border-radius: 0.5rem;
  padding: 0.32rem 0.65rem;
  font-size: 0.72rem;
  font-weight: 600;
}

.admin-empty {
  padding: 1rem 0 0;
  font-size: 0.875rem;
  color: var(--journal-muted);
}

:deep(.el-dialog) {
  border: 1px solid var(--journal-border);
  border-radius: 20px;
  background:
    radial-gradient(
      circle at top right,
      color-mix(in srgb, var(--journal-accent) 8%, transparent),
      transparent 18rem
    ),
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--journal-surface) 96%, var(--color-bg-base)),
      var(--journal-surface)
    );
  box-shadow: 0 24px 60px var(--color-shadow-soft);
}

:deep(.el-dialog__title) {
  color: var(--journal-ink);
}

:deep(.el-form-item__label) {
  color: var(--journal-muted);
}

:deep(.el-input__wrapper),
:deep(.el-textarea__inner) {
  border: 1px solid var(--journal-border);
  background: var(--journal-surface);
  color: var(--journal-ink);
  box-shadow: none;
}

:deep(.el-input__wrapper.is-focus),
:deep(.el-textarea__inner:focus) {
  border-color: color-mix(in srgb, var(--journal-accent) 48%, transparent);
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--journal-accent) 14%, transparent);
}

:deep(.el-input__inner),
:deep(.el-textarea__inner) {
  color: var(--journal-ink);
}

:deep(.el-input__inner::placeholder),
:deep(.el-textarea__inner::placeholder) {
  color: var(--journal-muted);
}

.image-header {
  display: grid;
  gap: 1.5rem;
  padding-bottom: 1.5rem;
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
}

.image-title {
  margin-top: 0.85rem;
  font-size: clamp(1.95rem, 2vw, 2.45rem);
  font-weight: 700;
  line-height: 1.06;
  color: var(--journal-ink);
}

.image-copy {
  margin-top: 0.7rem;
  max-width: 48rem;
  font-size: 0.92rem;
  line-height: 1.7;
  color: var(--journal-muted);
}

.image-header__side {
  display: grid;
  gap: 0.85rem;
  justify-items: start;
}

.image-header__actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
}

.image-summary-grid {
  display: grid;
  gap: 0.85rem;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.image-divider {
  margin: 1.2rem 0;
}

.image-board {
  padding-top: 0.2rem;
}

.image-board__head {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-end;
  justify-content: space-between;
  gap: 0.75rem;
}

.image-section-title {
  margin-top: 0.35rem;
  font-size: 1.15rem;
  font-weight: 700;
  color: var(--journal-ink);
}

.image-board__hint,
.image-row__time {
  font-size: 0.82rem;
  line-height: 1.6;
  color: var(--journal-muted);
}

.image-list {
  --image-list-columns: minmax(10rem, 1fr) minmax(8rem, 0.78fr) minmax(0, 1.5fr)
    minmax(7rem, 0.78fr) minmax(10rem, 0.92fr) auto;
  display: grid;
  gap: 0;
}

.image-directory-head {
  display: grid;
  grid-template-columns: var(--image-list-columns);
  gap: 1rem;
  padding: 1rem 0 0.8rem;
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.16em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.image-directory-head__actions {
  text-align: right;
}

.image-row {
  display: grid;
  gap: 1rem;
  grid-template-columns: var(--image-list-columns);
  align-items: start;
  padding: 1rem 0;
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
}

.image-row__name,
.image-row__tag,
.image-row__description,
.image-row__status,
.image-row__actions {
  min-width: 0;
}

.image-row__name,
.image-row__tag {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, 'Liberation Mono', monospace;
}

.image-row__name {
  font-size: 1rem;
  font-weight: 700;
  color: var(--journal-ink);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.image-row__tag {
  padding-top: 0.1rem;
  color: var(--journal-muted);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.image-row__description {
  display: -webkit-box;
  font-size: 0.88rem;
  line-height: 1.65;
  color: var(--journal-muted);
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.image-row__status {
  display: flex;
  align-items: flex-start;
}

.image-row__time {
  padding-top: 0.15rem;
  font-size: 0.82rem;
  line-height: 1.6;
  color: var(--journal-muted);
}

.image-row__actions {
  display: flex;
  justify-content: flex-end;
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
    radial-gradient(
      circle at top right,
      color-mix(in srgb, var(--journal-accent) 10%, transparent),
      transparent 18rem
    ),
    linear-gradient(180deg, rgba(15, 23, 42, 0.96), rgba(15, 23, 42, 0.9));
}

:global([data-theme='dark']) .admin-btn-ghost,
:global([data-theme='dark']) :deep(.el-input__wrapper),
:global([data-theme='dark']) :deep(.el-textarea__inner) {
  background: color-mix(in srgb, var(--journal-surface) 94%, transparent);
}

:global([data-theme='dark']) .admin-btn-danger {
  background: color-mix(in srgb, var(--color-danger) 12%, var(--journal-surface));
}

@media (max-width: 1040px) {
  .image-directory-head {
    display: none;
  }

  .image-row {
    grid-template-columns: minmax(0, 1fr);
  }

  .image-row__actions {
    align-items: flex-start;
    justify-content: flex-start;
  }
}

@media (max-width: 720px) {
  .image-summary-grid {
    grid-template-columns: minmax(0, 1fr);
  }
}
</style>
