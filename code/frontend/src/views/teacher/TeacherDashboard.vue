<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'

import StudentInsightPanel from '@/components/teacher/StudentInsightPanel.vue'
import { useTeacherWorkspace } from '@/composables/useTeacherWorkspace'

const router = useRouter()
const {
  classes,
  students,
  selectedClassName,
  selectedStudentId,
  selectedClass,
  selectedStudent,
  loadingClasses,
  loadingStudents,
  loadingDetails,
  error,
  progress,
  skillProfile,
  recommendations,
  solvedRate,
  weakDimensions,
  initialize,
  loadStudents,
  loadStudentDetails,
} = useTeacherWorkspace()

const studentPreview = computed(() => students.value.slice(0, 6))

function openChallenge(challengeId: string): void {
  router.push(`/challenges/${challengeId}`)
}

onMounted(() => {
  initialize()
})
</script>

<template>
  <div class="space-y-6">
    <section class="rounded-[28px] border border-[var(--color-border-default)] bg-[linear-gradient(135deg,rgba(8,145,178,0.16),rgba(15,23,42,0.04))] p-7 shadow-sm">
      <div class="flex flex-col gap-5 xl:flex-row xl:items-end xl:justify-between">
        <div class="space-y-3">
          <p class="text-xs font-semibold uppercase tracking-[0.28em] text-[var(--color-primary)]/85">Teacher Ops</p>
          <div>
            <h1 class="text-3xl font-semibold tracking-tight text-[var(--color-text-primary)]">教学概览</h1>
            <p class="mt-2 max-w-3xl text-sm leading-6 text-[var(--color-text-secondary)]">
              关注班级规模、学员完成率和薄弱维度，快速定位当前最值得介入辅导的对象。
            </p>
          </div>
        </div>

        <div class="rounded-2xl border border-[var(--color-border-default)] bg-[var(--color-bg-base)] px-5 py-4">
          <p class="text-xs uppercase tracking-[0.18em] text-[var(--color-text-secondary)]">当前班级</p>
          <p class="mt-2 text-xl font-semibold text-[var(--color-text-primary)]">{{ selectedClassName || '未选择' }}</p>
        </div>
      </div>
    </section>

    <div v-if="error" class="rounded-2xl border border-red-200 bg-red-50 px-5 py-4 text-sm text-red-600">
      {{ error }}
      <button type="button" class="ml-3 font-medium underline" @click="initialize">重试</button>
    </div>

    <section class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
      <div class="rounded-2xl border border-[var(--color-border-default)] bg-[var(--color-bg-surface)] px-5 py-5 shadow-sm">
        <p class="text-xs uppercase tracking-[0.18em] text-[var(--color-text-secondary)]">管理班级</p>
        <p class="mt-3 text-3xl font-semibold text-[var(--color-text-primary)]">{{ classes.length }}</p>
      </div>
      <div class="rounded-2xl border border-[var(--color-border-default)] bg-[var(--color-bg-surface)] px-5 py-5 shadow-sm">
        <p class="text-xs uppercase tracking-[0.18em] text-[var(--color-text-secondary)]">班级人数</p>
        <p class="mt-3 text-3xl font-semibold text-[var(--color-text-primary)]">{{ selectedClass?.student_count || students.length }}</p>
      </div>
      <div class="rounded-2xl border border-[var(--color-border-default)] bg-[var(--color-bg-surface)] px-5 py-5 shadow-sm">
        <p class="text-xs uppercase tracking-[0.18em] text-[var(--color-text-secondary)]">样本完成率</p>
        <p class="mt-3 text-3xl font-semibold text-[var(--color-primary)]">{{ solvedRate }}%</p>
      </div>
      <div class="rounded-2xl border border-[var(--color-border-default)] bg-[var(--color-bg-surface)] px-5 py-5 shadow-sm">
        <p class="text-xs uppercase tracking-[0.18em] text-[var(--color-text-secondary)]">重点薄弱维度</p>
        <p class="mt-3 text-sm font-medium leading-6 text-[var(--color-text-primary)]">
          {{ weakDimensions.length > 0 ? weakDimensions.join('、') : '暂无明显薄弱项' }}
        </p>
      </div>
    </section>

    <section class="grid gap-6 xl:grid-cols-[0.95fr_1.05fr]">
      <div class="rounded-2xl border border-[var(--color-border-default)] bg-[var(--color-bg-surface)] p-6 shadow-sm">
        <div class="flex items-center justify-between gap-4">
          <div>
            <h2 class="text-lg font-semibold text-[var(--color-text-primary)]">班级切换</h2>
            <p class="mt-1 text-sm text-[var(--color-text-secondary)]">切换班级后将自动刷新学员列表和详情。</p>
          </div>
        </div>

        <div class="mt-5 flex flex-wrap gap-3">
          <button
            v-for="item in classes"
            :key="item.name"
            type="button"
            class="rounded-full px-4 py-2 text-sm font-medium transition"
            :class="item.name === selectedClassName
              ? 'bg-[var(--color-primary)] text-white'
              : 'border border-[var(--color-border-default)] bg-[var(--color-bg-base)] text-[var(--color-text-primary)] hover:border-[var(--color-primary)]/60'"
            @click="loadStudents(item.name)"
          >
            {{ item.name }} · {{ item.student_count || 0 }}
          </button>
        </div>

        <div class="mt-6">
          <div class="mb-3 flex items-center justify-between">
            <h3 class="text-sm font-semibold text-[var(--color-text-primary)]">班级样本</h3>
            <span class="text-xs text-[var(--color-text-secondary)]">{{ loadingStudents ? '加载中...' : `${students.length} 名学员` }}</span>
          </div>

          <div v-if="loadingClasses || loadingStudents" class="space-y-3">
            <div v-for="index in 4" :key="index" class="h-16 animate-pulse rounded-xl bg-[var(--color-bg-base)]"></div>
          </div>

          <div v-else-if="students.length === 0" class="rounded-xl border border-dashed border-[var(--color-border-default)] px-4 py-8 text-center text-sm text-[var(--color-text-secondary)]">
            当前班级暂无学员。
          </div>

          <div v-else class="grid gap-3">
            <button
              v-for="student in studentPreview"
              :key="student.id"
              type="button"
              class="rounded-xl border px-4 py-4 text-left transition"
              :class="student.id === selectedStudentId
                ? 'border-[var(--color-primary)] bg-[var(--color-primary)]/8'
                : 'border-[var(--color-border-default)] bg-[var(--color-bg-base)] hover:border-[var(--color-primary)]/50'"
              @click="loadStudentDetails(student.id)"
            >
              <div class="flex items-center justify-between gap-3">
                <div>
                  <p class="font-medium text-[var(--color-text-primary)]">{{ student.name || student.username }}</p>
                  <p class="mt-1 text-sm text-[var(--color-text-secondary)]">@{{ student.username }}</p>
                </div>
                <span v-if="student.id === selectedStudentId" class="rounded-full bg-[var(--color-primary)]/12 px-3 py-1 text-xs font-medium text-[var(--color-primary)]">关注中</span>
              </div>
            </button>
          </div>
        </div>
      </div>

      <StudentInsightPanel
        :student="selectedStudent"
        :progress="progress"
        :profile="skillProfile"
        :recommendations="recommendations"
        :loading="loadingDetails"
        empty-text="请选择左侧学员，查看教学建议。"
        @open-challenge="openChallenge"
      />
    </section>
  </div>
</template>
