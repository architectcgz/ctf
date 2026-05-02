import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { flushPromises, mount } from '@vue/test-utils'

import TeacherStudentReviewArchive from '../TeacherStudentReviewArchive.vue'
import reviewArchiveSource from '../TeacherStudentReviewArchive.vue?raw'
import reviewArchiveWorkspaceSource from '@/widgets/teacher-review-archive/TeacherReviewArchiveWorkspace.vue?raw'
import reviewArchiveHeroSource from '@/components/teacher/review-archive/ReviewArchiveHero.vue?raw'
import { useAuthStore } from '@/stores/auth'

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
    setActivePinia(createPinia())
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
        {
          id: 'awd-hit-1',
          type: 'awd_attack_submit',
          title: 'awd-web',
          detail: 'AWD 攻击命中 Blue Team，得分 120',
          created_at: '2026-04-01T09:24:00Z',
          challenge_id: '21',
          is_correct: true,
          points: 120,
          meta: {
            raw_type: 'awd_attack_submit',
          },
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
        {
          type: 'awd_attack_submission',
          challenge_id: '21',
          title: 'awd-web',
          detail: 'AWD 攻击命中 Blue Team，得分 120',
          timestamp: '2026-04-01T09:24:00Z',
          meta: {
            event_stage: 'exploit',
            is_success: true,
            score_gained: 120,
            victim_team_name: 'Blue Team',
          },
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

    const authStore = useAuthStore()
    authStore.setAuth(
      {
        id: 'teacher-1',
        username: 'teacher',
        role: 'teacher',
        class_name: 'Class A',
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
    expect(wrapper.text()).toContain('练习复盘')
    expect(wrapper.text()).toContain('AWD 复盘')
    expect(wrapper.text()).toContain('Blue Team')
    expect(wrapper.text()).toContain('从回显到 flag')
    expect(wrapper.text()).toContain('misc-essay')
  })

  it('复盘归档操作按钮应接入共享 ui-btn 原语', () => {
    expect(reviewArchiveSource).toContain(
      "import { TeacherReviewArchiveWorkspace } from '@/widgets/teacher-review-archive'"
    )
    expect(reviewArchiveSource).toContain('<TeacherReviewArchiveWorkspace')
    expect(reviewArchiveWorkspaceSource).toContain('class="ui-btn ui-btn--primary"')
    expect(reviewArchiveWorkspaceSource).not.toContain('<ElButton')
    expect(reviewArchiveHeroSource).toContain('class="ui-btn ui-btn--secondary"')
    expect(reviewArchiveHeroSource).toContain('class="ui-btn ui-btn--primary"')
    expect(reviewArchiveHeroSource).not.toContain('<ElButton')
  })

  it('管理员在复盘归档页返回分析和班级页时应使用后台路由', async () => {
    const authStore = useAuthStore()
    authStore.setAuth(
      {
        id: 'admin-1',
        username: 'admin',
        role: 'admin',
        class_name: 'Class A',
      })

    const wrapper = mount(TeacherStudentReviewArchive)

    await flushPromises()

    wrapper.findComponent({ name: 'ReviewArchiveHero' }).vm.$emit('openAnalysis')
    wrapper.findComponent({ name: 'ReviewArchiveHero' }).vm.$emit('back')

    expect(pushMock).toHaveBeenCalledWith({
      name: 'PlatformStudentAnalysis',
      params: {
        className: 'Class A',
        studentId: 'stu-1',
      },
    })
    expect(pushMock).toHaveBeenCalledWith({
      name: 'PlatformClassStudents',
      params: { className: 'Class A' },
    })
  })

  it('导出复盘归档失败时不应抛到全局错误页', async () => {
    teacherApiMocks.exportStudentReviewArchive.mockRejectedValue(new Error('导出失败'))

    const wrapper = mount(TeacherStudentReviewArchive, {
      global: {
        stubs: {
          ReviewArchiveHero: {
            name: 'ReviewArchiveHero',
            template:
              '<button id="export-archive" type="button" @click="$emit(\'exportArchive\')">导出复盘归档</button>',
          },
        },
      },
    })

    await flushPromises()

    await expect(wrapper.get('#export-archive').trigger('click')).resolves.toBeUndefined()
    await flushPromises()

    expect(teacherApiMocks.exportStudentReviewArchive).toHaveBeenCalledWith('stu-1', {
      format: 'json',
    })
  })
})
