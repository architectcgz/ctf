<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import {
  Server,
  Activity,
  AlertTriangle,
  ExternalLink,
  Trash2,
  RefreshCw,
} from 'lucide-vue-next'

import type { AdminInstanceListItem } from '@/api/contracts'
import { getAdminInstances, deleteAdminInstance } from '@/api/admin'
import WorkspaceDataTable from '@/components/common/WorkspaceDataTable.vue'
import WorkspaceDirectoryPagination from '@/components/common/WorkspaceDirectoryPagination.vue'
import AppLoading from '@/components/common/AppLoading.vue'
import AppEmpty from '@/components/common/AppEmpty.vue'
import { useAdminDestructiveConfirm } from '@/composables/useAdminDestructiveConfirm'

const router = useRouter()
const list = ref<AdminInstanceListItem[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(15)
const loading = ref(false)

const { confirmDestruction } = useAdminDestructiveConfirm()

async function loadInstances(): Promise<void> {
  loading.value = true
  try {
    const data = await getAdminInstances({
      page: page.value,
      pageSize: pageSize.value,
    })
    list.value = data.items
    total.value = data.total
  } catch (err) {
    console.error('加载实例列表失败:', err)
  } finally {
    loading.value = false
  }
}

async function handleDestroyInstance(instance: AdminInstanceListItem): Promise<void> {
  const confirmed = await confirmDestruction({
    title: '强制销毁实例',
    message: `您确定要强制销毁实例 ${instance.id} 吗？此操作不可逆，用户当前的运行状态将丢失。`,
    confirmButtonText: '强制销毁',
  })

  if (!confirmed) return

  try {
    await deleteAdminInstance(instance.id)
    await loadInstances()
  } catch (err) {
    console.error('销毁实例失败:', err)
  }
}

function handlePageChange(p: number): void {
  page.value = p
  void loadInstances()
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

        <div class="instance-manage-body mt-10 space-y-10">
          <div class="metric-panel-grid metric-panel-grid--premium cols-3">
            <article class="metric-panel-card metric-panel-card--premium">
              <div class="metric-panel-label">
                <span>运行中</span>
                <Activity class="h-4 w-4" />
              </div>
              <div class="metric-panel-value">
                {{ list.filter(i => i.status === 'running').length.toString().padStart(2, '0') }}
              </div>
              <div class="metric-panel-helper">
                当前活跃实例
              </div>
            </article>

            <article class="metric-panel-card metric-panel-card--premium">
              <div class="metric-panel-label">
                <span>总实例数</span>
                <Server class="h-4 w-4" />
              </div>
              <div class="metric-panel-value">
                {{ total.toString().padStart(2, '0') }}
              </div>
              <div class="metric-panel-helper">
                系统托管总计
              </div>
            </article>

            <article class="metric-panel-card metric-panel-card--premium">
              <div class="metric-panel-label">
                <span>预警项</span>
                <AlertTriangle class="h-4 w-4" />
              </div>
              <div class="metric-panel-value">
                00
              </div>
              <div class="metric-panel-helper">
                即将过期或异常
              </div>
            </article>
          </div>

          <section class="workspace-directory-section">
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
                class="workspace-directory-list"
                :columns="columns"
                :rows="list"
                row-key="id"
              >
                <template #cell-id="{ row }">
                  <span class="font-mono text-xs">{{ (row as AdminInstanceListItem).id }}</span>
                </template>
                <template #cell-status="{ row }">
                  <span
                    class="px-2 py-0.5 rounded-full text-[10px] font-bold uppercase"
                    :class="(row as AdminInstanceListItem).status === 'running' ? 'bg-green-100 text-green-700' : 'bg-slate-100 text-slate-600'"
                  >
                    {{ (row as AdminInstanceListItem).status }}
                  </span>
                </template>
                <template #cell-actions="{ row }">
                  <div class="flex justify-end gap-2">
                    <button
                      type="button"
                      class="ui-btn ui-btn--ghost ui-btn--xs"
                      @click="handleDestroyInstance(row as AdminInstanceListItem)"
                    >
                      <Trash2 class="h-3 w-3 mr-1" />
                      销毁
                    </button>
                  </div>
                </template>
              </WorkspaceDataTable>

              <div class="mt-6">
                <WorkspaceDirectoryPagination
                  :page="page"
                  :total-pages="Math.max(1, Math.ceil(total / pageSize))"
                  :total="total"
                  total-label="个实例"
                  @change-page="handlePageChange"
                />
              </div>
            </template>
          </section>
        </div>
      </main>
    </div>
  </div>
</template>

<style scoped>
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
</style>