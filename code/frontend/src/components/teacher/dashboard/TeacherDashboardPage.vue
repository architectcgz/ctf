<script setup lang="ts">
import { computed } from 'vue'
import { ArrowRight, BookOpenCheck, GraduationCap, Radar, ShieldAlert, Users } from 'lucide-vue-next'

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
  students: TeacherStudentItem[]
  selectedClassName: string
  selectedClass: TeacherClassItem | null
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
  openReportExport: []
  openChallenge: [challengeId: string]
  selectClass: [className: string]
  selectStudent: [studentId: string]
}>()

const studentPreview = computed(() => props.students.slice(0, 6))
const highlightedStudents = computed(() => props.students.slice(0, 3))
const weakLabel = computed(() => (props.weakDimensions.length > 0 ? props.weakDimensions.join(' / ') : '暂无明显薄弱项'))
</script>

<template>
  <div class="space-y-6">
    <PageHeader
      eyebrow="Teacher Flight Deck"
      title="教学介入台"
      description="这里不再是通用仪表盘，而是围绕班级切换、样本筛查和干预线索单独设计的教师工作台。"
    >
      <ElButton plain @click="emit('openClassManagement')">班级管理</ElButton>
      <ElButton type="primary" @click="emit('openReportExport')">导出报告</ElButton>
    </PageHeader>

    <section class="grid gap-4 xl:grid-cols-[1.08fr_0.92fr]">
      <div class="overflow-hidden rounded-[30px] border border-cyan-400/20 bg-[radial-gradient(circle_at_top_left,rgba(34,211,238,0.18),transparent_42%),linear-gradient(145deg,rgba(15,23,42,0.96),rgba(6,78,99,0.82))] p-6 shadow-[0_24px_70px_var(--color-shadow-soft)]">
        <div class="flex flex-wrap items-center gap-2 text-[11px] font-semibold uppercase tracking-[0.24em] text-cyan-100/80">
          <span>Intervention Deck</span>
          <span class="rounded-full border border-white/10 bg-white/5 px-2 py-1">{{ selectedClassName || '未选择班级' }}</span>
        </div>
        <h2 class="mt-4 text-3xl font-semibold tracking-tight text-white">
          {{ selectedClassName ? `${selectedClassName} 的教学焦点` : '先选择一个班级' }}
        </h2>
        <p class="mt-3 max-w-2xl text-sm leading-7 text-cyan-50/78">
          {{
            selectedClassName
              ? '当前页面把班级切换、重点样本和教学建议收在同一个工作区，方便老师快速决定今天该先介入谁。'
              : '选择班级后，会同步刷新学员样本、进度、能力画像和推荐任务。'
          }}
        </p>

        <div class="mt-6 grid gap-3 md:grid-cols-3">
          <div class="rounded-[22px] border border-white/10 bg-white/6 px-4 py-4">
            <div class="text-[11px] uppercase tracking-[0.18em] text-cyan-100/60">班级人数</div>
            <div class="mt-2 text-2xl font-semibold text-white">{{ selectedClass?.student_count || students.length }}</div>
            <div class="mt-2 text-sm text-cyan-50/70">当前班级纳入视图的人数</div>
          </div>
          <div class="rounded-[22px] border border-white/10 bg-white/6 px-4 py-4">
            <div class="text-[11px] uppercase tracking-[0.18em] text-cyan-100/60">样本完成率</div>
            <div class="mt-2 text-2xl font-semibold text-white">{{ solvedRate }}%</div>
            <div class="mt-2 text-sm text-cyan-50/70">按当前选中学员样本计算</div>
          </div>
          <div class="rounded-[22px] border border-white/10 bg-white/6 px-4 py-4">
            <div class="text-[11px] uppercase tracking-[0.18em] text-cyan-100/60">薄弱维度</div>
            <div class="mt-2 text-lg font-semibold text-white">{{ weakLabel }}</div>
            <div class="mt-2 text-sm text-cyan-50/70">当前最值得介入的方向</div>
          </div>
        </div>
      </div>

      <div class="grid gap-3 md:grid-cols-3 xl:grid-cols-1">
        <article class="rounded-[24px] border border-border bg-surface/88 px-5 py-5 shadow-[0_18px_40px_var(--color-shadow-soft)]">
          <div class="flex items-start justify-between gap-3">
            <div>
              <div class="text-[11px] font-semibold uppercase tracking-[0.2em] text-text-muted">管理班级</div>
              <div class="mt-2 text-2xl font-semibold text-text-primary">{{ classes.length }}</div>
            </div>
            <div class="flex h-11 w-11 items-center justify-center rounded-2xl bg-primary/12 text-primary">
              <BookOpenCheck class="h-5 w-5" />
            </div>
          </div>
          <div class="mt-3 text-sm leading-6 text-text-secondary">教师权限下当前可访问的班级数量。</div>
        </article>

        <article class="rounded-[24px] border border-border bg-surface/88 px-5 py-5 shadow-[0_18px_40px_var(--color-shadow-soft)]">
          <div class="flex items-start justify-between gap-3">
            <div>
              <div class="text-[11px] font-semibold uppercase tracking-[0.2em] text-text-muted">当前样本</div>
              <div class="mt-2 text-2xl font-semibold text-text-primary">{{ selectedStudent?.name || selectedStudent?.username || '未选中' }}</div>
            </div>
            <div class="flex h-11 w-11 items-center justify-center rounded-2xl bg-primary/12 text-primary">
              <Users class="h-5 w-5" />
            </div>
          </div>
          <div class="mt-3 text-sm leading-6 text-text-secondary">右侧会针对当前样本展示能力画像和训练建议。</div>
        </article>

        <article class="rounded-[24px] border border-border bg-surface/88 px-5 py-5 shadow-[0_18px_40px_var(--color-shadow-soft)]">
          <div class="flex items-start justify-between gap-3">
            <div>
              <div class="text-[11px] font-semibold uppercase tracking-[0.2em] text-text-muted">教学干预</div>
              <div class="mt-2 text-xl font-semibold text-text-primary">{{ weakLabel }}</div>
            </div>
            <div class="flex h-11 w-11 items-center justify-center rounded-2xl bg-primary/12 text-primary">
              <ShieldAlert class="h-5 w-5" />
            </div>
          </div>
          <div class="mt-3 text-sm leading-6 text-text-secondary">系统会按薄弱维度动态收敛最适合老师介入的方向。</div>
        </article>
      </div>
    </section>

    <div v-if="error" class="rounded-2xl border border-red-200 bg-red-50 px-5 py-4 text-sm text-red-600">
      {{ error }}
      <button type="button" class="ml-3 font-medium underline" @click="emit('retry')">重试</button>
    </div>

    <section class="grid gap-6 xl:grid-cols-[0.94fr_1.06fr]">
      <div class="space-y-6">
        <SectionCard title="班级走廊" subtitle="先选班级，再从样本里挑需要介入的学员。">
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

          <div class="mt-6 grid gap-3">
            <div class="flex items-center justify-between">
              <h3 class="text-sm font-semibold text-text-primary">班级样本</h3>
              <span class="text-xs text-text-secondary">{{ loadingStudents ? '加载中...' : `${students.length} 名学员` }}</span>
            </div>

            <div v-if="loadingClasses || loadingStudents" class="space-y-3">
              <div v-for="index in 4" :key="index" class="h-16 animate-pulse rounded-xl bg-[var(--color-bg-base)]" />
            </div>

            <div v-else-if="studentPreview.length === 0" class="rounded-xl border border-dashed border-border px-4 py-8 text-center text-sm text-text-secondary">
              当前班级暂无学员。
            </div>

            <div v-else class="grid gap-3">
              <button
                v-for="student in studentPreview"
                :key="student.id"
                type="button"
                class="rounded-[24px] border px-4 py-4 text-left transition"
                :class="student.id === selectedStudentId
                  ? 'border-[var(--color-primary)] bg-[var(--color-primary)]/8'
                  : 'border-[var(--color-border-default)] bg-[var(--color-bg-base)] hover:border-[var(--color-primary)]/50'"
                @click="emit('selectStudent', student.id)"
              >
                <div class="flex items-center justify-between gap-3">
                  <div>
                    <p class="font-medium text-text-primary">{{ student.name || student.username }}</p>
                    <p class="mt-1 text-sm text-text-secondary">@{{ student.username }}</p>
                  </div>
                  <span
                    v-if="student.id === selectedStudentId"
                    class="rounded-full bg-[var(--color-primary)]/12 px-3 py-1 text-xs font-medium text-[var(--color-primary)]"
                  >
                    聚焦中
                  </span>
                </div>
              </button>
            </div>
          </div>
        </SectionCard>

        <SectionCard title="优先关注样本" subtitle="这里保留最近最值得老师先看的三位。">
          <div v-if="highlightedStudents.length === 0" class="rounded-xl border border-dashed border-border px-4 py-8 text-center text-sm text-text-secondary">
            当前还没有可用学员样本。
          </div>

          <div v-else class="grid gap-3">
            <button
              v-for="student in highlightedStudents"
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
              <span class="inline-flex items-center gap-1 text-sm font-medium text-primary">
                查看
                <ArrowRight class="h-4 w-4" />
              </span>
            </button>
          </div>
        </SectionCard>

        <SectionCard title="教学信号" subtitle="从当前样本快速读取干预优先级。">
          <div class="grid gap-3 md:grid-cols-3">
            <div class="rounded-[22px] border border-border bg-base/70 px-4 py-4">
              <div class="flex items-center gap-2 text-sm font-medium text-text-primary">
                <Users class="h-4 w-4 text-sky-300" />
                当前样本
              </div>
              <div class="mt-3 text-2xl font-semibold text-text-primary">{{ selectedStudent ? 1 : 0 }}</div>
              <div class="mt-2 text-sm text-text-secondary">右侧面板会跟随当前选中学员刷新。</div>
            </div>
            <div class="rounded-[22px] border border-border bg-base/70 px-4 py-4">
              <div class="flex items-center gap-2 text-sm font-medium text-text-primary">
                <Radar class="h-4 w-4 text-amber-300" />
                完成率
              </div>
              <div class="mt-3 text-2xl font-semibold text-text-primary">{{ solvedRate }}%</div>
              <div class="mt-2 text-sm text-text-secondary">基于当前选中样本的完成情况。</div>
            </div>
            <div class="rounded-[22px] border border-border bg-base/70 px-4 py-4">
              <div class="flex items-center gap-2 text-sm font-medium text-text-primary">
                <ShieldAlert class="h-4 w-4 text-fuchsia-300" />
                推荐任务
              </div>
              <div class="mt-3 text-2xl font-semibold text-text-primary">{{ recommendations.length }}</div>
              <div class="mt-2 text-sm text-text-secondary">可以直接布置给当前学员的补强题目数。</div>
            </div>
          </div>
        </SectionCard>
      </div>

      <StudentInsightPanel
        :student="selectedStudent"
        :progress="progress"
        :profile="skillProfile"
        :recommendations="recommendations"
        :loading="loadingDetails"
        empty-text="请先从左侧选择学员，再查看教学建议。"
        @open-challenge="emit('openChallenge', $event)"
      />
    </section>
  </div>
</template>
