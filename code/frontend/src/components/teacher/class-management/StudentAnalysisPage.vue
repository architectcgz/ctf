<script setup lang="ts">
import { ArrowLeftRight, FileDown, GraduationCap, Users } from 'lucide-vue-next'
import { ref } from 'vue'

import type {
  MyProgressData,
  RecommendationItem,
  SkillProfileData,
  TeacherEvidenceData,
  TeacherClassItem,
  TeacherManualReviewSubmissionDetailData,
  TeacherManualReviewSubmissionItemData,
  TeacherSubmissionWriteupItemData,
  TeacherStudentItem,
  TimelineEvent,
} from '@/api/contracts'
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
  timeline: TimelineEvent[]
  evidence: TeacherEvidenceData | null
  writeupSubmissions: TeacherSubmissionWriteupItemData[]
  manualReviewSubmissions: TeacherManualReviewSubmissionItemData[]
  activeManualReview: TeacherManualReviewSubmissionDetailData | null
  manualReviewLoading: boolean
  manualReviewSaving: boolean
  solvedRate: number
  weakDimensions: string[]
}>()

const emit = defineEmits<{
  retry: []
  openClassManagement: []
  openClassStudents: []
  openReportExport: []
  openReviewArchive: []
  exportReviewArchive: []
  selectClass: [className: string]
  selectStudent: [studentId: string]
  openChallenge: [challengeId: string]
  openManualReview: [submissionId: string]
  moderateWriteup: [
    payload: { submissionId: string; action: 'recommend' | 'unrecommend' | 'hide' | 'restore' },
  ]
  reviewManualReview: [
    payload: {
      submissionId: string
      reviewStatus: 'approved' | 'rejected'
      reviewComment?: string
    },
  ]
}>()

type WorkspaceTab = 'overview' | 'recommendations' | 'writeups' | 'manual-review' | 'evidence'

interface WorkspaceTabItem {
  key: WorkspaceTab
  label: string
  buttonId: string
  panelId: string
}

const workspaceTabs: WorkspaceTabItem[] = [
  {
    key: 'overview',
    label: '学员画像',
    buttonId: 'student-tab-overview',
    panelId: 'student-overview',
  },
  {
    key: 'recommendations',
    label: '推荐任务',
    buttonId: 'student-tab-recommendations',
    panelId: 'student-recommendations',
  },
  {
    key: 'writeups',
    label: '社区题解',
    buttonId: 'student-tab-writeups',
    panelId: 'student-writeups',
  },
  {
    key: 'manual-review',
    label: '人工审核',
    buttonId: 'student-tab-manual-review',
    panelId: 'student-manual-review',
  },
  {
    key: 'evidence',
    label: '证据链',
    buttonId: 'student-tab-evidence',
    panelId: 'student-evidence',
  },
]

const activeTab = ref<WorkspaceTab>('overview')
const tabButtonRefs: Partial<Record<WorkspaceTab, HTMLButtonElement | null>> = {}

function setTabButtonRef(tab: WorkspaceTab, element: HTMLButtonElement | null): void {
  tabButtonRefs[tab] = element
}

function selectTab(tab: WorkspaceTab): void {
  activeTab.value = tab
}

function focusTab(tab: WorkspaceTab): void {
  tabButtonRefs[tab]?.focus()
}

function handleTabKeydown(event: KeyboardEvent, index: number): void {
  if (
    event.key !== 'ArrowRight' &&
    event.key !== 'ArrowLeft' &&
    event.key !== 'Home' &&
    event.key !== 'End'
  ) {
    return
  }

  event.preventDefault()

  if (event.key === 'Home') {
    selectTab(workspaceTabs[0].key)
    focusTab(workspaceTabs[0].key)
    return
  }

  if (event.key === 'End') {
    const lastTab = workspaceTabs[workspaceTabs.length - 1]
    selectTab(lastTab.key)
    focusTab(lastTab.key)
    return
  }

  const direction = event.key === 'ArrowRight' ? 1 : -1
  const nextIndex = (index + direction + workspaceTabs.length) % workspaceTabs.length
  const nextTab = workspaceTabs[nextIndex]
  selectTab(nextTab.key)
  focusTab(nextTab.key)
}
</script>

