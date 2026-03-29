<script setup lang="ts">
import { ChevronLeft, ChevronRight } from 'lucide-vue-next'

import { computed } from 'vue'

import type {
  TeacherClassItem,
  TeacherClassReviewData,
  TeacherClassSummaryData,
  TeacherClassTrendData,
  TeacherStudentItem,
} from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'
import MetricCard from '@/components/common/MetricCard.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import SectionCard from '@/components/common/SectionCard.vue'
import TeacherClassInsightsPanel from '@/components/teacher/TeacherClassInsightsPanel.vue'
import TeacherInterventionPanel from '@/components/teacher/TeacherInterventionPanel.vue'
import TeacherClassReviewPanel from '@/components/teacher/TeacherClassReviewPanel.vue'
import TeacherClassTrendPanel from '@/components/teacher/TeacherClassTrendPanel.vue'

const props = defineProps<{
  classes: TeacherClassItem[]
  selectedClassName: string
  students: TeacherStudentItem[]
  review: TeacherClassReviewData | null
  summary: TeacherClassSummaryData | null
  trend: TeacherClassTrendData | null
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

const averageSolvedText = computed(() => {
  if (!props.summary) return '--'
  return props.summary.average_solved.toFixed(1)
})

const activeRateText = computed(() => {
  if (!props.summary) return '--'
  return `${Math.round(props.summary.active_rate)}%`
})
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
          label="班级人数"
          :value="props.summary?.student_count ?? students.length"
          hint="当前班级纳入统计的学生数"
          accent="success"
        />
        <MetricCard
          label="平均解题"
          :value="averageSolvedText"
          hint="当前班级学生的人均解题数"
          accent="warning"
        />
      </div>
    </section>

    <section class="grid gap-4 md:grid-cols-2">
      <MetricCard
        label="近 7 天活跃率"
        :value="activeRateText"
        hint="近 7 天至少有一次训练动作的学生占比"
        accent="primary"
      />
      <MetricCard
        label="近 7 天训练事件"
        :value="props.summary?.recent_event_count ?? '--'"
        hint="提交、实例启动与销毁等动作总数"
        accent="success"
      />
    </section>

    <div
      v-if="error"
      class="rounded-2xl border border-[var(--color-danger)]/20 bg-[var(--color-danger)]/10 px-5 py-4 text-sm text-[var(--color-danger)]"
    >
      {{ error }}
      <button type="button" class="ml-3 font-medium underline" @click="emit('retry')">重试</button>
    </div>

    <section class="grid gap-6">
      <TeacherClassTrendPanel
        :trend="trend"
        title="班级近 7 天训练趋势"
        subtitle="先看整体节奏，再下钻到具体学生。"
      />

      <TeacherClassReviewPanel
        :review="review"
        :class-name="selectedClassName"
      />

      <TeacherClassInsightsPanel :students="students" :class-name="selectedClassName" />

      <TeacherInterventionPanel :students="students" :class-name="selectedClassName" />

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
            class="w-full rounded-xl border border-border bg-surface px-4 py-3 text-sm text-[var(--color-text-primary)] outline-none transition focus:border-primary"
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

            <ElTableColumn label="解题数" width="120" align="center">
              <template #default="{ row }">
                <span class="text-sm font-medium text-text-primary">{{
                  row.solved_count ?? 0
                }}</span>
              </template>
            </ElTableColumn>

            <ElTableColumn label="得分" width="120" align="center">
              <template #default="{ row }">
                <span class="text-sm font-medium text-text-primary">{{
                  row.total_score ?? 0
                }}</span>
              </template>
            </ElTableColumn>

            <ElTableColumn label="薄弱项" min-width="160">
              <template #default="{ row }">
                <span class="text-sm text-text-secondary">{{ row.weak_dimension || '暂无' }}</span>
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
