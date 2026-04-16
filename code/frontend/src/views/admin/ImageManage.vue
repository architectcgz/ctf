<template>
  <section
    class="journal-shell journal-shell-admin journal-notes-rail journal-hero flex min-h-full flex-1 flex-col rounded-[30px] border px-6 py-6 md:px-8"
  >
    <header class="image-header">
      <div class="image-header__intro">
        <div class="workspace-overline">Image Registry</div>
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
        <div class="image-status-strip" aria-label="镜像状态摘要">
          <div class="image-status-strip__row">
            <div
              v-for="item in statusSummary"
              :key="item.key"
              :class="['image-status-pill', `image-status-pill--${item.tone}`]"
              data-testid="image-status-pill"
            >
              <span>{{ item.label }}</span>
              <strong>{{ item.value }}</strong>
            </div>
          </div>
          <div class="image-status-strip__note">{{ refreshHint }}</div>
        </div>
      </div>
    </header>

    <section class="image-board workspace-directory-section">
      <div class="image-board__head">
        <div>
          <div class="journal-note-label">Images</div>
          <h2 class="image-section-title">镜像列表</h2>
        </div>
        <div class="image-board__hint">按创建时间倒序</div>
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
              <span class="admin-status-chip" :style="getStatusStyle(row.status)">
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

    <AdminSurfaceModal
      :open="dialogVisible"
      title="创建镜像"
      subtitle="填写镜像名称、标签和说明，提交后会进入镜像目录并参与构建状态跟踪。"
      eyebrow="Image Registry"
      width="31.25rem"
      @close="dialogVisible = false"
      @update:open="dialogVisible = $event"
    >
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
    </AdminSurfaceModal>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, reactive, ref, watch } from 'vue'
import { getImages, createImage, deleteImage } from '@/api/admin'
import AdminPaginationControls from '@/components/admin/AdminPaginationControls.vue'
import AdminSurfaceModal from '@/components/common/modal-templates/AdminSurfaceModal.vue'
import { confirmDestructiveAction } from '@/composables/useDestructiveConfirm'
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

const statusSummary = computed(() => {
  const counts = {
    available: 0,
    pending: 0,
    building: 0,
    failed: 0,
  }

  for (const row of list.value) {
    counts[row.status] += 1
  }

  const summary = []

  if (counts.available > 0) {
    summary.push({ key: 'available', label: '可用', value: counts.available, tone: 'success' })
  }
  if (counts.building > 0) {
    summary.push({ key: 'building', label: '构建中', value: counts.building, tone: 'warning' })
  }
  if (counts.pending > 0) {
    summary.push({ key: 'pending', label: '等待中', value: counts.pending, tone: 'muted' })
  }
  if (counts.failed > 0) {
    summary.push({ key: 'failed', label: '失败', value: counts.failed, tone: 'danger' })
  }

  if (summary.length > 0) {
    return summary
  }

  return [{ key: 'empty', label: '当前页', value: 0, tone: 'muted' as const }]
})

const imageStatusMeta: Record<
  ImageStatus,
  { label: string; color: string; backgroundColor: string }
> = {
  pending: {
    label: '等待中',
    color: 'color-mix(in srgb, var(--journal-muted) 84%, var(--journal-ink))',
    backgroundColor: 'color-mix(in srgb, var(--journal-muted) 14%, transparent)',
  },
  building: {
    label: '构建中',
    color: 'var(--color-warning)',
    backgroundColor: 'color-mix(in srgb, var(--color-warning) 14%, transparent)',
  },
  available: {
    label: '可用',
    color: 'var(--color-success)',
    backgroundColor: 'color-mix(in srgb, var(--color-success) 14%, transparent)',
  },
  failed: {
    label: '失败',
    color: 'var(--color-danger)',
    backgroundColor: 'color-mix(in srgb, var(--color-danger) 14%, transparent)',
  },
}

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
  const confirmed = await confirmDestructiveAction({
    message: '确定要删除此镜像吗？',
  })
  if (!confirmed) {
    return
  }

  try {
    await deleteImage(id)
    toast.success('删除成功')
    refresh()
  } catch (error) {
    const message = error instanceof Error && error.message.trim() ? error.message : '删除失败'
    toast.error(message)
  }
}

async function handleManualRefresh() {
  await refresh()
}

function getStatusLabel(status: ImageStatus): string {
  return imageStatusMeta[status].label
}