<template>
  <div class="workspace-shell journal-eyebrow-text">
    <header class="workspace-topbar">
      <div class="topbar-leading">
        <span class="workspace-overline">Student Workspace</span>
        <span class="class-chip">{{ selectedClassName || '未选择班级' }}</span>
      </div>
    </header>

    <nav class="top-tabs" role="tablist" aria-label="学员分析标签页">
      <button
        v-for="(tab, index) in workspaceTabs"
        :id="tab.buttonId"
        :key="tab.key"
        :ref="(element) => setTabButtonRef(tab.key, element as HTMLButtonElement | null)"
        class="top-tab"
        :class="{ active: activeTab === tab.key }"
        type="button"
        role="tab"
        :tabindex="activeTab === tab.key ? 0 : -1"
        :aria-selected="activeTab === tab.key ? 'true' : 'false'"
        :aria-controls="tab.panelId"
        @click="selectTab(tab.key)"
        @keydown="handleTabKeydown($event, index)"
      >
        {{ tab.label }}
      </button>
    </nav>

    <div class="workspace-grid">
      <main class="content-pane">
        <header class="teacher-topbar">
          <div class="teacher-heading">
            <div class="teacher-eyebrow-row">
              <div class="journal-eyebrow">Student Analysis</div>
              <span class="teacher-student-chip">@{{ selectedStudent?.username || '未选择' }}</span>
            </div>
            <h1 class="teacher-title">
              {{ selectedStudent?.name || selectedStudent?.username || '学员分析' }}
            </h1>
            <p class="teacher-copy">
              用顶部 tabs
              切换学员画像、推荐任务、题解状态、人工审核和证据链，不再把所有内容堆在同一屏。
            </p>
          </div>

          <div class="teacher-actions" role="group" aria-label="学员分析快捷操作">
            <button
              type="button"
              class="teacher-btn teacher-btn--ghost"
              @click="emit('openClassStudents')"
            >
              返回学生列表
            </button>
            <button
              type="button"
              class="teacher-btn teacher-btn--ghost"
              @click="emit('openReviewArchive')"
            >
              完整复盘页
            </button>
            <button
              type="button"
              class="teacher-btn teacher-btn--primary"
              @click="emit('exportReviewArchive')"
            >
              导出复盘归档
            </button>
          </div>
        </header>

        <div v-if="error" class="workspace-alert" role="alert" aria-live="polite">
          <div class="workspace-alert-title">学员分析加载失败</div>
          <div class="workspace-alert-copy">{{ error }}</div>
          <div class="workspace-alert-actions">
            <button type="button" class="quick-action quick-action--compact" @click="emit('retry')">
              <span>重试加载</span>
              <span>→</span>
            </button>
          </div>
        </div>

        <section class="summary-strip">
          <article class="summary-card">
            <div class="summary-card__label">当前学员</div>
            <div class="summary-card__value">{{ selectedStudent?.username || '未选择' }}</div>
            <div class="summary-card__hint">当前聚焦的学生对象</div>
          </article>
          <article class="summary-card">
            <div class="summary-card__label">完成率</div>
            <div class="summary-card__value">{{ solvedRate }}%</div>
            <div class="summary-card__hint">基于当前学员训练数据计算</div>
          </article>
          <article class="summary-card">
            <div class="summary-card__label">薄弱维度</div>
            <div class="summary-card__value">{{ weakDimensions[0] || '暂无' }}</div>
            <div class="summary-card__hint">当前最需要补强的方向</div>
          </article>
          <article class="summary-card">
            <div class="summary-card__label">可切换学生</div>
            <div class="summary-card__value">{{ students.length }}</div>
            <div class="summary-card__hint">当前班级可切换的学生数量</div>
          </article>
        </section>

        <section
          id="student-overview"
          class="tab-panel section"
          :class="{ active: activeTab === 'overview' }"
          role="tabpanel"
          aria-labelledby="student-tab-overview"
          :aria-hidden="activeTab === 'overview' ? 'false' : 'true'"
          v-show="activeTab === 'overview'"
        >
          <StudentInsightPanel
            active-section="overview"
            :student="selectedStudent"
            :progress="progress"
            :profile="skillProfile"
            :recommendations="recommendations"
            :timeline="timeline"
            :evidence="evidence"
            :writeup-submissions="writeupSubmissions"
            :manual-review-submissions="manualReviewSubmissions"
            :active-manual-review="activeManualReview"
            :manual-review-loading="manualReviewLoading"
            :manual-review-saving="manualReviewSaving"
            :loading="loadingDetails"
            empty-text="请先从右侧目录选择一名学生。"
            @open-challenge="emit('openChallenge', $event)"
            @open-manual-review="emit('openManualReview', $event)"
            @moderate-writeup="emit('moderateWriteup', $event)"
            @review-manual-review="emit('reviewManualReview', $event)"
          />
        </section>

        <section
          id="student-recommendations"
          class="tab-panel section"
          :class="{ active: activeTab === 'recommendations' }"
          role="tabpanel"
          aria-labelledby="student-tab-recommendations"
          :aria-hidden="activeTab === 'recommendations' ? 'false' : 'true'"
          v-show="activeTab === 'recommendations'"
        >
          <StudentInsightPanel
            active-section="recommendations"
            :student="selectedStudent"
            :progress="progress"
            :profile="skillProfile"
            :recommendations="recommendations"
            :timeline="timeline"
            :evidence="evidence"
            :writeup-submissions="writeupSubmissions"
            :manual-review-submissions="manualReviewSubmissions"
            :active-manual-review="activeManualReview"
            :manual-review-loading="manualReviewLoading"
            :manual-review-saving="manualReviewSaving"
            :loading="loadingDetails"
            empty-text="请先从右侧目录选择一名学生。"
            @open-challenge="emit('openChallenge', $event)"
            @open-manual-review="emit('openManualReview', $event)"
            @moderate-writeup="emit('moderateWriteup', $event)"
            @review-manual-review="emit('reviewManualReview', $event)"
          />
        </section>

        <section
          id="student-writeups"
          class="tab-panel section"
          :class="{ active: activeTab === 'writeups' }"
          role="tabpanel"
          aria-labelledby="student-tab-writeups"
          :aria-hidden="activeTab === 'writeups' ? 'false' : 'true'"
          v-show="activeTab === 'writeups'"
        >
          <StudentInsightPanel
            active-section="writeups"
            :student="selectedStudent"
            :progress="progress"
            :profile="skillProfile"
            :recommendations="recommendations"
            :timeline="timeline"
            :evidence="evidence"
            :writeup-submissions="writeupSubmissions"
            :manual-review-submissions="manualReviewSubmissions"
            :active-manual-review="activeManualReview"
            :manual-review-loading="manualReviewLoading"
            :manual-review-saving="manualReviewSaving"
            :loading="loadingDetails"
            empty-text="请先从右侧目录选择一名学生。"
            @open-challenge="emit('openChallenge', $event)"
            @open-manual-review="emit('openManualReview', $event)"
            @moderate-writeup="emit('moderateWriteup', $event)"
            @review-manual-review="emit('reviewManualReview', $event)"
          />
        </section>

        <section
          id="student-manual-review"
          class="tab-panel section"
          :class="{ active: activeTab === 'manual-review' }"
          role="tabpanel"
          aria-labelledby="student-tab-manual-review"
          :aria-hidden="activeTab === 'manual-review' ? 'false' : 'true'"
          v-show="activeTab === 'manual-review'"
        >
          <StudentInsightPanel
            active-section="manual-review"
            :student="selectedStudent"
            :progress="progress"
            :profile="skillProfile"
            :recommendations="recommendations"
            :timeline="timeline"
            :evidence="evidence"
            :writeup-submissions="writeupSubmissions"
            :manual-review-submissions="manualReviewSubmissions"
            :active-manual-review="activeManualReview"
            :manual-review-loading="manualReviewLoading"
            :manual-review-saving="manualReviewSaving"
            :loading="loadingDetails"
            empty-text="请先从右侧目录选择一名学生。"
            @open-challenge="emit('openChallenge', $event)"
            @open-manual-review="emit('openManualReview', $event)"
            @moderate-writeup="emit('moderateWriteup', $event)"
            @review-manual-review="emit('reviewManualReview', $event)"
          />
        </section>

        <section
          id="student-evidence"
          class="tab-panel section"
          :class="{ active: activeTab === 'evidence' }"
          role="tabpanel"
          aria-labelledby="student-tab-evidence"
          :aria-hidden="activeTab === 'evidence' ? 'false' : 'true'"
          v-show="activeTab === 'evidence'"
        >
          <StudentInsightPanel
            active-section="evidence"
            :student="selectedStudent"
            :progress="progress"
            :profile="skillProfile"
            :recommendations="recommendations"
            :timeline="timeline"
            :evidence="evidence"
            :writeup-submissions="writeupSubmissions"
            :manual-review-submissions="manualReviewSubmissions"
            :active-manual-review="activeManualReview"
            :manual-review-loading="manualReviewLoading"
            :manual-review-saving="manualReviewSaving"
            :loading="loadingDetails"
            empty-text="请先从右侧目录选择一名学生。"
            @open-challenge="emit('openChallenge', $event)"
            @open-manual-review="emit('openManualReview', $event)"
            @moderate-writeup="emit('moderateWriteup', $event)"
            @review-manual-review="emit('reviewManualReview', $event)"
          />
        </section>
      </main>

      <aside class="context-rail">
        <SectionCard title="班级与学生切换" subtitle="先定班级，再切换当前分析的学生。">
          <div class="rail-stack">
            <div class="class-switch-list" role="group" aria-label="班级切换">
              <button
                v-for="item in classes"
                :key="item.name"
                type="button"
                class="class-switch"
                :class="{ active: item.name === selectedClassName }"
                @click="emit('selectClass', item.name)"
              >
                <span>{{ item.name }}</span>
                <span>{{ item.student_count || 0 }} 人</span>
              </button>
            </div>

            <div v-if="loadingClasses || loadingStudents" class="student-directory-skeleton">
              <div v-for="index in 4" :key="index" class="student-directory-skeleton__item" />
            </div>

            <div v-else class="student-directory" role="list" aria-label="学生目录">
              <div class="student-directory-head" aria-hidden="true">
                <span>学生</span>
                <span>状态</span>
              </div>

              <button
                v-for="student in students"
                :key="student.id"
                type="button"
                class="student-directory-row"
                :class="{ active: student.id === selectedStudentId }"
                role="listitem"
                @click="emit('selectStudent', student.id)"
              >
                <div class="student-directory-main">
                  <div class="student-directory-avatar">
                    <GraduationCap class="h-4 w-4" />
                  </div>
                  <div class="student-directory-copy">
                    <div class="student-directory-name">{{ student.name || student.username }}</div>
                    <div class="student-directory-meta">@{{ student.username }}</div>
                  </div>
                </div>
                <span class="student-directory-state">
                  {{ student.id === selectedStudentId ? '当前' : '切换' }}
                </span>
              </button>
            </div>
          </div>
        </SectionCard>

        <SectionCard title="快捷操作" subtitle="保留导航与导出入口，复盘页继续走独立路由。">
          <div class="rail-stack" role="group" aria-label="学员分析辅助操作">
            <button type="button" class="quick-action" @click="emit('openClassManagement')">
              <span class="quick-action__main">
                <span class="quick-action__icon quick-action__icon--neutral">
                  <Users class="h-4 w-4" />
                </span>
                <span>
                  <strong>班级管理</strong>
                  <small>返回教师班级管理入口。</small>
                </span>
              </span>
              <span>→</span>
            </button>

            <button type="button" class="quick-action" @click="emit('openReportExport')">
              <span class="quick-action__main">
                <span class="quick-action__icon quick-action__icon--warning">
                  <FileDown class="h-4 w-4" />
                </span>
                <span>
                  <strong>导出班级报告</strong>
                  <small>从当前教师路径继续进入报告导出。</small>
                </span>
              </span>
              <span>→</span>
            </button>

            <button
              type="button"
              class="quick-action quick-action--primary"
              @click="emit('openReviewArchive')"
            >
              <span class="quick-action__main">
                <span class="quick-action__icon quick-action__icon--primary">
                  <ArrowLeftRight class="h-4 w-4" />
                </span>
                <span>
                  <strong>打开完整复盘页</strong>
                  <small>查看摘要、证据链和评阅记录。</small>
                </span>
              </span>
              <span>→</span>
            </button>
          </div>
        </SectionCard>
      </aside>
    </div>
  </div>
