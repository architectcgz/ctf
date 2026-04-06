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
        <button class="admin-btn admin-btn-primary" @click="dialogVisible = true">创建镜像</button>
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

    <section class="image-board">
      <div class="image-board__head">
        <div>
          <div class="journal-note-label">Images</div>
          <h2 class="image-section-title">镜像列表</h2>
        </div>
        <div class="image-board__hint">每 10 秒自动刷新一次状态</div>
      </div>

      <div v-if="loading" class="flex items-center justify-center py-12">
        <div
          class="h-8 w-8 animate-spin rounded-full border-4 border-[var(--journal-border)] border-t-[var(--journal-accent)]"
        ></div>
      </div>

      <template v-else>
        <div v-if="list.length === 0" class="admin-empty">当前还没有镜像。</div>

        <div v-else class="image-list">
          <article v-for="row in list" :key="row.id" class="image-row">
            <div class="image-row__main">
              <div class="image-row__titleline">
                <h2 class="image-row__title">{{ row.name }}</h2>
                <span class="image-row__tag">:{{ row.tag }}</span>
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
              <p class="image-row__description">{{ row.description || '未填写镜像说明' }}</p>
            </div>

            <div class="image-row__aside">
              <div class="image-row__time">{{ new Date(row.created_at).toLocaleString() }}</div>
              <button
                class="admin-btn admin-btn-danger admin-btn-compact"
                @click="handleDelete(row.id)"
              >
                删除
              </button>
            </div>
          </article>
        </div>

        <div v-if="total > 0" class="admin-pagination">
          <span>共 {{ total }} 条</span>
          <div class="flex items-center gap-2">
            <button
              :disabled="page === 1"
              class="admin-btn admin-btn-ghost admin-btn-compact disabled:cursor-not-allowed disabled:opacity-50"
              @click="changePage(page - 1)"
            >
              上一页
            </button>
            <span>{{ page }} / {{ Math.ceil(total / pageSize) }}</span>
            <button
              :disabled="page >= Math.ceil(total / pageSize)"
              class="admin-btn admin-btn-ghost admin-btn-compact disabled:cursor-not-allowed disabled:opacity-50"
              @click="changePage(page + 1)"
            >
              下一页
            </button>
          </div>
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
import { onMounted, onUnmounted, reactive, ref } from 'vue'
import { ElMessageBox } from 'element-plus'
import { getImages, createImage, deleteImage } from '@/api/admin'
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

function getStatusLabel(status: ImageStatus): string {
  return { pending: '等待中', building: '构建中', available: '可用', failed: '失败' }[status]
}

function getStatusColor(status: ImageStatus): string {
  return { pending: '#8b949e', building: '#f59e0b', available: '#10b981', failed: '#ef4444' }[
    status
  ]
}

onMounted(() => {
  refresh()
  pollTimer = window.setInterval(refresh, 10000)
})

onUnmounted(() => {
  if (pollTimer) clearInterval(pollTimer)
})
</script>

<style scoped>
.journal-shell {
  --journal-ink: var(--color-text-primary);
  --journal-muted: var(--color-text-secondary);
  --journal-accent: var(--color-primary);
  --journal-border: color-mix(in srgb, var(--color-border-default) 82%, transparent);
  --journal-surface: color-mix(in srgb, var(--color-bg-surface) 88%, var(--color-bg-base));
  --journal-surface-subtle: color-mix(in srgb, var(--color-bg-surface) 74%, var(--color-bg-base));
}

.journal-hero,
.journal-panel {
  border-color: var(--journal-border);
  background:
    radial-gradient(circle at top right, color-mix(in srgb, var(--journal-accent) 8%, transparent), transparent 18rem),
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--journal-surface, var(--color-bg-surface)) 96%, var(--color-bg-base)),
      color-mix(
        in srgb,
        var(--journal-surface-subtle, var(--color-bg-elevated)) 94%,
        var(--color-bg-base)
      )
    );
  border-radius: 16px !important;
  box-shadow: 0 18px 40px var(--color-shadow-soft);
}

.journal-brief {
  background: var(--journal-surface-subtle);
  border-color: var(--journal-border);
  border-radius: 16px !important;
  box-shadow: 0 8px 18px var(--color-shadow-soft);
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
  border-radius: 14px;
  border: 1px solid var(--journal-border);
  background: var(--journal-surface);
  padding: 0.75rem 0.875rem;
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
  border-top: 1px dashed
    color-mix(in srgb, var(--journal-border, var(--color-border-default)) 88%, transparent);
}

.image-row {
  border: 1px solid var(--journal-border);
  border-radius: 18px;
  background: color-mix(
    in srgb,
    var(--journal-surface, var(--color-bg-surface)) 92%,
    var(--color-bg-base)
  );
  padding: 1rem;
  box-shadow: 0 10px 24px var(--color-shadow-soft);
}

.admin-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  min-height: 2.75rem;
  border-radius: 1rem;
  padding: 0.65rem 1rem;
  font-size: 0.875rem;
  font-weight: 600;
  transition: all 150ms ease;
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
  background: color-mix(
    in srgb,
    var(--journal-surface, var(--color-bg-surface)) 92%,
    var(--color-bg-base)
  );
  color: var(--journal-ink);
}

.admin-status-chip {
  display: inline-flex;
  align-items: center;
  border-radius: 999px;
  padding: 0.34rem 0.75rem;
  font-size: 0.72rem;
  font-weight: 600;
}

