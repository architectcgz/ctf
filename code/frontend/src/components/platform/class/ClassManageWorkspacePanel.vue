<script setup lang="ts">
import AppEmpty from '@/components/common/AppEmpty.vue'
import AppLoading from '@/components/common/AppLoading.vue'
import WorkspaceDataTable from '@/components/common/WorkspaceDataTable.vue'
import WorkspaceDirectoryPagination from '@/components/common/WorkspaceDirectoryPagination.vue'

interface ClassManageTableRow {
  id: string
  name: string
  student_count: number
  teacher_name: string
  created_at: string
  actions: string
  rowIndex: number
}

defineProps<{
  loading: boolean
  hasClasses: boolean
  rows: ClassManageTableRow[]
  page: number
  totalPages: number
  total: number
  error: string | null
}>()

const emit = defineEmits<{
  (event: 'open-class', className: string): void
  (event: 'change-page', page: number): void
}>()

const columns = [
  { key: 'name', label: '班级名称', widthClass: 'w-[30%] min-w-[12rem]' },
  { key: 'student_count', label: '学生人数', widthClass: 'w-[15%] min-w-[8rem]', align: 'center' as const },
  { key: 'teacher_name', label: '负责教师', widthClass: 'w-[15%] min-w-[10rem]' },
  { key: 'created_at', label: '创建时间', widthClass: 'w-[20%] min-w-[10rem]' },
  { key: 'actions', label: '操作', widthClass: 'w-[10rem]', align: 'right' as const },
]
</script>

<template>
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
        v-if="loading && !hasClasses"
        class="py-12 flex justify-center"
      >
        <AppLoading>同步班级目录...</AppLoading>
      </div>

      <template v-else>
        <AppEmpty
          v-if="!hasClasses"
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
              class="ui-btn ui-btn--primary ui-btn--sm"
              @click="emit('open-class', String((row as ClassManageTableRow).name))"
            >
              查看班级
            </button>
          </template>
        </WorkspaceDataTable>

        <div class="workspace-directory-pagination">
          <WorkspaceDirectoryPagination
            :page="page"
            :total-pages="totalPages"
            :total="total"
            total-label="个班级"
            @change-page="emit('change-page', $event)"
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
</template>

<style scoped>
.admin-class-manage-shell__content {
  display: flex;
  flex-direction: column;
  gap: var(--workspace-directory-page-block-gap);
  margin-top: var(--space-10);
}
</style>
