<script setup lang="ts">
import type { TeacherClassItem } from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'
import AppLoading from '@/components/common/AppLoading.vue'
import WorkspaceDataTable from '@/components/common/WorkspaceDataTable.vue'
import WorkspaceDirectoryPagination from '@/components/common/WorkspaceDirectoryPagination.vue'
import WorkspaceDirectoryToolbar from '@/components/common/WorkspaceDirectoryToolbar.vue'

interface StudentManageTableRow {
  id: string
  name: string
  username: string
  student_no: string
  class_name: string
  total_score: number
  actions: string
}

defineProps<{
  classes: TeacherClassItem[]
  loading: boolean
  loadingClasses: boolean
  error: string | null
  keyword: string
  classFilter: string
  total: number
  hasActiveFilters: boolean
  rows: StudentManageTableRow[]
  page: number
  totalPages: number
}>()

const emit = defineEmits<{
  (event: 'update:keyword', value: string): void
  (event: 'change:class-filter', value: string): void
  (event: 'reset-filters'): void
  (event: 'change-page', value: number): void
  (event: 'open-student', studentId: string): void
}>()

const columns = [
  { key: 'name', label: '学生姓名', widthClass: 'w-[20%] min-w-[12rem]' },
  { key: 'username', label: '用户名', widthClass: 'w-[18%] min-w-[10rem]' },
  { key: 'student_no', label: '学号', widthClass: 'w-[18%] min-w-[10rem]' },
  { key: 'class_name', label: '班级', widthClass: 'w-[18%] min-w-[10rem]' },
  { key: 'total_score', label: '得分', widthClass: 'w-[12%] min-w-[8rem]', align: 'center' as const },
  { key: 'actions', label: '操作', widthClass: 'w-[12%] min-w-[8rem]', align: 'right' as const },
]
</script>

<template>
  <div class="admin-student-manage-shell__content">
    <section class="workspace-directory-section admin-student-manage-directory">
      <WorkspaceDirectoryToolbar
        :model-value="keyword"
        :total="total"
        selected-sort-label=""
        :sort-options="[]"
        search-placeholder="检索姓名、用户名或学号..."
        filter-panel-title="学生筛选"
        total-suffix="名学生"
        reset-label="清空筛选"
        :reset-disabled="!hasActiveFilters"
        @update:model-value="emit('update:keyword', $event)"
        @reset-filters="emit('reset-filters')"
      >
        <template #filter-panel>
          <label class="workspace-student-filter-field">
            <span class="workspace-overline">班级范围</span>
            <select
              :value="classFilter"
              class="admin-input workspace-student-filter-control"
              @change="emit('change:class-filter', ($event.target as HTMLSelectElement).value)"
            >
              <option value="">
                全部班级
              </option>
              <option
                v-for="item in classes"
                :key="item.name"
                :value="item.name"
              >
                {{ item.name }}
              </option>
            </select>
          </label>
        </template>
      </WorkspaceDirectoryToolbar>

      <div
        v-if="(loading || loadingClasses) && rows.length === 0"
        class="py-12 flex justify-center"
      >
        <AppLoading>同步学员目录...</AppLoading>
      </div>

      <template v-else>
        <AppEmpty
          v-if="rows.length === 0"
          class="workspace-directory-empty"
          icon="Users"
          title="暂无学生数据"
          description="当前平台上没有任何学生账号。"
        />

        <WorkspaceDataTable
          v-else
          class="workspace-directory-list admin-student-manage-table"
          :columns="columns"
          :rows="rows"
          row-key="id"
        >
          <template #cell-actions="{ row }">
            <button
              type="button"
              class="ui-btn ui-btn--primary ui-btn--sm"
              @click="emit('open-student', String((row as StudentManageTableRow).id))"
            >
              查看学员
            </button>
          </template>
        </WorkspaceDataTable>

        <div class="workspace-directory-pagination">
          <WorkspaceDirectoryPagination
            :page="page"
            :total-pages="totalPages"
            :total="total"
            total-label="名学生"
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
.admin-student-manage-shell__content {
  display: flex;
  flex-direction: column;
  gap: var(--workspace-directory-page-block-gap);
  margin-top: var(--space-10);
}
</style>
