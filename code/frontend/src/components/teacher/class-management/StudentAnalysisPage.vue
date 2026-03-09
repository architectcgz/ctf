<script setup lang="ts">
import { ArrowLeftRight, ChevronLeft, FileDown, GraduationCap, Users } from 'lucide-vue-next'

import type {
  MyProgressData,
  RecommendationItem,
  SkillProfileData,
  TeacherClassItem,
  TeacherStudentItem,
} from '@/api/contracts'
import AppCard from '@/components/common/AppCard.vue'
import MetricCard from '@/components/common/MetricCard.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import SectionCard from '@/components/common/SectionCard.vue'
import StudentInsightPanel from '@/components/teacher/StudentInsightPanel.vue'

const props = defineProps<{
  classes: TeacherClassItem[]
  students: TeacherStudentItem[]
  selectedClassName: string
  selectedStudentId: string
  selectedStudent: TeacherStudentItem | null
  loadingClasses: boolean
  loadingStudents: boolean
  loadingDetails: boolean
  error: string | null
  progress: MyProgressData | null
  skillProfile: SkillProfileData | null
  recommendations: RecommendationItem[]
  solvedRate: number
  weakDimensions: string[]
}>()

const emit = defineEmits<{
  retry: []
  openClassManagement: []
  openClassStudents: []
  openReportExport: []
  selectClass: [className: string]
  selectStudent: [studentId: string]
  openChallenge: [challengeId: string]
}>()
</script>

