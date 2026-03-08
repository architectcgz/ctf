<script setup lang="ts">
import { ChevronLeft, ChevronRight } from 'lucide-vue-next'

import type { TeacherClassItem, TeacherStudentItem } from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'
import MetricCard from '@/components/common/MetricCard.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import SectionCard from '@/components/common/SectionCard.vue'

const props = defineProps<{
  classes: TeacherClassItem[]
  selectedClassName: string
  students: TeacherStudentItem[]
  studentNoQuery: string
  loadingStudents: boolean
  error: string | null
}>()

const emit = defineEmits<{
  retry: []
  openClassManagement: []
  openDashboard: []
  openReportExport: []
  updateStudentNoQuery: [value: string]
  openStudent: [studentId: string]
}>()
</script>

<template>
  <div class="space-y-6">
    <PageHeader
      eyebrow="Class Students"
      :title="selectedClassName ? `${selectedClassName} · 学生列表` : '班级学生'"
      description="查看当前班级的学生名单。"
    >
      <ElButton plain @click="emit('openClassManagement')">返回班级管理</ElButton>
      <ElButton plain @click="emit('openDashboard')">教学概览</ElButton>
      <ElButton type="primary" @click="emit('openReportExport')">导出报告</ElButton>
    </PageHeader>

    <section class="grid gap-4 md:grid-cols-3">
      <div class="grid gap-3 md:col-span-3 md:grid-cols-3">
        <MetricCard
          label="可访问班级"
          :value="classes.length"
          hint="教师账号可切换的班级数量"
          accent="primary"
        />
        <MetricCard
          label="当前学生"
          :value="students.length"
          hint="当前班级已加载学生数"
          accent="success"
        />
        <MetricCard
          label="浏览模式"
          value="班级 → 学生"
          hint="选择学生后进入分析页面"
          accent="warning"
        />
      </div>
    </section>

    <div
      v-if="error"
      class="rounded-2xl border border-red-200 bg-red-50 px-5 py-4 text-sm text-red-600"
    >
      {{ error }}
      <button type="button" class="ml-3 font-medium underline" @click="emit('retry')">重试</button>
    </div>

    <section class="grid gap-6">
      <SectionCard title="学生名单" subtitle="选择学生后进入学员分析。">
        <div class="mb-4 flex items-center justify-between">
          <div class="text-sm text-text-secondary">共 {{ students.length }} 名学生</div>
          <button
            type="button"
            class="inline-flex items-center gap-2 text-sm font-medium text-primary transition hover:text-primary/80"
            @click="emit('openClassManagement')"
          >
            <ChevronLeft class="h-4 w-4" />
            返回班级列表
          </button>
        </div>

        <label class="mb-4 block space-y-2">
          <span class="text-sm text-text-secondary">按学号查询</span>
          <input
            :value="studentNoQuery"
            type="text"
            placeholder="输入学号后实时查询"
            class="w-full rounded-xl border border-border bg-surface px-4 py-3 text-sm text-slate-100 outline-none transition focus:border-primary"
            @input="emit('updateStudentNoQuery', ($event.target as HTMLInputElement).value)"
          />
        </label>

        <div v-if="loadingStudents" class="space-y-3">
          <div
            v-for="index in 6"
            :key="index"
            class="h-14 animate-pulse rounded-2xl bg-[var(--color-bg-base)]"
          />
        </div>

        <AppEmpty
          v-else-if="students.length === 0"
          icon="Users"
          title="当前班级暂无学生"
          description="该班级下还没有可用学生记录。"
        />

        <div v-else>
          <ElTable
            :data="students"
            row-key="id"
            class="teacher-student-table"
            empty-text="当前班级暂无学生"
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

            <ElTableColumn label="操作" width="180" align="right">
              <template #default="{ row }">
                <ElButton type="primary" plain @click="emit('openStudent', row.id)">
                  查看学员分析
                  <ChevronRight class="ml-1 h-4 w-4" />
                </ElButton>
              </template>
            </ElTableColumn>
          </ElTable>
        </div>
      </SectionCard>
    </section>
  </div>
</template>

<style scoped>
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
