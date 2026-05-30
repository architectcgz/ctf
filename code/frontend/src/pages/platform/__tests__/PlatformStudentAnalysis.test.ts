import { existsSync, readFileSync } from 'node:fs'
import { resolve } from 'node:path'

import { flushPromises, mount } from '@vue/test-utils'
import { reactive } from 'vue'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import PlatformStudentAnalysis from '@/pages/platform/PlatformStudentAnalysisRoutePage.vue'
import {
  reportDialogStub,
  resetStudentAnalysisRouteTestState,
  type StudentAnalysisRouteTestState,
  type StudentAnalysisTeachingApiMocks,
} from '@/pages/__tests__/studentAnalysisRouteTestSupport'

const pushMock = vi.fn()
const replaceMock = vi.fn()
const routeMock = reactive<StudentAnalysisRouteTestState>({
  params: {
    className: 'Class A',
    studentId: 'stu-1',
  },
  query: {},
})

const teachingApiMocks = vi.hoisted<StudentAnalysisTeachingApiMocks>(() => ({
  getClasses: vi.fn(),
  getClassStudents: vi.fn(),
  getStudentProgress: vi.fn(),
  getStudentSkillProfile: vi.fn(),
  getStudentRecommendations: vi.fn(),
  getStudentTimeline: vi.fn(),
  getStudentEvidence: vi.fn(),
  getStudentAttackSessions: vi.fn(),
  getTeacherWriteupSubmissions: vi.fn(),
  recommendTeacherCommunityWriteup: vi.fn(),
  unrecommendTeacherCommunityWriteup: vi.fn(),
  hideTeacherCommunityWriteup: vi.fn(),
  restoreTeacherCommunityWriteup: vi.fn(),
  getTeacherManualReviewSubmissions: vi.fn(),
  getTeacherManualReviewSubmission: vi.fn(),
  reviewTeacherManualReviewSubmission: vi.fn(),
  exportStudentReviewArchive: vi.fn(),
}))

vi.mock('vue-router', async () => {
  const actual = await vi.importActual<typeof import('vue-router')>('vue-router')
  return {
    ...actual,
    useRouter: () => ({ push: pushMock, replace: replaceMock }),
    useRoute: () => routeMock,
  }
})

vi.mock('@/api/teaching', () => teachingApiMocks)

