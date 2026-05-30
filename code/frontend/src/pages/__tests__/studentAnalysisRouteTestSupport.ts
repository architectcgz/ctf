import { createPinia, setActivePinia } from 'pinia'
import { vi } from 'vitest'

import { useAuthStore } from '@/stores/auth'

type MockFunction = ReturnType<typeof vi.fn>

export interface StudentAnalysisRouteTestState {
  params: {
    className: string
    studentId: string
  }
  query: Record<string, unknown>
}

export interface StudentAnalysisTeachingApiMocks {
  getClasses: MockFunction
  getClassStudents: MockFunction
  getStudentProgress: MockFunction
  getStudentSkillProfile: MockFunction
  getStudentRecommendations: MockFunction
  getStudentTimeline: MockFunction
  getStudentEvidence: MockFunction
  getStudentAttackSessions: MockFunction
  getTeacherWriteupSubmissions: MockFunction
  recommendTeacherCommunityWriteup: MockFunction
  unrecommendTeacherCommunityWriteup: MockFunction
  hideTeacherCommunityWriteup: MockFunction
  restoreTeacherCommunityWriteup: MockFunction
  getTeacherManualReviewSubmissions: MockFunction
  getTeacherManualReviewSubmission: MockFunction
  reviewTeacherManualReviewSubmission: MockFunction
  exportStudentReviewArchive: MockFunction
}

interface ResetStudentAnalysisRouteTestStateOptions {
  pushMock: MockFunction
  replaceMock: MockFunction
  routeMock: StudentAnalysisRouteTestState
  teachingApiMocks: StudentAnalysisTeachingApiMocks
  role: 'teacher' | 'admin'
}

export const reportDialogStub = {
  name: 'ClassReportExportDialog',
  props: ['modelValue', 'defaultClassName'],
  template:
    '<div data-testid="class-report-dialog" :data-open="String(modelValue)" :data-default-class-name="defaultClassName || \'\'" />',
}

