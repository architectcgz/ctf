<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'

import StudentInsightPanel from '@/components/teacher/StudentInsightPanel.vue'
import { useTeacherWorkspace } from '@/composables/useTeacherWorkspace'

const router = useRouter()
const searchQuery = ref('')

const {
  classes,
  students,
  selectedClassName,
  selectedStudentId,
  selectedStudent,
  loadingClasses,
  loadingStudents,
  loadingDetails,
  error,
  progress,
  skillProfile,
  recommendations,
  initialize,
  loadStudents,
  loadStudentDetails,
} = useTeacherWorkspace()

const filteredStudents = computed(() => {
  const keyword = searchQuery.value.trim().toLowerCase()
  if (!keyword) return students.value
  return students.value.filter((student) => {
    const label = `${student.name || ''} ${student.username}`.toLowerCase()
    return label.includes(keyword)
  })
})

function openChallenge(challengeId: string): void {
  router.push(`/challenges/${challengeId}`)
}

onMounted(() => {
  initialize()
})
</script>

<template>
  <div class="space-y-6">
    <section class="rounded-[28px] border border-[var(--color-border-default)] bg-[linear-gradient(135deg,rgba(15,23,42,0.08),rgba(8,145,178,0.12))] p-7 shadow-sm">
      <div class="flex flex-col gap-5 xl:flex-row xl:items-end xl:justify-between">
        <div class="space-y-3">
          <p class="text-xs font-semibold uppercase tracking-[0.28em] text-[var(--color-primary)]/85">Class Control</p>
          <div>
            <h1 class="text-3xl font-semibold tracking-tight text-[var(--color-text-primary)]">班级管理</h1>
            <p class="mt-2 max-w-3xl text-sm leading-6 text-[var(--color-text-secondary)]">
              按班级筛选学员、检索目标对象，并联动查看完成进度、能力画像与推荐训练题目。
            </p>
          </div>
        </div>

        <div class="flex flex-col gap-3 sm:flex-row">
          <select
            :value="selectedClassName"
            class="rounded-xl border border-[var(--color-border-default)] bg-[var(--color-bg-base)] px-4 py-3 text-sm text-[var(--color-text-primary)] outline-none transition focus:border-[var(--color-primary)]"
            @change="loadStudents(($event.target as HTMLSelectElement).value)"
          >
            <option v-for="item in classes" :key="item.name" :value="item.name">
              {{ item.name }} · {{ item.student_count || 0 }}
            </option>
          </select>

          <input
            v-model="searchQuery"
            type="text"
            placeholder="搜索学员用户名..."
            class="rounded-xl border border-[var(--color-border-default)] bg-[var(--color-bg-base)] px-4 py-3 text-sm text-[var(--color-text-primary)] outline-none transition focus:border-[var(--color-primary)]"
          />
        </div>
      </div>
    </section>

    <div v-if="error" class="rounded-2xl border border-red-200 bg-red-50 px-5 py-4 text-sm text-red-600">
      {{ error }}
      <button type="button" class="ml-3 font-medium underline" @click="initialize">重试</button>
    </div>

    <section class="grid gap-6 xl:grid-cols-[360px_minmax(0,1fr)]">
      <aside class="rounded-2xl border border-[var(--color-border-default)] bg-[var(--color-bg-surface)] p-5 shadow-sm">
        <div class="flex items-center justify-between">
          <div>
            <h2 class="text-lg font-semibold text-[var(--color-text-primary)]">学员列表</h2>
            <p class="mt-1 text-sm text-[var(--color-text-secondary)]">
              {{ loadingClasses || loadingStudents ? '加载中...' : `共 ${filteredStudents.length} 名学员` }}
            </p>
          </div>
        </div>

        <div v-if="loadingClasses || loadingStudents" class="mt-5 space-y-3">
          <div v-for="index in 6" :key="index" class="h-16 animate-pulse rounded-xl bg-[var(--color-bg-base)]"></div>
        </div>

        <div v-else-if="filteredStudents.length === 0" class="mt-5 rounded-xl border border-dashed border-[var(--color-border-default)] px-4 py-8 text-center text-sm text-[var(--color-text-secondary)]">
          没有匹配到学员。
        </div>

        <div v-else class="mt-5 space-y-3">
          <button
            v-for="student in filteredStudents"
            :key="student.id"
            type="button"
            class="w-full rounded-xl border px-4 py-4 text-left transition"
            :class="student.id === selectedStudentId
              ? 'border-[var(--color-primary)] bg-[var(--color-primary)]/8'
              : 'border-[var(--color-border-default)] bg-[var(--color-bg-base)] hover:border-[var(--color-primary)]/50'"
            @click="loadStudentDetails(student.id)"
          >
            <div class="flex items-center justify-between gap-3">
              <div class="min-w-0">
                <p class="truncate font-medium text-[var(--color-text-primary)]">{{ student.name || student.username }}</p>
                <p class="mt-1 truncate text-sm text-[var(--color-text-secondary)]">@{{ student.username }}</p>
              </div>
              <span
                v-if="student.id === selectedStudentId"
                class="rounded-full bg-[var(--color-primary)]/12 px-3 py-1 text-xs font-medium text-[var(--color-primary)]"
              >
                已选择
              </span>
            </div>
          </button>
        </div>
      </aside>

      <StudentInsightPanel
        :student="selectedStudent"
        :progress="progress"
        :profile="skillProfile"
        :recommendations="recommendations"
        :loading="loadingDetails"
        empty-text="选择一名学员后，可在右侧查看完整画像。"
        @open-challenge="openChallenge"
      />
    </section>
  </div>
</template>
