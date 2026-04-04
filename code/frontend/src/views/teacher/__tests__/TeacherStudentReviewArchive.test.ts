import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import TeacherStudentReviewArchive from '../TeacherStudentReviewArchive.vue'

const pushMock = vi.fn()
const routeMock = {
  params: {
    className: 'Class A',
    studentId: 'stu-1',
  },
}

const teacherApiMocks = vi.hoisted(() => ({
  getStudentReviewArchive: vi.fn(),
  exportStudentReviewArchive: vi.fn(),
}))

const assessmentApiMocks = vi.hoisted(() => ({
  downloadReport: vi.fn(),
}))

vi.mock('vue-router', async () => {
  const actual = await vi.importActual<typeof import('vue-router')>('vue-router')
  return {
    ...actual,
    useRouter: () => ({ push: pushMock }),
    useRoute: () => routeMock,
  }
})

vi.mock('@/api/teacher', () => teacherApiMocks)
vi.mock('@/api/assessment', () => assessmentApiMocks)

describe('TeacherStudentReviewArchive', () => {
  beforeEach(() => {
    pushMock.mockReset()
    routeMock.params.className = 'Class A'
    routeMock.params.studentId = 'stu-1'
    Object.values(teacherApiMocks).forEach((mock) => mock.mockReset())
    Object.values(assessmentApiMocks).forEach((mock) => mock.mockReset())

    teacherApiMocks.getStudentReviewArchive.mockResolvedValue({
      generated_at: '2026-04-01T09:30:00Z',
      student: {
        id: 'stu-1',
        username: 'alice',
        name: 'Alice',
        class_name: 'Class A',
      },
      summary: {
        total_challenges: 18,
        total_solved: 6,
        total_score: 620,
        rank: 2,
        total_attempts: 15,
        timeline_event_count: 9,
        evidence_event_count: 6,
        writeup_count: 1,
        manual_review_count: 1,
        correct_submission_count: 1,
        last_activity_at: '2026-04-01T09:20:00Z',
      },
      skill_profile: {
        dimensions: [
          { key: 'web', name: 'Web', value: 82 },
          { key: 'crypto', name: '密码', value: 48 },
        ],
        updated_at: '2026-04-01T09:30:00Z',
      },
      timeline: [
        {
          id: 'hint-1',
          type: 'hint',
          title: 'web-1',
          detail: '解锁第 1 级提示',
          created_at: '2026-04-01T09:00:00Z',
          challenge_id: '11',
        },
        {
          id: 'solve-1',
          type: 'solve',
          title: 'web-1',
          detail: '提交命中 Flag',
          created_at: '2026-04-01T09:20:00Z',
          challenge_id: '11',
          points: 100,
        },
      ],
      evidence: [
        {
          type: 'instance_access',
          challenge_id: '11',
          title: 'web-1',
          detail: '访问攻击目标',
          timestamp: '2026-04-01T09:02:00Z',
          meta: { event_stage: 'access' },
        },
        {
          type: 'instance_proxy_request',
          challenge_id: '11',
          title: 'web-1',
          detail: '经平台代理发起 POST /login',
          timestamp: '2026-04-01T09:03:00Z',
          meta: { event_stage: 'exploit', method: 'POST' },
        },
      ],
      writeups: [
        {
          id: 'writeup-1',
          challenge_id: '11',
          challenge_title: 'web-1',
          title: '从回显到 flag',
          submission_status: 'published',
          visibility_status: 'visible',
          is_recommended: true,
          published_at: '2026-04-01T09:25:00Z',
          updated_at: '2026-04-01T09:25:00Z',
        },
      ],
      manual_reviews: [
        {
          id: 'manual-1',
          challenge_id: '12',
          challenge_title: 'misc-essay',
          answer: '完整答案正文',
          review_status: 'approved',
          submitted_at: '2026-04-01T09:28:00Z',
          score: 100,
          review_comment: '通过',
          reviewer_name: 'teacher-a',
        },
      ],
      teacher_observations: {
        items: [
          {
            key: 'training_closure',
            label: '训练闭环',
            level: 'good',
            summary: '已形成从利用到复盘输出的有效闭环。',
            evidence: '命中正确提交 1 次，复盘材料 1 份。',
          },
          {
            key: 'hint_usage',
            label: '提示依赖',
            level: 'attention',
            summary: '训练过程存在提示介入。',
            evidence: '共记录提示解锁 1 次。',
          },
        ],
      },
    })
  })

  it('应该渲染完整复盘页的核心区块', async () => {
    const wrapper = mount(TeacherStudentReviewArchive)

    await flushPromises()

    expect(teacherApiMocks.getStudentReviewArchive).toHaveBeenCalledWith('stu-1')
    expect(wrapper.text()).toContain('Alice')
    expect(wrapper.text()).toContain('教学复盘归档')
    expect(wrapper.text()).toContain('训练闭环')
    expect(wrapper.text()).toContain('提示依赖')
    expect(wrapper.text()).toContain('攻防证据链')
    expect(wrapper.text()).toContain('从回显到 flag')
    expect(wrapper.text()).toContain('misc-essay')
  })
})