</template>

<style scoped>
.workspace-shell {
  --journal-ink: var(--color-text-primary);
  --journal-muted: var(--color-text-secondary);
  --journal-border: color-mix(in srgb, var(--color-border-default) 82%, transparent);
  --journal-surface: color-mix(in srgb, var(--color-bg-surface) 88%, var(--color-bg-base));
  --journal-surface-subtle: color-mix(in srgb, var(--color-bg-surface) 74%, var(--color-bg-base));
  --journal-accent: var(--color-primary);
  --journal-accent-strong: color-mix(in srgb, var(--color-primary-hover) 82%, var(--journal-ink));
  --journal-eyebrow-spacing: 0.15em;
  --journal-eyebrow-color: var(--journal-accent-strong);
  --teacher-card-border: color-mix(in srgb, var(--journal-border) 76%, transparent);
  --teacher-control-border: color-mix(in srgb, var(--journal-border) 78%, transparent);
  --teacher-divider: color-mix(in srgb, var(--journal-border) 86%, transparent);
  --workspace-page: color-mix(in srgb, var(--color-bg-base) 94%, var(--color-bg-surface));
  --workspace-shell-bg: color-mix(in srgb, var(--color-bg-surface) 92%, var(--color-bg-base));
  --workspace-panel: color-mix(in srgb, var(--color-bg-surface) 90%, var(--color-bg-base));
  --workspace-line-soft: color-mix(in srgb, var(--color-text-primary) 10%, transparent);
  --workspace-faint: color-mix(in srgb, var(--color-text-secondary) 88%, var(--color-bg-base));
  --workspace-brand: color-mix(in srgb, var(--color-primary) 86%, var(--journal-ink));
  --workspace-brand-ink: color-mix(in srgb, var(--color-primary) 74%, var(--journal-ink));
  --workspace-brand-soft: color-mix(in srgb, var(--color-primary) 10%, transparent);
  --workspace-shadow-shell: 0 24px 84px
    color-mix(in srgb, var(--color-shadow-soft) 58%, transparent);
  --workspace-shadow-panel: 0 14px 34px
    color-mix(in srgb, var(--color-shadow-soft) 42%, transparent);
  --workspace-radius-xl: 28px;
  --workspace-radius-lg: 18px;
}