.admin-empty {
  border: 1px dashed
    color-mix(in srgb, var(--journal-border, var(--color-border-default)) 88%, transparent);
  border-radius: 16px;
  background: color-mix(in srgb, var(--journal-surface-subtle) 88%, var(--color-bg-base));
  padding: 1rem;
  font-size: 0.875rem;
  color: var(--journal-muted);
}

.admin-pagination {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  border-top: 1px dashed
    color-mix(in srgb, var(--journal-border, var(--color-border-default)) 88%, transparent);
  padding-top: 1rem;
  font-size: 0.875rem;
  color: var(--journal-muted);
}

:deep(.el-dialog) {
  border: 1px solid var(--journal-border);
  border-radius: 20px;
  background:
    radial-gradient(circle at top right, color-mix(in srgb, var(--journal-accent) 8%, transparent), transparent 18rem),
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--journal-surface) 96%, var(--color-bg-base)),
      color-mix(in srgb, var(--journal-surface-subtle) 94%, var(--color-bg-base))
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

:global([data-theme='dark']) .journal-shell {
  --journal-ink: var(--color-text-primary);
  --journal-muted: var(--color-text-secondary);
  --journal-accent: var(--color-primary-hover);
  --journal-border: color-mix(in srgb, var(--color-border-default) 82%, transparent);
  --journal-surface: color-mix(in srgb, var(--color-bg-surface) 88%, var(--color-bg-base));
  --journal-surface-subtle: color-mix(in srgb, var(--color-bg-surface) 74%, var(--color-bg-base));
}

:global([data-theme='dark']) .journal-hero,
:global([data-theme='dark']) .journal-panel {
  background:
    radial-gradient(circle at top right, color-mix(in srgb, var(--journal-accent) 10%, transparent), transparent 18rem),
    linear-gradient(180deg, rgba(15, 23, 42, 0.96), rgba(15, 23, 42, 0.9));
}

:global([data-theme='dark']) .image-row,
:global([data-theme='dark']) .admin-btn-ghost,
:global([data-theme='dark']) :deep(.el-input__wrapper),
:global([data-theme='dark']) :deep(.el-textarea__inner) {
  background: color-mix(in srgb, var(--journal-surface) 94%, transparent);
}

:global([data-theme='dark']) .admin-btn-danger {
  background: color-mix(in srgb, var(--color-danger) 12%, var(--journal-surface));
}

.journal-shell {
  --journal-surface: color-mix(in srgb, var(--color-bg-surface) 90%, var(--color-bg-base));
  --journal-surface-subtle: color-mix(in srgb, var(--color-bg-surface) 76%, var(--color-bg-base));
  --admin-control-border: color-mix(in srgb, var(--journal-border) 78%, transparent);
}

.journal-hero {
  background:
    radial-gradient(circle at top right, color-mix(in srgb, var(--journal-accent) 8%, transparent), transparent 16rem),
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--journal-surface) 97%, var(--color-bg-base)),
      color-mix(in srgb, var(--journal-surface-subtle) 95%, var(--color-bg-base))
    );
}

.image-header {
  display: grid;
  gap: 1rem;
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

.image-summary-grid {
  display: grid;
  gap: 0.85rem;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.image-divider {
  margin: 1.2rem 0;
}

.image-board {
  border: 1px solid var(--journal-border);
  border-radius: 22px;
  background: color-mix(in srgb, var(--journal-surface-subtle) 90%, var(--color-bg-base));
  padding: 1.15rem;
  box-shadow: 0 12px 28px var(--color-shadow-soft);
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
  margin-top: 1rem;
  display: grid;
  gap: 0.85rem;
}

.image-row {
  display: grid;
  gap: 1rem;
  grid-template-columns: minmax(0, 1fr) auto;
  border-radius: 20px;
  background: color-mix(in srgb, var(--journal-surface) 94%, var(--color-bg-base));
}

.image-row__titleline {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.55rem;
}

.image-row__title,
.image-row__tag {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, 'Liberation Mono', monospace;
}

.image-row__title {
  font-size: 1rem;
  font-weight: 700;
  color: var(--journal-ink);
}

.image-row__tag {
  color: var(--journal-muted);
}

.image-row__description {
  margin-top: 0.65rem;
  font-size: 0.88rem;
  line-height: 1.65;
  color: var(--journal-muted);
}

.image-row__aside {
  display: flex;
  min-width: 9rem;
  flex-direction: column;
  align-items: flex-end;
  justify-content: space-between;
  gap: 0.75rem;
}

.admin-btn {
  min-height: 2.65rem;
  border-radius: 999px;
  border: 1px solid transparent;
  padding: 0.62rem 1rem;
  transition:
    border-color 150ms ease,
    background 150ms ease,
    color 150ms ease;
}

.admin-btn-compact {
  min-height: 2.2rem;
  padding: 0.46rem 0.82rem;
}

.admin-btn-primary {
  border-color: color-mix(in srgb, var(--journal-accent) 24%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 12%, var(--journal-surface));
  color: color-mix(in srgb, var(--journal-accent) 86%, var(--journal-ink));
  box-shadow: none;
}

.admin-btn-ghost {
  border-color: var(--admin-control-border);
  background: color-mix(in srgb, var(--journal-surface) 95%, var(--color-bg-base));
}

.admin-empty,
.admin-pagination {
  margin-top: 1rem;
}

@media (max-width: 1040px) {
  .image-row {
    grid-template-columns: minmax(0, 1fr);
  }

  .image-row__aside {
    min-width: 0;
    align-items: flex-start;
  }
}

@media (max-width: 720px) {
  .image-summary-grid {
    grid-template-columns: minmax(0, 1fr);
  }
}
</style>
