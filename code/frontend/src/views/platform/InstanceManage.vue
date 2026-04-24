<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import {
  Server,
  Activity,
  AlertTriangle,
  Trash2,
  RefreshCw,
} from 'lucide-vue-next'

import type { TeacherInstanceItem } from '@/api/contracts'
import { destroyTeacherInstance, getTeacherInstances } from '@/api/teacher'
import WorkspaceDataTable from '@/components/common/WorkspaceDataTable.vue'
import WorkspaceDirectoryPagination from '@/components/common/WorkspaceDirectoryPagination.vue'
import AppLoading from '@/components/common/AppLoading.vue'
import AppEmpty from '@/components/common/AppEmpty.vue'
import { confirmDestructiveAction } from '@/composables/useDestructiveConfirm'

const router = useRouter()
const list = ref<TeacherInstanceItem[]>([])
const page = ref(1)
const pageSize = ref(15)
const loading = ref(false)
const destroyingId = ref('')
const error = ref<string | null>(null)

const total = computed(() => list.value.length)
const totalPages = computed(() => Math.max(1, Math.ceil(total.value / pageSize.value)))
const pageRows = computed(() => {
  const start = (page.value - 1) * pageSize.value
  return list.value.slice(start, start + pageSize.value).map((item) => ({
    id: item.id,
    challenge: item.challenge_title,
    user: item.student_name || item.student_username,
    user_meta: `${item.student_username} · ${item.class_name}`,
    ip_address: item.access_url || '暂未分配',
    status: item.status,
    status_label: formatStatus(item.status),
    created_at: formatDateTime(item.created_at),
    actions: '销毁',
  }))
})
const runningCount = computed(() => list.value.filter((item) => item.status === 'running').length)
const warningCount = computed(
  () => list.value.filter((item) => item.status !== 'running' || item.remaining_time <= 600).length
)

function formatStatus(status: string): string {
  switch (status) {
    case 'running':
      return '运行中'
    case 'creating':
      return '创建中'
    case 'expired':
      return '已过期'
    case 'failed':
      return '异常'
    default:
      return status
  }
}

function formatDateTime(value: string): string {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return '--'
  }

  return new Intl.DateTimeFormat('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  }).format(date)
}

async function loadInstances(): Promise<void> {
  loading.value = true
  error.value = null
  try {
    list.value = await getTeacherInstances({
      class_name: undefined,
      keyword: undefined,
      student_no: undefined,
    })
    if (page.value > totalPages.value) {
      page.value = totalPages.value
    }
  } catch (err) {
    console.error('加载实例列表失败:', err)
    error.value = '加载实例列表失败，请稍后重试'
    list.value = []
  } finally {
    loading.value = false
  }
}

async function handleDestroyInstance(instance: TeacherInstanceItem): Promise<void> {
  const confirmed = await confirmDestructiveAction({
    title: '强制销毁实例',
    message: `您确定要强制销毁实例 ${instance.id} 吗？此操作不可逆，用户当前的运行状态将丢失。`,
    confirmButtonText: '强制销毁',
    cancelButtonText: '取消',
  })

  if (!confirmed) return

  try {
    destroyingId.value = instance.id
    await destroyTeacherInstance(instance.id)
    list.value = list.value.filter((item) => item.id !== instance.id)
    if (page.value > totalPages.value) {
      page.value = totalPages.value
    }
  } catch (err) {
    console.error('销毁实例失败:', err)
    error.value = '销毁实例失败，请稍后重试'
  } finally {
    destroyingId.value = ''
  }
}

function requestDestroyById(id: string): void {
  const instance = list.value.find((item) => item.id === id)
  if (!instance) {
    return
  }

  void handleDestroyInstance(instance)
}

function handlePageChange(p: number): void {
  page.value = p
}

onMounted(() => {
  void loadInstances()
})

const columns = [
  { key: 'id', label: '实例 ID', widthClass: 'w-[20%] min-w-[12rem]' },
  { key: 'challenge', label: '关联题目', widthClass: 'w-[15%] min-w-[10rem]' },
  { key: 'user', label: '所属用户', widthClass: 'w-[15%] min-w-[10rem]' },
  { key: 'ip_address', label: '访问地址', widthClass: 'w-[15%] min-w-[10rem]' },
  { key: 'status', label: '状态', widthClass: 'w-[10%] min-w-[6rem]', align: 'center' as const },
  { key: 'created_at', label: '创建时间', widthClass: 'w-[15%] min-w-[10rem]' },
  { key: 'actions', label: '操作', widthClass: 'w-[8rem]', align: 'right' as const },
]
</script>

