<script setup lang="ts">
import { computed, ref } from 'vue'

import AppEmpty from '@/components/common/AppEmpty.vue'
import AppLoading from '@/components/common/AppLoading.vue'
import WorkspaceDataTable from '@/components/common/WorkspaceDataTable.vue'
import WorkspaceDirectoryPagination from '@/components/common/WorkspaceDirectoryPagination.vue'
import WorkspaceDirectoryToolbar from '@/components/common/WorkspaceDirectoryToolbar.vue'

interface ClassManageTableRow {
  id: string
  name: string
  student_count: number
  teacher_name: string
  created_at: string
  actions: string
  rowIndex: number
}

type ClassStatusFilter = 'ready' | 'empty' | ''

const props = defineProps<{
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

const filterQuery = ref('')
const statusFilter = ref<ClassStatusFilter>('')

const columns = [
  { key: 'name', label: '班级名称', widthClass: 'w-[30%] min-w-[12rem]' },
  { key: 'student_count', label: '学生人数', widthClass: 'w-[15%] min-w-[8rem]', align: 'center' as const },
  { key: 'teacher_name', label: '负责教师', widthClass: 'w-[15%] min-w-[10rem]' },
  { key: 'created_at', label: '创建时间', widthClass: 'w-[20%] min-w-[10rem]' },
  { key: 'actions', label: '操作', widthClass: 'w-[10rem]', align: 'right' as const },
]

const filteredRows = computed(() => {
  const keyword = filterQuery.value.trim().toLowerCase()

  return props.rows.filter((row) => {
    const status: Exclude<ClassStatusFilter, ''> = row.student_count > 0 ? 'ready' : 'empty'
    const matchesKeyword = !keyword || row.name.toLowerCase().includes(keyword)
    const matchesStatus = !statusFilter.value || status === statusFilter.value
    return matchesKeyword && matchesStatus
  })
})

const hasActiveFilters = computed(() => Boolean(filterQuery.value.trim() || statusFilter.value))

function resetFilters(): void {
  filterQuery.value = ''
  statusFilter.value = ''
}

function handleStatusFilterChange(event: Event): void {
  const target = event.target
  statusFilter.value = target instanceof HTMLSelectElement ? target.value as ClassStatusFilter : ''
}
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

      <WorkspaceDirectoryToolbar
        v-model="filterQuery"
        :total="filteredRows.length"
        selected-sort-label=""
        :sort-options="[]"
        search-placeholder="搜索班级名称..."
        total-suffix="个班级"
        filter-panel-title="班级筛选"
        reset-label="清空筛选"
        :reset-disabled="!hasActiveFilters"
        @reset-filters="resetFilters"
      >
        <template #filter-panel>
          <label class="admin-class-manage-filter-field">
            <span class="workspace-overline">班级状态</span>
            <select
              :value="statusFilter"
              class="admin-input workspace-directory-filter-control admin-class-manage-filter-control"
              @change="handleStatusFilterChange"
            >
              <option value="">
                全部状态
              </option>
              <option value="ready">
                已有学生
              </option>
              <option value="empty">
                暂无学生
              </option>
            </select>
          </label>
        </template>
      </WorkspaceDirectoryToolbar>

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

        <AppEmpty
          v-else-if="filteredRows.length === 0"
          class="workspace-directory-empty"
          icon="Search"
          title="没有匹配班级"
          description="调整搜索关键词或筛选条件后再试。"
        />

        <WorkspaceDataTable
          v-else
          class="workspace-directory-list admin-class-manage-table"
          :columns="columns"
          :rows="filteredRows"
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

        <WorkspaceDirectoryPagination
          v-if="total > 0 && filteredRows.length > 0"
          :page="page"
          :total-pages="totalPages"
          :total="total"
          total-label="个班级"
          @change-page="emit('change-page', $event)"
        />
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

.admin-class-manage-filter-field {
  display: grid;
  gap: var(--space-2);
}

</style>
