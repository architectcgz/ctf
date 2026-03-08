<script setup lang="ts">
import { computed } from 'vue'
import { BookOpenCheck, GraduationCap, LayoutDashboard, Search, Users } from 'lucide-vue-next'

import type {
  MyProgressData,
  RecommendationItem,
  SkillProfileData,
  TeacherClassItem,
  TeacherStudentItem,
} from '@/api/contracts'
import AppCard from '@/components/common/AppCard.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import SectionCard from '@/components/common/SectionCard.vue'
import StudentInsightPanel from '@/components/teacher/StudentInsightPanel.vue'

const props = defineProps<{
  classes: TeacherClassItem[]
  selectedClassName: string
  searchQuery: string
  filteredStudents: TeacherStudentItem[]
  selectedStudentId: string
  selectedStudent: TeacherStudentItem | null
  loadingClasses: boolean
  loadingStudents: boolean
  loadingDetails: boolean
  error: string | null
  progress: MyProgressData | null
  skillProfile: SkillProfileData | null
  recommendations: RecommendationItem[]
}>()

const emit = defineEmits<{
  retry: []
  updateSearchQuery: [value: string]
  selectClass: [className: string]
  selectStudent: [studentId: string]
  openChallenge: [challengeId: string]
  openDashboard: []
  openReportExport: []
}>()

const focusStudents = computed(() => props.filteredStudents.slice(0, 3))
</script>