function getStatusStyle(status: ImageStatus): Record<string, string> {
  const meta = imageStatusMeta[status]
  return {
    backgroundColor: meta.backgroundColor,
    color: meta.color,
  }
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
  --admin-summary-grid-columns: repeat(2, minmax(0, 1fr));
  --admin-control-border: color-mix(in srgb, var(--journal-border) 78%, transparent);
  --journal-divider-border: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
  --journal-shell-hero-radial-strength: 7%;
  --journal-shell-hero-radial-size: 22rem;
  --journal-shell-hero-end: var(--journal-surface);
  --journal-shell-hero-shadow: 0 22px 50px var(--color-shadow-soft);
  --journal-shell-dark-ink: var(--color-text-primary);
  --journal-shell-dark-accent: var(--color-primary-hover);
  --journal-shell-dark-surface: color-mix(
    in srgb,
    var(--color-bg-surface) 92%,
    var(--color-bg-base)
  );
  --journal-shell-dark-surface-subtle: color-mix(
    in srgb,
    var(--color-bg-surface) 78%,
    var(--color-bg-base)
  );
  --journal-shell-dark-hero-radial-strength: 10%;
  --journal-shell-dark-hero-top: color-mix(
    in srgb,
    var(--journal-surface) 97%,
    var(--color-bg-base)
  );
  --journal-shell-dark-hero-end: color-mix(
    in srgb,
    var(--journal-surface-subtle) 95%,
    var(--color-bg-base)
  );
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
  min-height: 2.35rem;
  padding: var(--space-2) var(--space-3);
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
  padding: var(--space-1) var(--space-2-5);
  font-size: var(--font-size-0-72);
  font-weight: 600;
}

.admin-empty {
  padding: var(--space-4) 0 0;
  font-size: var(--font-size-0-875);
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
  gap: var(--space-6);
  padding-bottom: var(--space-6);
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
}

.image-copy {
  max-width: 48rem;
}

.image-header__side {
  display: grid;
  gap: var(--space-3);
  justify-items: start;
}

.image-header__actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-3);
}

.image-status-strip {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3) var(--space-4);
}

.image-status-strip__row {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2-5);
}

.image-status-strip__note {
  font-size: var(--font-size-0-82);
  line-height: 1.6;
  color: var(--journal-muted);
}

.image-status-pill {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
  min-height: 2.25rem;
  padding: 0 var(--space-3);
  border: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
  border-radius: 999px;
  background: color-mix(in srgb, var(--journal-surface) 92%, var(--color-bg-base));
  color: var(--journal-muted);
  font-size: var(--font-size-0-82);
  line-height: 1;
}

.image-status-pill strong {
  font-size: var(--font-size-0-9);
  font-weight: 700;
  color: var(--journal-ink);
}

.image-status-pill--success {
  border-color: color-mix(in srgb, var(--color-success) 22%, transparent);
  background: color-mix(in srgb, var(--color-success) 12%, var(--journal-surface));
  color: color-mix(in srgb, var(--color-success) 82%, var(--journal-ink));
}

.image-status-pill--warning {
  border-color: color-mix(in srgb, var(--color-warning) 24%, transparent);
  background: color-mix(in srgb, var(--color-warning) 12%, var(--journal-surface));
  color: color-mix(in srgb, var(--color-warning) 84%, var(--journal-ink));
}

.image-status-pill--danger {
  border-color: color-mix(in srgb, var(--color-danger) 24%, transparent);
  background: color-mix(in srgb, var(--color-danger) 10%, var(--journal-surface));
  color: color-mix(in srgb, var(--color-danger) 84%, var(--journal-ink));
}

.image-status-pill--muted {
  border-color: color-mix(in srgb, var(--journal-muted) 18%, transparent);
  background: color-mix(in srgb, var(--journal-muted) 10%, var(--journal-surface));
  color: var(--journal-muted);
}

.image-board {
  padding-top: var(--space-1);
}

.image-board__head {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-end;
  justify-content: space-between;
  gap: var(--space-3);
}

.image-section-title {
  margin-top: var(--space-1-5);
  font-size: var(--font-size-1-15);
  font-weight: 700;
  color: var(--journal-ink);
}

.image-board__hint,
.image-row__time {
  font-size: var(--font-size-0-82);
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
  gap: var(--space-4);
  padding: var(--space-4) 0 var(--space-3);
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
  font-size: var(--font-size-0-72);
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
  gap: var(--space-4);
  grid-template-columns: var(--image-list-columns);
  align-items: start;
  padding: var(--space-4) 0;
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
  font-family: var(--font-family-mono);
}

.image-row__name {
  font-size: var(--font-size-1-00);
  font-weight: 700;
  color: var(--journal-ink);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.image-row__tag {
  padding-top: var(--space-0-5);
  color: var(--journal-muted);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.image-row__description {
  display: -webkit-box;
  font-size: var(--font-size-0-88);
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
  padding-top: var(--space-0-5);
  font-size: var(--font-size-0-82);
  line-height: 1.6;
  color: var(--journal-muted);
}

.image-row__actions {
  display: flex;
  justify-content: flex-end;
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
  .image-status-strip {
    align-items: flex-start;
  }

  .image-status-strip__note {
    width: 100%;
  }
}
</style>