.workspace-grid {
  grid-template-columns: minmax(0, 1fr) minmax(19rem, 24rem);
}

.context-rail {
  min-width: 0;
  padding: 28px 28px 28px 0;
}

.teacher-eyebrow-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.65rem;
}

.teacher-student-chip {
  display: inline-flex;
  align-items: center;
  min-height: 1.85rem;
  padding: 0 0.75rem;
  border-radius: 999px;
  border: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
  background: color-mix(in srgb, var(--journal-surface) 88%, transparent);
  font-size: 0.78rem;
  font-weight: 600;
  color: var(--journal-muted);
}

.workspace-alert {
  margin-bottom: 1.5rem;
  border: 1px solid var(--workspace-line-soft);
  border-radius: var(--workspace-radius-lg);
  background: color-mix(in srgb, var(--workspace-panel) 88%, transparent);
  box-shadow: var(--workspace-shadow-panel);
  padding: 1rem 1.1rem;
}

.workspace-alert-title {
  font-size: 0.92rem;
  font-weight: 700;
  color: var(--journal-ink);
}

.workspace-alert-copy {
  margin-top: 0.45rem;
  font-size: 0.84rem;
  line-height: 1.65;
  color: var(--journal-muted);
}

.workspace-alert-actions {
  margin-top: 0.85rem;
}

