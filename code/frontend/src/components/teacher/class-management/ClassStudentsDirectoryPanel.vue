<script setup lang="ts">
import { ArrowRight, Search } from 'lucide-vue-next'
import { computed } from 'vue'

import type { TeacherStudentItem } from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'
import AppLoading from '@/components/common/AppLoading.vue'
import WorkspaceDataTable from '@/components/common/WorkspaceDataTable.vue'
import { ChallengeCategoryPill, toChallengeCategory } from '@/entities/challenge'

interface ClassStudentDirectoryRow {
  id: string
  student_no: string
  name: string
  username: string
  weak_dimension: string
  metrics: string
  solved_count: number
  total_score: number
  actions: 'open'
}

const props = defineProps<{
  students: TeacherStudentItem[]
  studentNoQuery: string
  loadingStudents: boolean
}>()

const emit = defineEmits<{
  updateStudentNoQuery: [value: string]
  openStudent: [studentId: string]
}>()

const rows = computed<ClassStudentDirectoryRow[]>(() =>
  props.students.map((student) => ({
    id: student.id,
    student_no: student.student_no || '未设置学号',
    name: student.name || '未设置姓名',
    username: student.username,
    weak_dimension: student.weak_dimension || '暂无薄弱项',
    metrics: `${student.solved_count ?? 0} 题 / ${student.total_score ?? 0} 分`,
    solved_count: student.solved_count ?? 0,
    total_score: student.total_score ?? 0,
    actions: 'open',
  }))
)

const columns = [
  { key: 'student_no', label: '学号', widthClass: 'w-[14%] min-w-[8rem]' },
  { key: 'name', label: '学生名称', widthClass: 'w-[20%] min-w-[11rem]' },
  { key: 'username', label: '昵称', widthClass: 'w-[18%] min-w-[10rem]' },
  { key: 'weak_dimension', label: '薄弱项', widthClass: 'w-[18%] min-w-[10rem]' },
  { key: 'metrics', label: '做题数 / 得分数', widthClass: 'w-[16%] min-w-[10rem]' },
  { key: 'actions', label: '操作', widthClass: 'w-[9rem]', align: 'right' as const },
]

function studentWeakCategory(student: { weak_dimension?: string | null }) {
  return toChallengeCategory(student.weak_dimension)
}
</script>

