<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import {
  FolderKanban,
  Users,
  Calendar,
  Plus,
  RefreshCw,
} from 'lucide-vue-next'

import type { AdminClassListItem } from '@/api/contracts'
import { getAdminClasses } from '@/api/admin'
import WorkspaceDataTable from '@/components/common/WorkspaceDataTable.vue'
import WorkspaceDirectoryPagination from '@/components/common/WorkspaceDirectoryPagination.vue'
import AppLoading from '@/components/common/AppLoading.vue'
import AppEmpty from '@/components/common/AppEmpty.vue'

const router = useRouter()
const list = ref<AdminClassListItem[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(15)
const loading = ref(false)

async function loadClasses(): Promise<void> {
  loading.value = true
  try {
    const data = await getAdminClasses({
      page: page.value,
      pageSize: pageSize.value,
    })
    list.value = data.items
    total.value = data.total
  } catch (err) {
    console.error('加载班级列表失败:', err)
  } finally {
    loading.value = false
  }
}

function handlePageChange(p: number): void {
  page.value = p
  void loadClasses()
}

onMounted(() => {
  void loadClasses()
})

const columns = [
  { key: 'name', label: '班级名称', widthClass: 'w-[30%] min-w-[12rem]' },
  { key: 'student_count', label: '学生人数', widthClass: 'w-[15%] min-w-[8rem]', align: 'center' as const },
  { key: 'teacher_name', label: '负责教师', widthClass: 'w-[15%] min-w-[10rem]' },
  { key: 'created_at', label: '创建时间', widthClass: 'w-[20%] min-w-[10rem]' },
  { key: 'actions', label: '操作', widthClass: 'w-[10rem]', align: 'right' as const },
]
</script>

<template>
  <div class="workspace-shell">
    <div class="workspace-grid">
      <main class="content-pane">
        <section class="workspace-hero">
          <div class="workspace-tab-heading__main">
            <div class="workspace-overline">
              Class Workspace
            </div>
            <h1 class="hero-title">
              班级管理
            </h1>
            <p class="hero-summary">
              在后台视角查看班级目录、学生规模，并快速进入班级详情。
            </p>
          </div>

          <div class="awd-library-hero-actions">
            <div class="quick-actions">
              <button
                type="button"
                class="ui-btn ui-btn--primary"
                @click="loadClasses()"
              >
                <RefreshCw class="h-4 w-4" />
                刷新目录
              </button>
            </div>
          </div>
        </section>

        <div class="class-manage-body mt-10 space-y-10">
          <div class="metric-panel-grid metric-panel-grid--premium cols-3">
            <article class="metric-panel-card metric-panel-card--premium">
              <div class="metric-panel-label">
                <span>班级总量</span>
                <FolderKanban class="h-4 w-4" />
              </div>
              <div class="metric-panel-value">
                {{ total.toString().padStart(2, '0') }}
              </div>
              <div class="metric-panel-helper">
                平台已接入班级
              </div>
            </article>

            <article class="metric-panel-card metric-panel-card--premium">
              <div class="metric-panel-label">
                <span>总人数</span>
                <Users class="h-4 w-4" />
              </div>
              <div class="metric-panel-value">
                00
              </div>
              <div class="metric-panel-helper">
                全平台在籍学生
              </div>
            </article>

            <article class="metric-panel-card metric-panel-card--premium">
              <div class="metric-panel-label">
                <span>教学周期</span>
                <Calendar class="h-4 w-4" />
              </div>
              <div class="metric-panel-value">
                --
              </div>
              <div class="metric-panel-helper">
                本学期教学活跃度
              </div>
            </article>
          </div>

          <section class="workspace-directory-section">
            <header class="list-heading">
              <div>
                <div class="workspace-overline">
                  Class Directory
                </div>
                <h2 class="list-heading__title">
                  班级目录
                </h2>
              </div>
            </header>

            <div
              v-if="loading && list.length === 0"
              class="py-12 flex justify-center"
            >
              <AppLoading>同步班级目录...</AppLoading>
            </div>

            <template v-else>
              <AppEmpty
                v-if="list.length === 0"
                class="workspace-directory-empty"
                icon="FolderKanban"
                title="暂无班级数据"
                description="当前平台尚未创建任何班级。"
              />

              <WorkspaceDataTable
                v-else
                class="workspace-directory-list"
                :columns="columns"
                :rows="list"
                row-key="id"
              />

              <div class="mt-6">
                <WorkspaceDirectoryPagination
                  :page="page"
                  :total-pages="Math.max(1, Math.ceil(total / pageSize))"
                  :total="total"
                  total-label="个班级"
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