.summary-strip {
  display: grid;
  gap: 0.9rem;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  margin-bottom: 1.5rem;
}

.summary-card {
  min-width: 0;
  border: 1px solid var(--teacher-card-border);
  border-radius: 16px;
  background: color-mix(in srgb, var(--journal-surface) 92%, var(--color-bg-base));
  padding: 0.95rem 1rem;
  box-shadow: 0 10px 24px var(--color-shadow-soft);
}

.summary-card__label {
  font-size: 0.7rem;
  font-weight: 700;
  letter-spacing: 0.15em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.summary-card__value {
  margin-top: 0.45rem;
  font-size: 1.15rem;
  font-weight: 700;
  line-height: 1.4;
  color: var(--journal-ink);
}

.summary-card__hint {
  margin-top: 0.45rem;
  font-size: 0.8rem;
  line-height: 1.55;
  color: var(--journal-muted);
}

.rail-stack {
  display: grid;
  gap: 0.95rem;
}

.class-switch-list {
  display: grid;
  gap: 0.6rem;
}

.class-switch {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  min-height: 2.75rem;
  padding: 0.72rem 0.9rem;
  border: 1px solid var(--teacher-control-border);
  border-radius: 0.9rem;
  background: color-mix(in srgb, var(--journal-surface) 92%, var(--color-bg-base));
  font-size: 0.86rem;
  font-weight: 600;
  color: var(--journal-ink);
  transition:
    border-color 160ms ease,
    background 160ms ease,
    color 160ms ease;
}

.class-switch:hover,
.class-switch:focus-visible,
.class-switch.active {
  border-color: color-mix(in srgb, var(--journal-accent) 40%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 8%, var(--journal-surface));
  color: var(--journal-accent-strong);
  outline: none;
}

.student-directory {
  display: grid;
}

.student-directory-head,
.student-directory-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: 0.75rem;
  align-items: center;
}

