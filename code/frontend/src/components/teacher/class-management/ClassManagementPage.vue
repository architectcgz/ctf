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
      <div class="overflow-hidden rounded-[30px] border border-cyan-400/20 bg-[radial-gradient(circle_at_top_left,rgba(34,211,238,0.16),transparent_42%),linear-gradient(145deg,rgba(15,23,42,0.96),rgba(8,47,73,0.84))] p-6 shadow-[0_24px_70px_var(--color-shadow-soft)]">
        <div class="flex flex-wrap items-center gap-2 text-[11px] font-semibold uppercase tracking-[0.24em] text-cyan-100/80">
          <span>Class Control</span>
          <span class="rounded-full border border-white/10 bg-white/5 px-2 py-1">{{ selectedClassName || '未选择班级' }}</span>
        </div>
        <h2 class="mt-4 text-3xl font-semibold tracking-tight text-white">
          {{ selectedClassName ? `${selectedClassName} 的样本编排` : '先选择一个班级' }}
        </h2>
        <p class="mt-3 max-w-2xl text-sm leading-7 text-cyan-50/78">
          这页专门服务老师做班级运营决策。左侧负责筛班级和找人，右侧保留学员详情与推荐任务，不再走通用管理页布局。
        </p>

        <div class="mt-6 grid gap-3 md:grid-cols-2">
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
      </div>

      <div class="grid gap-3 md:grid-cols-3 xl:grid-cols-1">
        <article class="rounded-[24px] border border-border bg-surface/88 px-5 py-5 shadow-[0_18px_40px_var(--color-shadow-soft)]">
          <div class="flex items-start justify-between gap-3">
            <div>
              <div class="text-[11px] font-semibold uppercase tracking-[0.2em] text-text-muted">班级数量</div>
              <div class="mt-2 text-2xl font-semibold text-text-primary">{{ classes.length }}</div>
            </div>
            <div class="flex h-11 w-11 items-center justify-center rounded-2xl bg-primary/12 text-primary">
              <BookOpenCheck class="h-5 w-5" />
            </div>
          </div>
          <div class="mt-3 text-sm leading-6 text-text-secondary">当前教师权限下可访问的班级总数。</div>
        </article>

        <article class="rounded-[24px] border border-border bg-surface/88 px-5 py-5 shadow-[0_18px_40px_var(--color-shadow-soft)]">
          <div class="flex items-start justify-between gap-3">
            <div>
              <div class="text-[11px] font-semibold uppercase tracking-[0.2em] text-text-muted">匹配样本</div>
              <div class="mt-2 text-2xl font-semibold text-text-primary">{{ filteredStudents.length }}</div>
            </div>
            <div class="flex h-11 w-11 items-center justify-center rounded-2xl bg-primary/12 text-primary">
              <Users class="h-5 w-5" />
            </div>
          </div>
          <div class="mt-3 text-sm leading-6 text-text-secondary">当前班级和搜索条件下可见的学员数量。</div>
        </article>

        <article class="rounded-[24px] border border-border bg-surface/88 px-5 py-5 shadow-[0_18px_40px_var(--color-shadow-soft)]">
          <div class="flex items-start justify-between gap-3">
            <div>
              <div class="text-[11px] font-semibold uppercase tracking-[0.2em] text-text-muted">当前样本</div>
              <div class="mt-2 text-2xl font-semibold text-text-primary">{{ selectedStudent?.name || selectedStudent?.username || '未选择' }}</div>
            </div>
            <div class="flex h-11 w-11 items-center justify-center rounded-2xl bg-primary/12 text-primary">
              <GraduationCap class="h-5 w-5" />
            </div>
          </div>
          <div class="mt-3 text-sm leading-6 text-text-secondary">选中后右侧会同步显示画像、进度和推荐任务。</div>
        </article>
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
            <button
              v-for="student in filteredStudents"
              :key="student.id"
              type="button"
              class="w-full rounded-[24px] border px-4 py-4 text-left transition"
              :class="student.id === selectedStudentId
                ? 'border-[var(--color-primary)] bg-[var(--color-primary)]/8'
                : 'border-[var(--color-border-default)] bg-[var(--color-bg-base)] hover:border-[var(--color-primary)]/50'"
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
            </button>
          </div>
        </SectionCard>

        <SectionCard title="重点关注" subtitle="保留三个最先进入视野的样本入口。">
          <div v-if="focusStudents.length === 0" class="rounded-xl border border-dashed border-border px-4 py-8 text-center text-sm text-text-secondary">
            当前没有可展示的重点样本。
          </div>

          <div v-else class="grid gap-3">
            <button
              v-for="student in focusStudents"
              :key="student.id"
              type="button"
              class="flex items-center justify-between gap-3 rounded-[24px] border border-border bg-[linear-gradient(180deg,rgba(15,23,42,0.88),rgba(8,15,32,0.72))] px-4 py-4 text-left transition hover:border-primary/60"
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
            </button>
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