describe('PlatformStudentAnalysis route owner', () => {
  const studentAnalysisPageStub = {
    name: 'StudentAnalysisPage',
    props: [
      'selectedStudent',
      'loadingDetails',
      'error',
      'progress',
      'skillProfile',
      'recommendations',
      'timeline',
      'evidence',
      'attackSessions',
      'reviewChallengeOptions',
      'reviewWorkspaceLoading',
      'reviewWorkspaceQuery',
      'writeupSubmissions',
      'writeupPage',
      'writeupTotal',
      'writeupTotalPages',
      'writeupPaginationLoading',
      'manualReviewSubmissions',
      'activeManualReview',
      'manualReviewLoading',
      'manualReviewSaving',
      'solvedRate',
      'weakDimensions',
    ],
    template: `
      <div data-testid="student-analysis-page">
        <button id="retry" type="button" @click="$emit('retry')">retry</button>
        <button id="open-class-students" type="button" @click="$emit('openClassStudents')">
          open-class-students
        </button>
        <button id="open-review-archive" type="button" @click="$emit('openReviewArchive')">
          open-review-archive
        </button>
        <button id="open-report-export" type="button" @click="$emit('openReportExport')">
          open-report-export
        </button>
        <button id="open-challenge" type="button" @click="$emit('openChallenge', 'challenge-1')">
          open-challenge
        </button>
        <button id="export-review-archive" type="button" @click="$emit('exportReviewArchive')">
          export-review-archive
        </button>
        <button id="open-manual-review" type="button" @click="$emit('openManualReview', 'manual-1')">
          open-manual-review
        </button>
        <button
          id="moderate-writeup"
          type="button"
          @click="$emit('moderateWriteup', { submissionId: 'writeup-1', action: 'recommend' })"
        >
          moderate-writeup
        </button>
        <button
          id="review-manual-review"
          type="button"
          @click="$emit('reviewManualReview', { submissionId: 'manual-1', reviewStatus: 'approved' })"
        >
          review-manual-review
        </button>
        <button id="change-writeup-page" type="button" @click="$emit('changeWriteupPage', 2)">
          change-writeup-page
        </button>
        <button
          id="update-review-workspace-mode"
          type="button"
          @click="$emit('updateReviewWorkspaceFilters', { mode: 'awd' })"
        >
          update-review-workspace-mode
        </button>
        <button
          id="update-review-workspace-challenge"
          type="button"
          @click="$emit('updateReviewWorkspaceFilters', { challenge_id: '11' })"
        >
          update-review-workspace-challenge
        </button>
      </div>
    `,
  }

  beforeEach(() => {
    resetStudentAnalysisRouteTestState({
      pushMock,
      replaceMock,
      routeMock,
      teachingApiMocks,
      role: 'admin',
    })
    teachingApiMocks.getStudentRecommendations.mockResolvedValue({
      weak_dimensions: [],
      challenges: [],
    })
    teachingApiMocks.getStudentTimeline.mockResolvedValue([])
    teachingApiMocks.getStudentEvidence.mockResolvedValue({
      summary: {
        total_events: 0,
        proxy_request_count: 0,
        submit_count: 0,
        success_count: 0,
        challenge_id: '',
      },
      events: [],
    })
    teachingApiMocks.getStudentAttackSessions.mockResolvedValue({
      summary: {
        total_sessions: 0,
        success_count: 0,
        failed_count: 0,
        in_progress_count: 0,
        unknown_count: 0,
        event_count: 0,
        capture_available_count: 0,
      },
      sessions: [],
    })
    teachingApiMocks.getTeacherWriteupSubmissions.mockResolvedValue({
      list: [],
      total: 0,
      page: 1,
      page_size: 6,
    })
    teachingApiMocks.getTeacherManualReviewSubmissions.mockResolvedValue({
      list: [],
      total: 0,
      page: 1,
      page_size: 6,
    })
  })

  it('应使用平台 route view，并通过中性 feature 承接共享 page workflow', () => {
    const platformViewPath = resolve(
      process.cwd(),
      'src/pages/platform/PlatformStudentAnalysisRoutePage.vue'
    )
    const platformRoutesPath = resolve(process.cwd(), 'src/router/routes/platformRoutes.ts')

    expect(existsSync(platformViewPath)).toBe(true)

    const platformRoutesSource = readFileSync(platformRoutesPath, 'utf-8')
    expect(platformRoutesSource).toContain(
      "component: () => import('@/pages/platform/PlatformStudentAnalysisRoutePage.vue')"
    )

    if (!existsSync(platformViewPath)) {
      return
    }

    const platformViewSource = readFileSync(platformViewPath, 'utf-8')
    expect(platformViewSource).toContain("from '@/features/teaching/student-analysis-workspace'")
    expect(platformViewSource).toContain('StudentAnalysisPage')
    expect(platformViewSource).toContain('useStudentAnalysisPage')
    expect(platformViewSource).toContain(
      "import { ClassReportExportDialog } from '@/features/teaching/class-report-export'"
    )
    expect(platformViewSource).not.toContain("from '@/components/class-management'")
    expect(platformViewSource).not.toContain(
      "from '@/pages/teacher/TeacherStudentAnalysisRoutePage.vue'"
    )
    expect(platformViewSource).not.toContain("from '@/api/teacher'")
    expect(platformViewSource).not.toContain(
      '@/components/teacher/class-management/StudentAnalysisPage.vue'
    )
    expect(platformViewSource).not.toContain('ClassReportExportDialog.vue')
    expect(platformViewSource).not.toContain(':classes="classes"')
    expect(platformViewSource).not.toContain(':students="students"')
    expect(platformViewSource).not.toContain(':selected-class-name="selectedClassName"')
    expect(platformViewSource).not.toContain(':selected-student-id="selectedStudentId"')
    expect(platformViewSource).not.toContain(':loading-classes="loadingClasses"')
    expect(platformViewSource).not.toContain(':loading-students="loadingStudents"')
    expect(platformViewSource).not.toContain('@open-class-management="openClassManagement"')
    expect(platformViewSource).not.toContain('@select-class="selectClass"')
    expect(platformViewSource).not.toContain('@select-student="selectStudent"')
  })

  it('应在运行时把共享页面事件桥接到平台路由和导出弹窗 owner', async () => {
    const wrapper = mount(PlatformStudentAnalysis, {
      global: {
        stubs: {
          StudentAnalysisPage: studentAnalysisPageStub,
          ClassReportExportDialog: reportDialogStub,
        },
      },
    })

    await flushPromises()

    expect(teachingApiMocks.getClassStudents).toHaveBeenCalledWith('Class A')
    expect(teachingApiMocks.getStudentProgress).toHaveBeenCalledWith('stu-1')

    await wrapper.get('#open-class-students').trigger('click')
    await wrapper.get('#open-review-archive').trigger('click')
    await wrapper.get('#open-challenge').trigger('click')
    await wrapper.get('#open-report-export').trigger('click')
    await flushPromises()

    expect(pushMock).toHaveBeenCalledWith({
      name: 'PlatformClassStudents',
      params: { className: 'Class A' },
    })
    expect(pushMock).toHaveBeenCalledWith({
      name: 'PlatformStudentReviewArchive',
      params: {
        className: 'Class A',
        studentId: 'stu-1',
      },
    })
    expect(pushMock).toHaveBeenCalledWith({
      name: 'ChallengeDetail',
      params: {
        id: 'challenge-1',
      },
    })

    const dialog = wrapper.get('[data-testid="class-report-dialog"]')
    expect(dialog.attributes('data-open')).toBe('true')
    expect(dialog.attributes('data-default-class-name')).toBe('Class A')
  })

  it('应继续桥接共享页面剩余的 review workspace 与审核事件', async () => {
    const wrapper = mount(PlatformStudentAnalysis, {
      global: {
        stubs: {
          StudentAnalysisPage: studentAnalysisPageStub,
          ClassReportExportDialog: reportDialogStub,
        },
      },
    })

    await flushPromises()

    await wrapper.get('#open-manual-review').trigger('click')
    await flushPromises()
    expect(teachingApiMocks.getTeacherManualReviewSubmission).toHaveBeenCalledWith('manual-1')

    await wrapper.get('#moderate-writeup').trigger('click')
    await flushPromises()
    expect(teachingApiMocks.recommendTeacherCommunityWriteup).toHaveBeenCalledWith('writeup-1')

    await wrapper.get('#review-manual-review').trigger('click')
    await flushPromises()
    expect(teachingApiMocks.reviewTeacherManualReviewSubmission).toHaveBeenCalledWith('manual-1', {
      review_status: 'approved',
      review_comment: undefined,
    })

    teachingApiMocks.getTeacherWriteupSubmissions.mockClear()
    await wrapper.get('#change-writeup-page').trigger('click')
    await flushPromises()
    expect(teachingApiMocks.getTeacherWriteupSubmissions).toHaveBeenCalledWith({
      student_id: 'stu-1',
      submission_status: 'published',
      page: 2,
      page_size: 6,
    })

    teachingApiMocks.getStudentAttackSessions.mockClear()
    replaceMock.mockClear()
    await wrapper.get('#update-review-workspace-mode').trigger('click')
    await flushPromises()
    expect(teachingApiMocks.getStudentAttackSessions).toHaveBeenCalledWith('stu-1', {
      with_events: true,
      limit: 20,
      offset: 0,
      mode: 'awd',
    })
    expect(replaceMock).toHaveBeenCalledWith({
      query: {
        reviewMode: 'awd',
        reviewResult: undefined,
        reviewChallengeId: undefined,
      },
    })

    teachingApiMocks.getStudentEvidence.mockClear()
    teachingApiMocks.getStudentAttackSessions.mockClear()
    replaceMock.mockClear()
    await wrapper.get('#update-review-workspace-challenge').trigger('click')
    await flushPromises()
    expect(teachingApiMocks.getStudentEvidence).toHaveBeenCalledWith('stu-1', {
      challenge_id: '11',
    })
    expect(teachingApiMocks.getStudentAttackSessions).toHaveBeenCalledWith('stu-1', {
      with_events: true,
      limit: 20,
      offset: 0,
      mode: 'awd',
      challenge_id: '11',
    })
    expect(replaceMock).toHaveBeenCalledWith({
      query: {
        reviewMode: 'awd',
        reviewResult: undefined,
        reviewChallengeId: '11',
      },
    })
  })

  it('应继续桥接共享页面的复盘归档导出事件而不抛到全局错误页', async () => {
    teachingApiMocks.exportStudentReviewArchive.mockRejectedValue(new Error('导出失败'))

    const wrapper = mount(PlatformStudentAnalysis, {
      global: {
        stubs: {
          StudentAnalysisPage: studentAnalysisPageStub,
          ClassReportExportDialog: reportDialogStub,
        },
      },
    })

    await flushPromises()

    await expect(wrapper.get('#export-review-archive').trigger('click')).resolves.toBeUndefined()
    await flushPromises()

    expect(teachingApiMocks.exportStudentReviewArchive).toHaveBeenCalledWith('stu-1', {
      format: 'json',
    })
  })

  it('应把 retry 事件重新桥接到页面初始化链路', async () => {
    const wrapper = mount(PlatformStudentAnalysis, {
      global: {
        stubs: {
          StudentAnalysisPage: studentAnalysisPageStub,
          ClassReportExportDialog: reportDialogStub,
        },
      },
    })

    await flushPromises()

    teachingApiMocks.getClassStudents.mockClear()
    teachingApiMocks.getStudentProgress.mockClear()
    teachingApiMocks.getStudentEvidence.mockClear()
    teachingApiMocks.getStudentAttackSessions.mockClear()

    await wrapper.get('#retry').trigger('click')
    await flushPromises()

    expect(teachingApiMocks.getClassStudents).toHaveBeenCalledWith('Class A')
    expect(teachingApiMocks.getStudentProgress).toHaveBeenCalledWith('stu-1')
    expect(teachingApiMocks.getStudentEvidence).toHaveBeenCalledWith('stu-1', {})
    expect(teachingApiMocks.getStudentAttackSessions).toHaveBeenCalledWith('stu-1', {
      with_events: true,
      limit: 20,
      offset: 0,
    })
  })
})