.student-directory-head {
  padding: 0 0.15rem 0.55rem;
  border-bottom: 1px solid var(--teacher-divider);
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.student-directory-row {
  padding: 0.9rem 0.1rem;
  border-bottom: 1px solid var(--teacher-divider);
  background: transparent;
  text-align: left;
  transition: background 160ms ease;
}

.student-directory-row:hover,
.student-directory-row:focus-visible,
.student-directory-row.active {
  background: color-mix(in srgb, var(--journal-accent) 6%, transparent);
  outline: none;
}

.student-directory-main {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  min-width: 0;
}

.student-directory-avatar {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 2.25rem;
  height: 2.25rem;
  border-radius: 0.9rem;
  border: 1px solid color-mix(in srgb, var(--journal-accent) 16%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 10%, transparent);
  color: var(--journal-accent);
  flex-shrink: 0;
}

.student-directory-copy {
  min-width: 0;
}

.student-directory-name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 0.92rem;
  font-weight: 600;
  color: var(--journal-ink);
}

.student-directory-meta {
  margin-top: 0.2rem;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 0.8rem;
  color: var(--journal-muted);
}

.student-directory-state {
  font-size: 0.78rem;
  font-weight: 600;
  color: var(--workspace-brand-ink);
}

.student-directory-skeleton {
  display: grid;
  gap: 0.75rem;
}

