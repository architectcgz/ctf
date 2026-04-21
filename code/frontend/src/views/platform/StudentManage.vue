<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { 
  Users, 
  GraduationCap, 
  UserPlus, 
  RefreshCw,
  Search,
  Filter
} from 'lucide-vue-next'

import { usePlatformStudentDirectory } from '@/composables/usePlatformStudentDirectory'
import WorkspaceDataTable from '@/components/common/WorkspaceDataTable.vue'
import WorkspaceDirectoryPagination from '@/components/common/WorkspaceDirectoryPagination.vue'
import WorkspaceDirectoryToolbar from '@/components/common/WorkspaceDirectoryToolbar.vue'
import AppLoading from '@/components/common/AppLoading.vue'
import AppEmpty from '@/components/common/AppEmpty.vue'

const studentDirectoryQuery = usePlatformStudentDirectory()

const {
  list,
  total,
  page,
  pageSize,
  loading,
  keyword,
  classFilter,
  clearFilters,
} = studentDirectoryQuery

async function initialize(): Promise<void> {
  await studentDirectoryQuery.loadStudents({
    page: page.value,
    pageSize: pageSize.value,
    keyword: keyword.value,
    classId: classFilter.value,
  })
}

const directoryParams = computed(() => ({
  page: page.value,
  pageSize: pageSize.value,
  keyword: keyword.value,
  classId: classFilter.value,
}))

watch(directoryParams, () => {
  void studentDirectoryQuery.loadStudents(directoryParams.value)
})

onMounted(() => {
  void initialize()
})

const columns = [
  { key: 'username', label: '用户名', widthClass: 'w-[15%]' },
  { key: 'nickname', label: '昵称', widthClass: 'w-[15%]' },
  { key: 'email', label: '邮箱', widthClass: 'w-[20%]' },
  { key: 'class_name', label: '所属班级', widthClass: 'w-[15%]' },
  { key: 'created_at', label: '加入时间', widthClass: 'w-[15%]' },
  { key: 'actions', label: '操作', widthClass: 'w-[10%]', align: 'right' as const },
]
</script>

<template>
  <div class="workspace-shell">
    <div class="workspace-grid">
      <main class="content-pane">
        <section class="workspace-hero">
          <div class="workspace-tab-heading__main">
            <div class="workspace-overline">
              Student Workspace
            </div>
            <h1 class="hero-title">
              学生管理
            </h1>
            <p class="hero-summary">
              在后台视角查看学生目录、班级归属与学习表现，并快速进入学员分析。
            </p>
          </div>

          <div class="awd-library-hero-actions">
            <div class="quick-actions">
              <button
                type="button"
                class="ui-btn ui-btn--primary"
                @click="initialize()"
              >
                <RefreshCw class="h-4 w-4" />
                刷新目录
              </button>
            </div>
          </div>
        </section>

        <div class="student-manage-body mt-10 space-y-10">
          <div class="metric-panel-grid metric-panel-grid--premium cols-3">
            <article class="metric-panel-card metric-panel-card--premium">
              <div class="metric-panel-label">
                <span>学生总量</span>
                <Users class="h-4 w-4" />
              </div>
              <div class="metric-panel-value">
                {{ total.toString().padStart(2, '0') }}
              </div>
              <div class="metric-panel-helper">
                平台注册学员总数
              </div>
            </article>

            <article class="metric-panel-card metric-panel-card--premium">
              <div class="metric-panel-label">
                <span>活跃学员</span>
                <Activity class="h-4 w-4" />
              </div>
              <div class="metric-panel-value">
                {{ total.toString().padStart(2, '0') }}
              </div>
              <div class="metric-panel-helper">
                最近 30 天有登录记录
              </div>
            </article>

            <article class="metric-panel-card metric-panel-card--premium">
              <div class="metric-panel-label">
                <span>正式班级</span>
                <GraduationCap class="h-4 w-4" />
              </div>
              <div class="metric-panel-value">
                00
              </div>
              <div class="metric-panel-helper">
                已分配班级的学员
              </div>
            </article>
          </div>

          <section class="workspace-directory-section">
            <WorkspaceDirectoryToolbar
              v-model="keyword"
              :total="total"
              search-placeholder="检索学生姓名或邮箱..."
              @reset-filters="clearFilters"
            />

            <div
              v-if="loading && list.length === 0"
              class="py-12 flex justify-center"
            >
              <AppLoading>同步学员目录...</AppLoading>
            </div>

            <template v-else>
              <AppEmpty
                v-if="list.length === 0"
                class="workspace-directory-empty"
                icon="Users"
                title="暂无学生数据"
                description="当前平台上没有任何学生账号。"
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
                  total-label="名学生"
                  @change-page="page = $event"
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