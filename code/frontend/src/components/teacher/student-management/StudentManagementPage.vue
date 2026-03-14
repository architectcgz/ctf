<script setup lang="ts">
import { Search, Users } from 'lucide-vue-next'

import type { TeacherClassItem, TeacherStudentItem } from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'
import MetricCard from '@/components/common/MetricCard.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import SectionCard from '@/components/common/SectionCard.vue'

const props = defineProps<{
  classes: TeacherClassItem[]
  selectedClassName: string
  searchQuery: string
  studentNoQuery: string
  filteredStudents: TeacherStudentItem[]
  totalStudents: number
  loadingClasses: boolean
  loadingStudents: boolean
  error: string | null
}>()

const emit = defineEmits<{
  retry: []
  openClassManagement: []
  openReportExport: []
  updateSearchQuery: [value: string]
  updateStudentNoQuery: [value: string]
  selectClass: [className: string]
  openStudent: [studentId: string]
}>()
</script>

<template>
  <div class="space-y-6">
    <PageHeader eyebrow="Student Directory" title="学生管理" description="按班级筛选并搜索学生。">
      <ElButton plain @click="emit('openClassManagement')">班级管理</ElButton>
      <ElButton type="primary" @click="emit('openReportExport')">导出报告</ElButton>
    </PageHeader>

    <section class="grid gap-4 md:grid-cols-3">
      <MetricCard
        label="可访问班级"
        :value="classes.length"
        hint="当前教师可切换的班级数量"
        accent="primary"
      />
      <MetricCard
        label="当前班级学生"
        :value="totalStudents"
        hint="当前选中班级的学生总数"
        accent="success"
      />
      <MetricCard
        label="搜索结果"
        :value="filteredStudents.length"
        hint="当前搜索条件下匹配的学生数量"
        accent="warning"
      />
    </section>

    <div
      v-if="error"
      class="rounded-2xl border border-red-200 bg-red-50 px-5 py-4 text-sm text-red-600"
    >
      {{ error }}
      <button type="button" class="ml-3 font-medium underline" @click="emit('retry')">重试</button>
    </div>

    <SectionCard title="学生筛选" subtitle="先选班级，再搜索学生。">
      <div class="grid gap-4 md:grid-cols-[220px_1fr_1fr]">
        <label class="space-y-2">
          <span class="text-sm text-text-secondary">班级</span>
          <select
            :value="selectedClassName"
            class="teacher-filter-field w-full rounded-xl border border-border bg-surface px-4 py-3 text-sm text-text-primary outline-none transition focus:border-primary disabled:cursor-not-allowed disabled:opacity-60"
            :disabled="loadingClasses"
            @change="emit('selectClass', ($event.target as HTMLSelectElement).value)"
          >
            <option v-for="item in classes" :key="item.name" :value="item.name">
              {{ item.name }} · {{ item.student_count || 0 }}
            </option>
          </select>
        </label>

        <label class="space-y-2">
          <span class="text-sm text-text-secondary">搜索姓名或用户名</span>
          <div
            class="teacher-filter-field flex items-center gap-2 rounded-xl border border-border bg-surface px-4 py-3"
          >
            <Search class="h-4 w-4 text-text-muted" />
            <input
              :value="searchQuery"
              type="text"
              placeholder="搜索姓名或用户名"
              class="w-full bg-transparent text-sm text-text-primary outline-none placeholder:text-text-muted"
              @input="emit('updateSearchQuery', ($event.target as HTMLInputElement).value)"
            />
          </div>
        </label>

        <label class="space-y-2">
          <span class="text-sm text-text-secondary">按学号查询</span>
          <div
            class="teacher-filter-field flex items-center gap-2 rounded-xl border border-border bg-surface px-4 py-3"
          >
            <Search class="h-4 w-4 text-text-muted" />
            <input
              :value="studentNoQuery"
              type="text"
              placeholder="输入学号精确查询"
              class="w-full bg-transparent text-sm text-text-primary outline-none placeholder:text-text-muted"
              @input="emit('updateStudentNoQuery', ($event.target as HTMLInputElement).value)"
            />
          </div>
        </label>
      </div>
    </SectionCard>

    <SectionCard title="学生列表" subtitle="搜索结果会即时收敛到当前班级。">
      <div v-if="loadingStudents" class="space-y-3">
        <div
          v-for="index in 6"
          :key="index"
          class="h-14 animate-pulse rounded-2xl bg-[var(--color-bg-base)]"
        />
      </div>

      <AppEmpty
        v-else-if="filteredStudents.length === 0"
        icon="Users"
        title="没有匹配学生"
        description="调整搜索词或切换班级后再试。"
      />

      <div v-else>
        <ElTable
          :data="filteredStudents"
          row-key="id"
          class="teacher-student-table"
          empty-text="没有匹配学生"
        >
          <ElTableColumn label="姓名" min-width="220">
            <template #default="{ row }">
              <div class="py-1">
                <div class="font-semibold text-text-primary">{{ row.name || row.username }}</div>
                <div class="mt-1 text-sm text-text-secondary">@{{ row.username }}</div>
              </div>
            </template>
          </ElTableColumn>

          <ElTableColumn prop="username" label="用户名" min-width="220">
            <template #default="{ row }">
              <span class="text-sm text-text-secondary">@{{ row.username }}</span>
            </template>
          </ElTableColumn>

          <ElTableColumn label="学号" min-width="180">
            <template #default="{ row }">
              <span class="text-sm text-text-secondary">{{ row.student_no || '未设置' }}</span>
            </template>
          </ElTableColumn>

          <ElTableColumn label="解题数" width="120" align="center">
            <template #default="{ row }">
              <span class="text-sm font-medium text-text-primary">{{ row.solved_count ?? 0 }}</span>
            </template>
          </ElTableColumn>

          <ElTableColumn label="得分" width="120" align="center">
            <template #default="{ row }">
              <span class="text-sm font-medium text-text-primary">{{ row.total_score ?? 0 }}</span>
            </template>
          </ElTableColumn>

          <ElTableColumn label="薄弱项" min-width="160">
            <template #default="{ row }">
              <span class="text-sm text-text-secondary">{{ row.weak_dimension || '暂无' }}</span>
            </template>
          </ElTableColumn>

          <ElTableColumn label="操作" width="180" align="right">
            <template #default="{ row }">
              <ElButton type="primary" plain @click="emit('openStudent', row.id)"
                >查看学员分析</ElButton
              >
            </template>
          </ElTableColumn>
        </ElTable>
      </div>
    </SectionCard>
  </div>