.student-directory-skeleton__item {
  height: 3.5rem;
  border-radius: 0.95rem;
  background: linear-gradient(
    90deg,
    color-mix(in srgb, var(--journal-border) 80%, transparent),
    color-mix(in srgb, var(--journal-surface-subtle) 96%, var(--color-bg-base))
  );
  animation: pulse 1.35s ease-in-out infinite;
}

.quick-action {
  display: inline-flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  min-height: 52px;
  padding: 0.95rem 1rem;
  border: 1px solid var(--workspace-line-soft);
  border-radius: 14px;
  background: color-mix(in srgb, var(--workspace-panel) 82%, transparent);
  color: var(--journal-ink);
  cursor: pointer;
  transition:
    border-color 160ms ease,
    background 160ms ease,
    color 160ms ease;
}

.quick-action:hover,
.quick-action:focus-visible {
  border-color: color-mix(in srgb, var(--workspace-brand) 34%, transparent);
  background: color-mix(in srgb, var(--workspace-brand) 8%, var(--workspace-panel));
  color: var(--workspace-brand-ink);
  outline: none;
}

.quick-action--compact {
  min-height: 42px;
}

.quick-action--primary {
  border-color: color-mix(in srgb, var(--journal-accent) 18%, transparent);
}

.quick-action__main {
  display: flex;
  align-items: center;
  gap: 0.85rem;
  min-width: 0;
}

.quick-action__main span:last-child {
  display: grid;
  gap: 0.2rem;
  min-width: 0;
}

.quick-action__main strong {
  font-size: 0.9rem;
  font-weight: 600;
  color: inherit;
}

.quick-action__main small {
  font-size: 0.78rem;
  line-height: 1.55;
  color: var(--journal-muted);
}

.quick-action__icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 2.2rem;
  height: 2.2rem;
  border-radius: 0.85rem;
  flex-shrink: 0;
}

.quick-action__icon--neutral {
  border: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
  background: color-mix(in srgb, var(--journal-surface) 88%, transparent);
}

.quick-action__icon--warning {
  border: 1px solid color-mix(in srgb, var(--color-warning) 18%, transparent);
  background: color-mix(in srgb, var(--color-warning) 12%, transparent);
  color: color-mix(in srgb, var(--color-warning) 82%, var(--journal-ink));
}

.quick-action__icon--primary {
  border: 1px solid color-mix(in srgb, var(--journal-accent) 18%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 10%, transparent);
  color: var(--journal-accent-strong);
}

:deep(.section-card) {
  padding: 1.1rem 1.1rem 1.05rem;
  border: 1px solid var(--teacher-card-border);
  border-radius: 16px;
  background: var(--journal-surface-subtle);
  box-shadow: 0 10px 24px var(--color-shadow-soft);
}

:deep(.section-card__header) {
  margin-bottom: 1rem;
  border-bottom: 1px dashed var(--teacher-divider);
  padding-bottom: 0.75rem;
}

:deep(.section-card__body) {
  padding-left: 0;
}

@keyframes pulse {
  0%,
  100% {
    opacity: 0.58;
  }

  50% {
    opacity: 1;
  }
}

@media (max-width: 1279px) {
  .workspace-grid {
    grid-template-columns: 1fr;
  }

  .context-rail {
    padding: 0 28px 28px;
  }
}

@media (max-width: 1023px) {
  .summary-strip {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .teacher-topbar {
    align-items: flex-start;
    flex-direction: column;
  }
}

@media (max-width: 767px) {
  .workspace-topbar,
  .top-tabs,
  .content-pane,
  .context-rail {
    padding-left: 20px;
    padding-right: 20px;
  }

  .top-tabs {
    gap: 18px;
  }

  .summary-strip {
    grid-template-columns: 1fr;
  }

  .teacher-actions {
    width: 100%;
  }

  .teacher-btn {
    flex: 1 1 100%;
  }
}
</style>