<template>
  <section class="teacher-student-list-section">
    <section class="teacher-directory-shell workspace-directory-list">
      <section class="teacher-directory-filters" aria-label="学生过滤">
        <div class="teacher-filter-grid">
          <label class="teacher-field">
            <span class="teacher-field-label">学号查询</span>
            <div class="teacher-field-control teacher-filter-control">
              <Search class="h-4 w-4 text-text-muted" />
              <input
                :value="studentNoQuery"
                type="text"
                placeholder="输入学号精确查询"
                class="teacher-input"
                @input="emit('updateStudentNoQuery', ($event.target as HTMLInputElement).value)"
              />
            </div>
          </label>
          <button
            v-if="studentNoQuery"
            type="button"
            class="ui-btn ui-btn--secondary teacher-filter-reset teacher-filter-clear"
            @click="emit('updateStudentNoQuery', '')"
          >
            清空学号
          </button>
        </div>
      </section>

      <div v-if="loadingStudents" class="workspace-directory-loading">
        <AppLoading>同步学生目录...</AppLoading>
      </div>

      <AppEmpty
        v-else-if="students.length === 0"
        class="teacher-empty-state workspace-directory-empty"
        icon="Users"
        title="暂无学生"
        description="该班级下还没有可用学生记录。"
      />

      <div v-else class="teacher-directory">
        <WorkspaceDataTable
          class="teacher-student-directory-table"
          :columns="columns"
          :rows="rows"
          row-key="id"
        >
          <template #cell-student_no="{ row }">
            <span class="teacher-directory-cell-student-no">
              {{ (row as ClassStudentDirectoryRow).student_no }}
            </span>
          </template>

          <template #cell-name="{ row }">
            <div class="teacher-directory-cell-name">
              <h4
                class="teacher-directory-row-title"
                :title="(row as ClassStudentDirectoryRow).name"
              >
                {{ (row as ClassStudentDirectoryRow).name }}
              </h4>
            </div>
          </template>

          <template #cell-username="{ row }">
            <span
              class="teacher-directory-row-points"
              :title="(row as ClassStudentDirectoryRow).username"
            >
              {{ (row as ClassStudentDirectoryRow).username }}
            </span>
          </template>

          <template #cell-weak_dimension="{ row }">
            <ChallengeCategoryPill
              v-if="studentWeakCategory(row as ClassStudentDirectoryRow)"
              :category="studentWeakCategory(row as ClassStudentDirectoryRow)!"
            />
            <span
              v-else
              class="teacher-directory-chip teacher-directory-chip-muted workspace-directory-status-pill workspace-directory-status-pill--muted"
            >
              {{ (row as ClassStudentDirectoryRow).weak_dimension }}
            </span>
          </template>

          <template #cell-metrics="{ row }">
            <span>{{ (row as ClassStudentDirectoryRow).metrics }}</span>
          </template>

          <template #cell-actions="{ row }">
            <div class="workspace-directory-row-actions teacher-directory-row-cta">
              <button
                type="button"
                class="ui-btn ui-btn--primary ui-btn--xs"
                :aria-label="`${(row as ClassStudentDirectoryRow).name}，${(row as ClassStudentDirectoryRow).solved_count} 题，${(row as ClassStudentDirectoryRow).total_score} 分，查看学员分析`"
                @click="emit('openStudent', (row as ClassStudentDirectoryRow).id)"
              >
                学员分析
                <ArrowRight class="h-4 w-4" />
              </button>
            </div>
          </template>
        </WorkspaceDataTable>
      </div>
    </section>
  </section>
</template>

<style scoped>
.teacher-directory-shell {
  --workspace-directory-shell-padding: var(--space-5);
  --workspace-directory-shell-radius: var(--radius-2xl);
  --workspace-directory-shell-border: color-mix(in srgb, var(--journal-border) 84%, transparent);
  --workspace-directory-shell-background:
    radial-gradient(
      circle at top right,
      color-mix(in srgb, var(--color-primary) 6%, transparent),
      transparent 38%
    ),
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--journal-surface) 98%, var(--color-bg-base)),
      color-mix(in srgb, var(--journal-surface-subtle) 74%, var(--color-bg-base))
    );
  display: grid;
  gap: var(--space-4);
  box-shadow: 0 calc(var(--space-4) + var(--space-0-5)) calc(var(--space-8) + var(--space-0-5))
    color-mix(in srgb, var(--color-shadow-soft) 20%, transparent);
}

.teacher-directory-filters {
  display: grid;
  gap: var(--space-4);
}

.teacher-filter-grid {
  display: grid;
  gap: var(--space-4);
  grid-template-columns: minmax(0, 20rem) auto;
}

.teacher-student-directory-table {
  --workspace-directory-shell-border: color-mix(in srgb, var(--teacher-card-border) 86%, transparent);
}

.teacher-directory {
  display: flex;
  flex-direction: column;
}

.teacher-directory-cell-student-no {
  font-size: var(--font-size-0-76);
  font-weight: 800;
  letter-spacing: 0.02em;
  color: var(--color-text-muted);
  font-variant-numeric: tabular-nums;
}

.teacher-directory-row-title {
  margin: 0;
  min-width: 0;
  font-size: var(--font-size-0-98);
  font-weight: 800;
  color: var(--color-text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.teacher-student-directory-table :deep(.workspace-data-table__row:hover) .teacher-directory-row-title {
  color: var(--color-primary);
}

.teacher-directory-row-points {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.teacher-filter-reset {
  align-self: end;
}

.teacher-directory-row-cta {
  justify-content: flex-end;
}

@media (max-width: 1080px) {
  .teacher-directory-row-cta {
    justify-content: flex-start;
  }
}
</style>