<template>
  <div class="space-y-6">
    <PageHeader
      eyebrow="Student Analysis"
      :title="selectedStudent?.name || selectedStudent?.username || '学员分析'"
      description="查看学员的能力画像、进度和推荐任务。"
    >
      <ElButton plain @click="emit('openClassManagement')">班级管理</ElButton>
      <ElButton plain @click="emit('openClassStudents')">返回学生列表</ElButton>
      <ElButton type="primary" @click="emit('openReportExport')">导出报告</ElButton>
    </PageHeader>

    <section class="grid gap-4 xl:grid-cols-[1.08fr_0.92fr]">
      <AppCard
        variant="hero"
        accent="primary"
        eyebrow="Focused Student"
        :title="selectedStudent?.name || selectedStudent?.username || '未选择学员'"
        subtitle="当前学员训练概览。"
      >
        <template #header>
          <span
            class="rounded-full border px-3 py-1 text-[11px] font-semibold uppercase tracking-[0.16em]"
            style="border-color: color-mix(in srgb, var(--color-primary) 18%, var(--color-border-default)); background-color: var(--color-primary-soft); color: var(--color-primary);"
          >
            {{ selectedClassName || '未选择班级' }}
          </span>
        </template>

        <div class="grid gap-3 md:grid-cols-3">
          <div class="rounded-[24px] border border-white/10 bg-white/6 px-4 py-4">
            <div class="text-[11px] uppercase tracking-[0.18em] text-cyan-100/60">当前学员</div>
            <div class="mt-2 text-2xl font-semibold text-white">{{ selectedStudent?.username || '未选择' }}</div>
            <div class="mt-2 text-sm text-cyan-50/70">当前聚焦的学生对象</div>
          </div>
          <div class="rounded-[24px] border border-white/10 bg-white/6 px-4 py-4">
            <div class="text-[11px] uppercase tracking-[0.18em] text-cyan-100/60">完成率</div>
            <div class="mt-2 text-2xl font-semibold text-white">{{ solvedRate }}%</div>
            <div class="mt-2 text-sm text-cyan-50/70">基于当前学员训练数据计算</div>
          </div>
          <div class="rounded-[24px] border border-white/10 bg-white/6 px-4 py-4">
            <div class="text-[11px] uppercase tracking-[0.18em] text-cyan-100/60">薄弱维度</div>
            <div class="mt-2 text-2xl font-semibold text-white">{{ weakDimensions[0] || '暂无' }}</div>
            <div class="mt-2 text-sm text-cyan-50/70">当前最需要补强的方向</div>
          </div>
        </div>
      </AppCard>

      <div class="grid gap-3 md:grid-cols-3 xl:grid-cols-1">
        <MetricCard label="同班学生" :value="students.length" hint="当前班级可切换的学生数量" accent="primary" />
        <MetricCard label="推荐任务" :value="recommendations.length" hint="当前可布置的补强题目数" accent="success" />
        <MetricCard label="查看方式" value="学生画像" hint="当前学员分析视图" accent="warning" />
      </div>
    </section>

    <div v-if="error" class="rounded-2xl border border-red-200 bg-red-50 px-5 py-4 text-sm text-red-600">
      {{ error }}
      <button type="button" class="ml-3 font-medium underline" @click="emit('retry')">重试</button>
    </div>

    <section class="grid gap-6 xl:grid-cols-[0.84fr_1.16fr]">
      <div class="space-y-6">
        <SectionCard title="班级与学生切换" subtitle="先定班级，再切换当前分析的学生。">
          <div class="space-y-4">
            <div class="flex flex-wrap gap-3">
              <button
                v-for="item in classes"
                :key="item.name"
                type="button"
                class="rounded-full px-4 py-2 text-sm font-medium transition"
                :class="item.name === selectedClassName
                  ? 'bg-[var(--color-primary)] text-white'
                  : 'border border-[var(--color-border-default)] bg-[var(--color-bg-base)] text-[var(--color-text-primary)] hover:border-[var(--color-primary)]/60'"
                @click="emit('selectClass', item.name)"
              >
                {{ item.name }} · {{ item.student_count || 0 }}
              </button>
            </div>

            <div v-if="loadingClasses || loadingStudents" class="space-y-3">
              <div v-for="index in 4" :key="index" class="h-16 animate-pulse rounded-xl bg-[var(--color-bg-base)]" />
            </div>

            <div v-else class="grid gap-3">
              <AppCard
                v-for="student in students"
                :key="student.id"
                as="button"
                variant="action"
                :accent="student.id === selectedStudentId ? 'primary' : 'neutral'"
                interactive
                class="text-left"
                @click="emit('selectStudent', student.id)"
              >
                <div class="flex items-center justify-between gap-3">
                  <div class="flex items-center gap-3">
                    <div class="flex h-10 w-10 items-center justify-center rounded-2xl border border-primary/16 bg-primary/10 text-primary">
                      <GraduationCap class="h-4 w-4" />
                    </div>
                    <div>
                      <div class="font-medium text-text-primary">{{ student.name || student.username }}</div>
                      <div class="mt-1 text-sm text-text-secondary">@{{ student.username }}</div>
                    </div>
                  </div>
                  <span
                    v-if="student.id === selectedStudentId"
                    class="rounded-full bg-primary/12 px-3 py-1 text-xs font-medium text-primary"
                  >
                    当前
                  </span>
                </div>
              </AppCard>
            </div>
          </div>
        </SectionCard>

        <SectionCard title="操作入口" subtitle="从分析页返回上一层，或者直接导出报告。">
          <div class="grid gap-3">
            <AppCard as="button" variant="action" accent="primary" interactive class="text-left" @click="emit('openClassStudents')">
              <div class="flex items-center justify-between gap-3">
                <div class="flex items-center gap-3">
                  <div class="flex h-10 w-10 items-center justify-center rounded-2xl border border-primary/16 bg-primary/10 text-primary">
                    <Users class="h-4 w-4" />
                  </div>
                  <div>
                    <div class="font-medium text-text-primary">返回学生列表</div>
                    <div class="mt-1 text-sm text-text-secondary">回到当前班级，查看全部学生。</div>
                  </div>
                </div>
                <ChevronLeft class="h-4 w-4 text-primary" />
              </div>
            </AppCard>

            <AppCard as="button" variant="action" accent="warning" interactive class="text-left" @click="emit('openReportExport')">
              <div class="flex items-center justify-between gap-3">
                <div class="flex items-center gap-3">
                  <div class="flex h-10 w-10 items-center justify-center rounded-2xl border border-amber-500/16 bg-amber-500/10 text-amber-300">
                    <FileDown class="h-4 w-4" />
                  </div>
                  <div>
                    <div class="font-medium text-text-primary">导出班级报告</div>
                    <div class="mt-1 text-sm text-text-secondary">从当前教师路径直接进入报告导出。</div>
                  </div>
                </div>
                <ArrowLeftRight class="h-4 w-4 text-amber-300" />
              </div>
            </AppCard>
          </div>
        </SectionCard>
      </div>

      <StudentInsightPanel
        :student="selectedStudent"
        :progress="progress"
        :profile="skillProfile"
        :recommendations="recommendations"
        :loading="loadingDetails"
        empty-text="请先从左侧选择一名学生。"
        @open-challenge="emit('openChallenge', $event)"
      />
    </section>
  </div>
</template>