export function resetStudentAnalysisRouteTestState(
  options: ResetStudentAnalysisRouteTestStateOptions
): void {
  const { pushMock, replaceMock, routeMock, teachingApiMocks, role } = options

  setActivePinia(createPinia())
  localStorage.clear()
  pushMock.mockReset()
  replaceMock.mockReset()
  routeMock.params.className = 'Class A'
  routeMock.params.studentId = 'stu-1'
  routeMock.query = {}

  Object.values(teachingApiMocks).forEach((mock) => mock.mockReset())

  teachingApiMocks.getClasses.mockResolvedValue([{ name: 'Class A', student_count: 2 }])
  teachingApiMocks.getClassStudents.mockResolvedValue([
    { id: 'stu-1', username: 'alice' },
    { id: 'stu-2', username: 'bob' },
  ])
  teachingApiMocks.getStudentProgress.mockResolvedValue({
    total_challenges: 4,
    solved_challenges: 2,
    by_category: {},
    by_difficulty: {},
  })
  teachingApiMocks.getStudentSkillProfile.mockResolvedValue({
    dimensions: [{ key: 'crypto', name: '密码', value: 35 }],
  })
  teachingApiMocks.getStudentRecommendations.mockResolvedValue({
    weak_dimensions: [
      {
        dimension: 'crypto',
        label: '密码',
        severity: 'warning',
        confidence: 0.83,
        evidence: '当前密码维度已经形成高置信度薄弱信号。',
      },
    ],
    challenges: [
      {
        challenge_id: '12',
        title: 'crypto-lab',
        category: 'crypto',
        difficulty: 'medium',
        summary: '针对薄弱维度：密码',
        evidence: '当前密码维度已经形成高置信度薄弱信号。',
      },
    ],
  })
  teachingApiMocks.getStudentTimeline.mockResolvedValue([
    {
      id: 'challenge_detail_view-11-2026-03-11T09:00:00Z',
      type: 'challenge_detail_view',
      title: 'web-1',
      detail: '查看题目详情，开始分析题面与环境线索',
      created_at: '2026-03-11T09:00:00Z',
      challenge_id: '11',
      meta: { raw_type: 'challenge_detail_view' },
    },
    {
      id: 'hint_unlock-11-2026-03-11T09:30:00Z',
      type: 'hint',
      title: 'web-1',
      detail: '解锁第 1 级提示：先看回显',
      created_at: '2026-03-11T09:30:00Z',
      challenge_id: '11',
      meta: { raw_type: 'hint_unlock' },
    },
    {
      id: 'instance_access-11-2026-03-11T09:40:00Z',
      type: 'instance_access',
      title: 'web-1',
      detail: '访问攻击目标，开始与靶机进行实际交互',
      created_at: '2026-03-11T09:40:00Z',
      challenge_id: '11',
      meta: { raw_type: 'instance_access' },
    },
    {
      id: 'instance_extend-11-2026-03-11T09:45:00Z',
      type: 'instance_extend',
      title: 'web-1',
      detail: '延长实例有效期，继续当前利用过程',
      created_at: '2026-03-11T09:45:00Z',
      challenge_id: '11',
      meta: { raw_type: 'instance_extend' },
    },
    {
      id: 'flag_submit-11-2026-03-11T10:00:00Z',
      type: 'solve',
      title: 'web-1',
      detail: '第 2 次提交命中 Flag，获得 100 分',
      created_at: '2026-03-11T10:00:00Z',
      challenge_id: '11',
      points: 100,
      meta: { raw_type: 'flag_submit' },
    },
  ])
  teachingApiMocks.getStudentEvidence.mockResolvedValue({
    summary: {
      total_events: 5,
      proxy_request_count: 1,
      submit_count: 2,
      success_count: 1,
      challenge_id: '11',
    },
    events: [
      {
        type: 'instance_access',
        challenge_id: '11',
        title: 'web-1',
        detail: '访问攻击目标，开始与靶机进行实际交互',
        timestamp: '2026-03-11T09:40:00Z',
        meta: { event_stage: 'access' },
      },
      {
        type: 'instance_proxy_request',
        challenge_id: '11',
        title: 'web-1',
        detail: '经平台代理发起 POST /login，请求返回 200，携带请求摘要',
        timestamp: '2026-03-11T09:42:00Z',
        meta: { event_stage: 'exploit', method: 'POST' },
      },
    ],
  })
  teachingApiMocks.getStudentAttackSessions.mockResolvedValue({
    summary: {
      total_sessions: 1,
      success_count: 1,
      failed_count: 0,
      in_progress_count: 0,
      unknown_count: 0,
      event_count: 3,
      capture_available_count: 0,
    },
    sessions: [
      {
        id: 'sess-1',
        mode: 'practice',
        student_id: 'stu-1',
        challenge_id: '11',
        title: 'web-1',
        started_at: '2026-03-11T09:40:00Z',
        ended_at: '2026-03-11T10:00:00Z',
        result: 'success',
        event_count: 3,
        capture_count: 0,
        events: [
          {
            id: 'evt-1',
            session_id: 'sess-1',
            type: 'instance_proxy_request',
            stage: 'exploit',
            source: 'audit_logs',
            occurred_at: '2026-03-11T09:42:00Z',
            actor: { user_id: 'stu-1' },
            target: { challenge_id: '11' },
            summary: '经平台代理发起 POST /login，请求返回 200，携带请求摘要',
            meta: {
              request_method: 'POST',
              target_path: '/login',
              status_code: 200,
            },
            capture_available: false,
          },
        ],
      },
    ],
  })
  teachingApiMocks.getTeacherWriteupSubmissions.mockResolvedValue({
    list: [
      {
        id: 'writeup-1',
        user_id: 'stu-1',
        student_username: 'alice',
        challenge_id: '11',
        challenge_title: 'web-1',
        title: '从回显到 flag',
        content_preview: '先看登录回显，再确定注入点。',
        submission_status: 'published',
        visibility_status: 'visible',
        is_recommended: true,
        published_at: '2026-03-11T10:50:00Z',
        updated_at: '2026-03-11T11:00:00Z',
      },
    ],
    total: 1,
    page: 1,
    page_size: 6,
  })
  teachingApiMocks.getTeacherManualReviewSubmissions.mockResolvedValue({
    list: [
      {
        id: 'manual-1',
        user_id: 'stu-1',
        student_username: 'alice',
        challenge_id: '12',
        challenge_title: 'misc-essay',
        answer_preview: '先提交利用思路，再说明证据链。',
        review_status: 'pending',
        submitted_at: '2026-03-11T12:00:00Z',
        updated_at: '2026-03-11T12:10:00Z',
      },
    ],
    total: 1,
    page: 1,
    page_size: 6,
  })
  teachingApiMocks.getTeacherManualReviewSubmission.mockResolvedValue({
    id: 'manual-1',
    user_id: 'stu-1',
    student_username: 'alice',
    challenge_id: '12',
    challenge_title: 'misc-essay',
    answer: '完整答案正文',
    is_correct: false,
    score: 0,
    review_status: 'pending',
    submitted_at: '2026-03-11T12:00:00Z',
    updated_at: '2026-03-11T12:10:00Z',
  })
  teachingApiMocks.reviewTeacherManualReviewSubmission.mockResolvedValue({
    id: 'manual-1',
    user_id: 'stu-1',
    student_username: 'alice',
    challenge_id: '12',
    challenge_title: 'misc-essay',
    answer: '完整答案正文',
    is_correct: true,
    score: 100,
    review_status: 'approved',
    review_comment: '通过',
    submitted_at: '2026-03-11T12:00:00Z',
    updated_at: '2026-03-11T12:20:00Z',
  })
  teachingApiMocks.recommendTeacherCommunityWriteup.mockResolvedValue({
    id: 'writeup-1',
    user_id: 'stu-1',
    challenge_id: '11',
    title: '从回显到 flag',
    content: '完整题解',
    submission_status: 'published',
    visibility_status: 'visible',
    is_recommended: true,
    published_at: '2026-03-11T10:50:00Z',
    created_at: '2026-03-11T10:40:00Z',
    updated_at: '2026-03-11T11:00:00Z',
  })
  teachingApiMocks.unrecommendTeacherCommunityWriteup.mockResolvedValue({
    id: 'writeup-1',
    user_id: 'stu-1',
    challenge_id: '11',
    title: '从回显到 flag',
    content: '完整题解',
    submission_status: 'published',
    visibility_status: 'visible',
    is_recommended: false,
    published_at: '2026-03-11T10:50:00Z',
    created_at: '2026-03-11T10:40:00Z',
    updated_at: '2026-03-11T11:00:00Z',
  })
  teachingApiMocks.hideTeacherCommunityWriteup.mockResolvedValue({
    id: 'writeup-1',
    user_id: 'stu-1',
    challenge_id: '11',
    title: '从回显到 flag',
    content: '完整题解',
    submission_status: 'published',
    visibility_status: 'hidden',
    is_recommended: false,
    published_at: '2026-03-11T10:50:00Z',
    created_at: '2026-03-11T10:40:00Z',
    updated_at: '2026-03-11T11:00:00Z',
  })
  teachingApiMocks.restoreTeacherCommunityWriteup.mockResolvedValue({
    id: 'writeup-1',
    user_id: 'stu-1',
    challenge_id: '11',
    title: '从回显到 flag',
    content: '完整题解',
    submission_status: 'published',
    visibility_status: 'visible',
    is_recommended: false,
    published_at: '2026-03-11T10:50:00Z',
    created_at: '2026-03-11T10:40:00Z',
    updated_at: '2026-03-11T11:00:00Z',
  })
  teachingApiMocks.exportStudentReviewArchive.mockResolvedValue({
    report_id: 'report-1',
    status: 'processing',
  })

  const authStore = useAuthStore()
  authStore.setAuth({
    id: `${role}-1`,
    username: role,
    role,
    class_name: role === 'teacher' ? 'Class A' : '',
  })
}
