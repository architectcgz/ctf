<template>
  <section
    class="journal-shell journal-hero flex min-h-full flex-1 flex-col rounded-[30px] border px-6 py-6 md:px-8"
  >
    <div class="grid gap-6 xl:grid-cols-[1.06fr_0.94fr]">
      <div>
        <div class="journal-eyebrow">Image Library</div>
        <h1
          class="mt-3 text-3xl font-semibold tracking-tight text-[var(--journal-ink)] md:text-[2.45rem]"
        >
          镜像管理
        </h1>
        <p class="mt-3 max-w-2xl text-sm leading-7 text-[var(--journal-muted)]">
          在这里查看镜像状态，并继续创建或清理镜像资源。
        </p>

        <div class="mt-6 flex flex-wrap gap-3">
          <button class="admin-btn admin-btn-primary" @click="dialogVisible = true">
            创建镜像
          </button>
        </div>
      </div>

      <article class="journal-brief rounded-[24px] border px-5 py-5">
        <div class="journal-note-label">镜像概况</div>
        <div class="mt-5 grid gap-3 sm:grid-cols-2">
          <div class="journal-note">
            <div class="journal-note-label">镜像总量</div>
            <div class="journal-note-value">{{ total }}</div>
            <div class="journal-note-helper">当前库中的镜像总数</div>
          </div>
          <div class="journal-note">
            <div class="journal-note-label">当前页</div>
            <div class="journal-note-value">{{ list.length }}</div>
            <div class="journal-note-helper">当前分页内的镜像数量</div>
          </div>
        </div>
      </article>
    </div>
    <div class="journal-divider" />

    <div class="space-y-3">
      <div v-if="loading" class="flex items-center justify-center py-12">
        <div
          class="h-8 w-8 animate-spin rounded-full border-4 border-[var(--journal-border)] border-t-[var(--journal-accent)]"
        ></div>
      </div>

      <template v-else>
        <div v-if="list.length === 0" class="admin-empty">当前还没有镜像。</div>

        <div v-else class="space-y-3">
          <article v-for="row in list" :key="row.id" class="image-row">
            <div class="flex flex-wrap items-start justify-between gap-4">
              <div class="min-w-0">
                <div class="flex flex-wrap items-center gap-2">
                  <h2 class="font-mono text-base font-semibold text-[var(--journal-ink)]">
                    {{ row.name }}
                  </h2>
                  <span class="font-mono text-sm text-[var(--journal-muted)]">:{{ row.tag }}</span>
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
                <p class="mt-2 text-sm text-[var(--journal-muted)]">
                  {{ row.description || '未填写镜像说明' }}
                </p>
              </div>
              <div class="text-right text-sm text-[var(--journal-muted)]">
                {{ new Date(row.created_at).toLocaleString() }}
              </div>
            </div>

            <div class="journal-divider mt-4" />

            <div class="mt-4 flex justify-end">
              <button
                class="admin-btn admin-btn-danger admin-btn-compact"
                @click="handleDelete(row.id)"
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
    </div>

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
  --journal-accent: #2563eb;
  --journal-border: color-mix(in srgb, var(--color-border-default) 82%, transparent);
  --journal-surface: color-mix(in srgb, var(--color-bg-surface) 88%, var(--color-bg-base));
  --journal-surface-subtle: color-mix(in srgb, var(--color-bg-surface) 74%, var(--color-bg-base));
}

.journal-hero,
.journal-panel {
  border-color: var(--journal-border);
  background:
    radial-gradient(circle at top right, rgba(37, 99, 235, 0.08), transparent 18rem),
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
    radial-gradient(circle at top right, rgba(37, 99, 235, 0.08), transparent 18rem),
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
  --journal-accent: #60a5fa;
  --journal-border: color-mix(in srgb, var(--color-border-default) 82%, transparent);
  --journal-surface: color-mix(in srgb, var(--color-bg-surface) 88%, var(--color-bg-base));
  --journal-surface-subtle: color-mix(in srgb, var(--color-bg-surface) 74%, var(--color-bg-base));
}

:global([data-theme='dark']) .journal-hero,
:global([data-theme='dark']) .journal-panel {
  background:
    radial-gradient(circle at top right, rgba(96, 165, 250, 0.1), transparent 18rem),
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
</style>
