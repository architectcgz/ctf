<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'

import { getClasses } from '@/api/teacher'
import type { TeacherClassItem } from '@/api/contracts'
import WorkspaceDataTable from '@/components/common/WorkspaceDataTable.vue'
import WorkspaceDirectoryPagination from '@/components/common/WorkspaceDirectoryPagination.vue'
import AppLoading from '@/components/common/AppLoading.vue'
import AppEmpty from '@/components/common/AppEmpty.vue'
import ClassManageHeroPanel from '@/components/platform/class/ClassManageHeroPanel.vue'
import { DEFAULT_PAGE_SIZE } from '@/utils/constants'

const router = useRouter()
const list = ref<TeacherClassItem[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(DEFAULT_PAGE_SIZE)
const loading = ref(false)
const error = ref<string | null>(null)

async function loadClasses(): Promise<void> {
  loading.value = true
  error.value = null
  try {
    const data = await getClasses({
      page: page.value,
      page_size: pageSize.value,
    })
    if (Array.isArray(data)) {
      list.value = data
      total.value = data.length
      return
    }

    list.value = data.list
    total.value = data.total
    page.value = data.page
    pageSize.value = data.page_size
  } catch (err) {
    console.error('加载班级列表失败:', err)
    error.value = '加载班级列表失败，请稍后重试'
    list.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

function handlePageChange(p: number): void {
  page.value = p
  void loadClasses()
}

function openClass(className: string): void {
  void router.push({
    name: 'PlatformClassStudents',
    params: { className },
  })
}

const totalStudents = computed(() =>
  list.value.reduce((sum, item) => sum + (item.student_count || 0), 0)
)

const rows = computed(() =>
  list.value.map((item, index) => ({
    id: item.name,
    name: item.name,
    student_count: item.student_count || 0,
    teacher_name: '--',
    created_at: '--',
    actions: '查看班级',
    rowIndex: index,
  }))
)

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
        <ClassManageHeroPanel
          :total="total"
          :total-students="totalStudents"
          @refresh="void loadClasses()"
        />

        <div class="admin-class-manage-shell__content">
          <section class="workspace-directory-section admin-class-manage-directory">
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
                class="workspace-directory-list admin-class-manage-table"
                :columns="columns"
                :rows="rows"
                row-key="id"
              >
                <template #cell-actions="{ row }">
                  <button
                    type="button"
                    class="ui-btn ui-btn--ghost"
                    @click="openClass(String((row as { name: string }).name))"
                  >
                    查看班级
                  </button>
                </template>
              </WorkspaceDataTable>

              <div class="workspace-directory-pagination">
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
.admin-class-manage-shell__content {
  display: flex;
  flex-direction: column;
  gap: var(--workspace-directory-page-block-gap);
  margin-top: var(--space-10);
}

</style>