<template>
  <div class="space-y-6">
    <PageHeader
      eyebrow="Class Operations"
      title="班级运营台"
      description="这里不是普通列表页，而是围绕班级切换、学员筛选和个体画像联动单独设计的教师班级工作台。"
    >
      <ElButton plain @click="emit('openDashboard')">教学概览</ElButton>
      <ElButton type="primary" @click="emit('openReportExport')">导出报告</ElButton>
    </PageHeader>

    <section class="grid gap-4 xl:grid-cols-[1.08fr_0.92fr]">
      <AppCard
        variant="hero"
        accent="primary"
        eyebrow="Class Control"
        :title="selectedClassName ? `${selectedClassName} 的样本编排` : '先选择一个班级'"
        subtitle="这页专门服务老师做班级运营决策。左侧负责筛班级和找人，右侧保留学员详情与推荐任务，不再走通用管理页布局。"
      >
        <template #header>
          <span
            class="rounded-full border px-3 py-1 text-[11px] font-semibold uppercase tracking-[0.16em]"
            style="border-color: color-mix(in srgb, var(--color-primary) 18%, var(--color-border-default)); background-color: var(--color-primary-soft); color: var(--color-primary);"
          >
            {{ selectedClassName || '未选择班级' }}
          </span>
        </template>

        <div class="grid gap-3 md:grid-cols-2">
          <label class="space-y-2">
            <span class="text-[11px] font-semibold uppercase tracking-[0.18em] text-cyan-100/60">当前班级</span>
            <select
              :value="selectedClassName"
              class="w-full rounded-2xl border border-white/10 bg-white/6 px-4 py-3 text-sm text-white outline-none transition focus:border-cyan-300/60"
              @change="emit('selectClass', ($event.target as HTMLSelectElement).value)"
            >
              <option v-for="item in classes" :key="item.name" :value="item.name" class="text-black">
                {{ item.name }} · {{ item.student_count || 0 }}
              </option>
            </select>
          </label>

          <label class="space-y-2">
            <span class="text-[11px] font-semibold uppercase tracking-[0.18em] text-cyan-100/60">快速搜索</span>
            <div class="flex items-center gap-2 rounded-2xl border border-white/10 bg-white/6 px-4 py-3 text-white">
              <Search class="h-4 w-4 text-cyan-100/70" />
              <input
                :value="searchQuery"
                type="text"
                placeholder="搜索学员用户名..."
                class="w-full bg-transparent text-sm text-white outline-none placeholder:text-cyan-100/50"
                @input="emit('updateSearchQuery', ($event.target as HTMLInputElement).value)"
              />
            </div>
          </label>
        </div>
      </AppCard>

      <div class="grid gap-3 md:grid-cols-3 xl:grid-cols-1">
        <AppCard variant="metric" accent="primary" eyebrow="班级数量" :title="String(classes.length)" subtitle="当前教师权限下可访问的班级总数。">
          <template #header>
            <div class="flex h-11 w-11 items-center justify-center rounded-2xl border border-primary/20 bg-primary/12 text-primary">
              <BookOpenCheck class="h-5 w-5" />
            </div>
          </template>
        </AppCard>

        <AppCard variant="metric" accent="primary" eyebrow="匹配样本" :title="String(filteredStudents.length)" subtitle="当前班级和搜索条件下可见的学员数量。">
          <template #header>
            <div class="flex h-11 w-11 items-center justify-center rounded-2xl border border-primary/20 bg-primary/12 text-primary">
              <Users class="h-5 w-5" />
            </div>
          </template>
        </AppCard>

        <AppCard variant="metric" accent="warning" eyebrow="当前样本" :title="selectedStudent?.name || selectedStudent?.username || '未选择'" subtitle="选中后右侧会同步显示画像、进度和推荐任务。">
          <template #header>
            <div class="flex h-11 w-11 items-center justify-center rounded-2xl border border-amber-500/20 bg-amber-500/10 text-amber-300">
              <GraduationCap class="h-5 w-5" />
            </div>
          </template>
        </AppCard>
      </div>
    </section>

    <div v-if="error" class="rounded-2xl border border-red-200 bg-red-50 px-5 py-4 text-sm text-red-600">
      {{ error }}
      <button type="button" class="ml-3 font-medium underline" @click="emit('retry')">重试</button>
    </div>

    <section class="grid gap-6 xl:grid-cols-[0.94fr_1.06fr]">
      <div class="space-y-6">
        <SectionCard title="样本名单" subtitle="先找人，再决定是否进入右侧详情。">
          <div v-if="loadingClasses || loadingStudents" class="space-y-3">
            <div v-for="index in 6" :key="index" class="h-16 animate-pulse rounded-xl bg-[var(--color-bg-base)]" />
          </div>

          <div v-else-if="filteredStudents.length === 0" class="rounded-xl border border-dashed border-border px-4 py-8 text-center text-sm text-text-secondary">
            没有匹配到学员。
          </div>

          <div v-else class="grid gap-3">
            <AppCard
              v-for="student in filteredStudents"
              :key="student.id"
              as="button"
              variant="action"
              :accent="student.id === selectedStudentId ? 'primary' : 'neutral'"
              interactive
              class="w-full text-left"
              @click="emit('selectStudent', student.id)"
            >
              <div class="flex items-center justify-between gap-3">
                <div class="min-w-0">
                  <p class="truncate font-medium text-text-primary">{{ student.name || student.username }}</p>
                  <p class="mt-1 truncate text-sm text-text-secondary">@{{ student.username }}</p>
                </div>
                <span
                  v-if="student.id === selectedStudentId"
                  class="rounded-full bg-[var(--color-primary)]/12 px-3 py-1 text-xs font-medium text-[var(--color-primary)]"
                >
                  已选择
                </span>
              </div>
            </AppCard>
          </div>
        </SectionCard>

        <SectionCard title="重点关注" subtitle="保留三个最先进入视野的样本入口。">
          <div v-if="focusStudents.length === 0" class="rounded-xl border border-dashed border-border px-4 py-8 text-center text-sm text-text-secondary">
            当前没有可展示的重点样本。
          </div>

          <div v-else class="grid gap-3">
            <AppCard
              v-for="student in focusStudents"
              :key="student.id"
              as="button"
              variant="action"
              accent="primary"
              interactive
              class="text-left"
              @click="emit('selectStudent', student.id)"
            >
              <div class="flex items-center gap-3">
                <div class="flex h-11 w-11 items-center justify-center rounded-2xl bg-primary/12 text-primary">
                  <GraduationCap class="h-5 w-5" />
                </div>
                <div>
                  <div class="font-medium text-text-primary">{{ student.name || student.username }}</div>
                  <div class="mt-1 text-sm text-text-secondary">@{{ student.username }}</div>
                </div>
              </div>
              <LayoutDashboard class="h-4 w-4 text-primary" />
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
        empty-text="选择一名学员后，可在右侧查看完整画像。"
        @open-challenge="emit('openChallenge', $event)"
      />
    </section>
  </div>
</template>