<template>
  <div class="workspace-shell">
    <div class="workspace-grid">
      <main class="content-pane">
        <section class="workspace-hero">
          <div class="workspace-tab-heading__main">
            <div class="workspace-overline">
              Instance Workspace
            </div>
            <h1 class="hero-title">
              实例管理
            </h1>
            <p class="hero-summary">
              在后台视角查看实例状态、到期节奏与访问地址，并快速销毁异常环境。
            </p>
          </div>

          <div class="awd-library-hero-actions">
            <div class="quick-actions">
              <button
                type="button"
                class="ui-btn ui-btn--ghost"
                @click="router.push({ name: 'PlatformOverview' })"
              >
                返回概览
              </button>
              <button
                type="button"
                class="ui-btn ui-btn--primary"
                @click="loadInstances"
              >
                <RefreshCw class="h-4 w-4" />
                刷新列表
              </button>
            </div>
          </div>
        </section>

        <div class="admin-instance-manage-shell__content">
          <div class="admin-summary-grid admin-instance-manage-shell__summary progress-strip metric-panel-grid metric-panel-default-surface metric-panel-workspace-surface">
            <article class="journal-note progress-card metric-panel-card">
              <div class="journal-note-label progress-card-label metric-panel-label">
                <span>运行中</span>
                <Activity class="h-4 w-4" />
              </div>
              <div class="journal-note-value progress-card-value metric-panel-value">
                {{ runningCount.toString().padStart(2, '0') }}
              </div>
              <div class="journal-note-helper progress-card-hint metric-panel-helper">
                当前活跃实例
              </div>
            </article>

            <article class="journal-note progress-card metric-panel-card">
              <div class="journal-note-label progress-card-label metric-panel-label">
                <span>总实例数</span>
                <Server class="h-4 w-4" />
              </div>
              <div class="journal-note-value progress-card-value metric-panel-value">
                {{ total.toString().padStart(2, '0') }}
              </div>
              <div class="journal-note-helper progress-card-hint metric-panel-helper">
                系统托管总计
              </div>
            </article>

            <article class="journal-note progress-card metric-panel-card">
              <div class="journal-note-label progress-card-label metric-panel-label">
                <span>预警项</span>
                <AlertTriangle class="h-4 w-4" />
              </div>
              <div class="journal-note-value progress-card-value metric-panel-value">
                {{ warningCount.toString().padStart(2, '0') }}
              </div>
              <div class="journal-note-helper progress-card-hint metric-panel-helper">
                即将过期或异常
              </div>
            </article>
          </div>

          <section class="workspace-directory-section admin-instance-manage-directory">
            <header class="list-heading">
              <div>
                <div class="workspace-overline">
                  Active Instances
                </div>
                <h2 class="list-heading__title">
                  实时实例列表
                </h2>
              </div>
            </header>

            <div
              v-if="loading && list.length === 0"
              class="py-12 flex justify-center"
            >
              <AppLoading>同步实例状态...</AppLoading>
            </div>

            <template v-else>
              <AppEmpty
                v-if="list.length === 0"
                class="workspace-directory-empty"
                icon="Server"
                title="暂无运行中的实例"
                description="当前平台上没有任何用户开启题目环境。"
              />

              <WorkspaceDataTable
                v-else
                class="workspace-directory-list admin-instance-manage-table"
                :columns="columns"
                :rows="pageRows"
                row-key="id"
              >
                <template #cell-id="{ row }">
                  <span class="font-mono text-xs">{{ (row as { id: string }).id }}</span>
                </template>
                <template #cell-user="{ row }">
                  <div class="flex flex-col items-start gap-1">
                    <span>{{ (row as { user: string }).user }}</span>
                    <span class="font-mono text-[11px] text-[var(--journal-muted)]">
                      {{ (row as { user_meta: string }).user_meta }}
                    </span>
                  </div>
                </template>
                <template #cell-status="{ row }">
                  <span
                    class="instance-status-pill"
                    :class="(row as { status: string }).status === 'running' ? 'instance-status-pill--running' : 'instance-status-pill--inactive'"
                  >
                    {{ (row as { status_label: string }).status_label }}
                  </span>
                </template>
                <template #cell-actions="{ row }">
                  <div class="flex justify-end gap-2">
                    <button
                      type="button"
                      class="ui-btn ui-btn--ghost ui-btn--xs"
                      :disabled="destroyingId === (row as { id: string }).id"
                      @click="requestDestroyById((row as { id: string }).id)"
                    >
                      <Trash2 class="h-3 w-3 mr-1" />
                      {{ destroyingId === (row as { id: string }).id ? '销毁中' : '销毁' }}
                    </button>
                  </div>
                </template>
              </WorkspaceDataTable>

              <div class="workspace-directory-pagination">
                <WorkspaceDirectoryPagination
                  :page="page"
                  :total-pages="totalPages"
                  :total="total"
                  total-label="个实例"
                  @change-page="handlePageChange"
                />
              </div>
            </template>
          </section>

          <div
            v-if="error"
            class="teacher-surface-error"
          >
            {{ error }}
          </div>
        </div>
      </main>
    </div>
  </div>
</template>

<style scoped>
.admin-instance-manage-shell__content {
  display: flex;
  flex-direction: column;
  gap: var(--workspace-directory-page-block-gap);
  margin-top: var(--space-10);
}

.workspace-hero {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: var(--space-7);
  padding-bottom: var(--space-6);
  border-bottom: 1px solid var(--workspace-line-soft);
}

.hero-title {
  margin: 0.5rem 0 0;
  font-size: var(--workspace-page-title-font-size);
  line-height: var(--workspace-page-title-line-height);
  letter-spacing: var(--workspace-page-title-letter-spacing);
  color: var(--journal-ink);
}

.hero-summary {
  max-width: 760px;
  margin-top: var(--space-3-5);
  font-size: var(--font-size-15);
  line-height: 1.9;
  color: var(--journal-muted);
}

.awd-library-hero-actions {
  display: flex;
  align-items: flex-end;
  padding-bottom: 0.5rem;
}

.quick-actions {
  display: flex;
  gap: 0.75rem;
}

.instance-status-pill {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 1.4rem;
  padding: 0 0.5rem;
  border-radius: 999px;
  border: 1px solid transparent;
  font-size: var(--font-size-10);
  font-weight: 700;
  text-transform: uppercase;
}

.instance-status-pill--running {
  background: color-mix(in srgb, var(--color-success) 10%, transparent);
  border-color: color-mix(in srgb, var(--color-success) 24%, transparent);
  color: color-mix(in srgb, var(--color-success) 82%, var(--color-text-primary));
}

.instance-status-pill--inactive {
  background: color-mix(in srgb, var(--color-text-muted) 10%, transparent);
  border-color: color-mix(in srgb, var(--color-border-default) 92%, transparent);
  color: var(--color-text-secondary);
}
</style>