</template>

<style scoped>
:deep(.teacher-filter-field) {
  color: var(--color-text-primary);
}

:deep(.teacher-filter-field option) {
  background-color: var(--color-bg-surface);
  color: var(--color-text-primary);
}

:deep(.teacher-filter-field select),
:deep(.teacher-filter-field input) {
  color: var(--color-text-primary);
}

:deep(.teacher-student-table) {
  --el-table-bg-color: transparent;
  --el-table-tr-bg-color: transparent;
  --el-table-expanded-cell-bg-color: transparent;
  --el-table-header-bg-color: color-mix(in srgb, var(--color-bg-surface) 84%, var(--color-bg-base));
  --el-table-border-color: var(--color-border-default);
  --el-table-row-hover-bg-color: color-mix(
    in srgb,
    var(--color-primary) 8%,
    var(--color-bg-surface)
  );
  --el-table-text-color: var(--color-text-primary);
  --el-table-header-text-color: var(--color-text-secondary);
}

:deep(.teacher-student-table th.el-table__cell) {
  background: color-mix(in srgb, var(--color-bg-surface) 84%, var(--color-bg-base));
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.16em;
  text-transform: uppercase;
}

:deep(.teacher-student-table td.el-table__cell),
:deep(.teacher-student-table th.el-table__cell) {
  border-bottom-color: var(--color-border-default);
}

:deep(.teacher-student-table .el-table__inner-wrapper::before) {
  display: none;
}
</style>
