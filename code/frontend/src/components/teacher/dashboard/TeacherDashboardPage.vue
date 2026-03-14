<script setup lang="ts">
import { computed } from 'vue'

import type {
  TeacherClassItem,
  TeacherClassReviewData,
  TeacherClassSummaryData,
  TeacherClassTrendData,
  TeacherStudentItem,
} from '@/api/contracts'
import PageHeader from '@/components/common/PageHeader.vue'
import TeacherClassInsightsPanel from '@/components/teacher/TeacherClassInsightsPanel.vue'
import TeacherInterventionPanel from '@/components/teacher/TeacherInterventionPanel.vue'
import TeacherClassReviewPanel from '@/components/teacher/TeacherClassReviewPanel.vue'
import TeacherClassTrendPanel from '@/components/teacher/TeacherClassTrendPanel.vue'

const props = defineProps<{
  classes: TeacherClassItem[]
  students: TeacherStudentItem[]
  selectedClassName: string
  selectedClass: TeacherClassItem | null
  review: TeacherClassReviewData | null
  summary: TeacherClassSummaryData | null
  trend: TeacherClassTrendData | null
  error: string | null
}>()

const emit = defineEmits<{
  retry: []
  openClassManagement: []
  openReportExport: []
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
    <PageHeader eyebrow="Teacher Flight Deck" title="教学介入台" description="查看班级状态。">
      <ElButton plain @click="emit('openClassManagement')">班级管理</ElButton>
      <ElButton type="primary" @click="emit('openReportExport')">导出报告</ElButton>
    </PageHeader>

    <section>
      <div
        class="rounded-[30px] border border-cyan-500/20 bg-[linear-gradient(145deg,rgba(8,47,73,0.82),rgba(15,23,42,0.94))] p-6 shadow-[0_24px_70px_var(--color-shadow-soft)]"
      >
        <div
          class="flex flex-wrap items-center gap-2 text-[11px] font-semibold uppercase tracking-[0.22em] text-cyan-100/75"
        >
          <span>Intervention Deck</span>
          <span class="rounded-full border border-white/10 bg-white/5 px-2 py-1">{{
            selectedClassName || '未选择班级'
          }}</span>
        </div>
        <h2 class="mt-3 text-3xl font-semibold tracking-tight text-white">
          {{ selectedClassName ? `${selectedClassName} 的教学概览` : '先选择一个班级' }}
        </h2>
        <p class="mt-3 text-sm leading-7 text-cyan-50/80">
          {{
            selectedClassName
              ? '查看当前班级人数、学生数量和教学入口。'
              : '选择班级后查看当前班级概览。'
          }}
        </p>

        <div class="mt-6 grid gap-3 md:grid-cols-3">
          <div class="rounded-[24px] border border-white/10 bg-white/6 px-4 py-4">
            <div class="text-[11px] uppercase tracking-[0.18em] text-cyan-100/60">班级人数</div>
            <div class="mt-2 text-2xl font-semibold text-white">
              {{ summary?.student_count || selectedClass?.student_count || students.length }}
            </div>
            <div class="mt-2 text-sm text-cyan-50/70">当前班级纳入视图的人数</div>
          </div>
          <div class="rounded-[24px] border border-white/10 bg-white/6 px-4 py-4">
            <div class="text-[11px] uppercase tracking-[0.18em] text-cyan-100/60">平均解题</div>
            <div class="mt-2 text-2xl font-semibold text-white">{{ averageSolvedText }}</div>
            <div class="mt-2 text-sm text-cyan-50/70">当前班级学生的人均解题数</div>
          </div>
          <div class="rounded-[24px] border border-white/10 bg-white/6 px-4 py-4">
            <div class="text-[11px] uppercase tracking-[0.18em] text-cyan-100/60">
              近 7 天活跃率
            </div>
            <div class="mt-2 text-2xl font-semibold text-white">{{ activeRateText }}</div>
            <div class="mt-2 text-sm text-cyan-50/70">近 7 天有训练动作的学生占比</div>
          </div>
        </div>

        <div class="mt-4 grid gap-3 md:grid-cols-2">
          <div class="rounded-[24px] border border-white/10 bg-white/6 px-4 py-4">
            <div class="text-[11px] uppercase tracking-[0.18em] text-cyan-100/60">
              近 7 天活跃学生
            </div>
            <div class="mt-2 text-2xl font-semibold text-white">
              {{ summary?.active_student_count ?? '--' }}
            </div>
            <div class="mt-2 text-sm text-cyan-50/70">至少有一次训练动作的学生数量</div>
          </div>
          <div class="rounded-[24px] border border-white/10 bg-white/6 px-4 py-4">
            <div class="text-[11px] uppercase tracking-[0.18em] text-cyan-100/60">
              近 7 天训练事件
            </div>
            <div class="mt-2 text-2xl font-semibold text-white">
              {{ summary?.recent_event_count ?? '--' }}
            </div>
            <div class="mt-2 text-sm text-cyan-50/70">提交、实例启动和销毁等训练动作总数</div>
          </div>
        </div>
      </div>
    </section>

    <div
      v-if="error"
      class="rounded-2xl border border-red-200 bg-red-50 px-5 py-4 text-sm text-red-600"
    >
      {{ error }}
      <button type="button" class="ml-3 font-medium underline" @click="emit('retry')">重试</button>
    </div>

    <TeacherClassTrendPanel
      :trend="trend"
      title="班级近 7 天训练趋势"
      subtitle="把训练事件、成功解题和活跃学生放在同一条时间轴上观察。"
    />

    <TeacherClassReviewPanel
      :review="review"
      :class-name="selectedClassName"
    />

    <TeacherClassInsightsPanel :students="students" :class-name="selectedClassName" />

    <TeacherInterventionPanel :students="students" :class-name="selectedClassName" />
  </div>
</template>
