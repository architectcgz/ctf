import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { flushPromises, mount } from '@vue/test-utils'

import TeacherStudentAnalysis from '../TeacherStudentAnalysis.vue'

const pushMock = vi.fn()
const routeMock = {
  params: {
    className: 'Class A',
    studentId: 'stu-1',
  },
}

const teacherApiMocks = vi.hoisted(() => ({
  getClasses: vi.fn(),
  getClassStudents: vi.fn(),
  getStudentProgress: vi.fn(),
  getStudentSkillProfile: vi.fn(),
  getStudentRecommendations: vi.fn(),
  getStudentTimeline: vi.fn(),
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

describe('TeacherStudentAnalysis', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    localStorage.clear()
    pushMock.mockReset()
    routeMock.params.className = 'Class A'
    routeMock.params.studentId = 'stu-1'

    Object.values(teacherApiMocks).forEach((mock) => mock.mockReset())

    teacherApiMocks.getClasses.mockResolvedValue([{ name: 'Class A', student_count: 2 }])
    teacherApiMocks.getClassStudents.mockResolvedValue([
      { id: 'stu-1', username: 'alice' },
      { id: 'stu-2', username: 'bob' },
    ])
    teacherApiMocks.getStudentProgress.mockResolvedValue({
      total_challenges: 4,
      solved_challenges: 2,
      by_category: {},
      by_difficulty: {},
    })
    teacherApiMocks.getStudentSkillProfile.mockResolvedValue({
      dimensions: [{ key: 'crypto', name: '密码', value: 35 }],
    })
    teacherApiMocks.getStudentRecommendations.mockResolvedValue([
      { challenge_id: '12', title: 'crypto-lab', category: 'crypto', difficulty: 'medium', reason: '针对薄弱维度：密码' },
    ])
    teacherApiMocks.getStudentTimeline.mockResolvedValue([
      {
        id: 'challenge_detail_view-11-2026-03-11T09:00:00Z',
        type: 'challenge_detail_view',
        title: 'web-1',
        detail: '查看题目详情，开始分析题面与环境线索',
        created_at: '2026-03-11T09:00:00Z',
        challenge_id: '11',
        meta: {
          raw_type: 'challenge_detail_view',
        },
      },
      {
        id: 'hint_unlock-11-2026-03-11T09:30:00Z',
        type: 'hint',
        title: 'web-1',
        detail: '解锁第 1 级提示：先看回显',
        created_at: '2026-03-11T09:30:00Z',
        challenge_id: '11',
        meta: {
          raw_type: 'hint_unlock',
        },
      },
      {
        id: 'instance_access-11-2026-03-11T09:40:00Z',
        type: 'instance_access',
        title: 'web-1',
        detail: '访问攻击目标，开始与靶机进行实际交互',
        created_at: '2026-03-11T09:40:00Z',
        challenge_id: '11',
        meta: {
          raw_type: 'instance_access',
        },
      },
      {
        id: 'instance_extend-11-2026-03-11T09:45:00Z',
        type: 'instance_extend',
        title: 'web-1',
        detail: '延长实例有效期，继续当前利用过程',
        created_at: '2026-03-11T09:45:00Z',
        challenge_id: '11',
        meta: {
          raw_type: 'instance_extend',
        },
      },
      {
        id: 'flag_submit-11-2026-03-11T10:00:00Z',
        type: 'solve',
        title: 'web-1',
        detail: '第 2 次提交命中 Flag，获得 100 分',
        created_at: '2026-03-11T10:00:00Z',
        challenge_id: '11',
        points: 100,
        meta: {
          raw_type: 'flag_submit',
        },
      },
    ])
  })

  it('应该展示当前学员分析内容', async () => {
    const wrapper = mount(TeacherStudentAnalysis, {
      global: {
        stubs: {
          SkillRadar: true,
        },
      },
    })

    await flushPromises()

    expect(wrapper.text()).toContain('alice')
    expect(wrapper.text()).toContain('50%')
    expect(wrapper.text()).toContain('crypto-lab')
    expect(wrapper.text()).toContain('web-1')
    expect(wrapper.text()).toContain('查看题目详情')
    expect(wrapper.text()).toContain('解锁第 1 级提示')
    expect(wrapper.text()).toContain('访问攻击目标')
    expect(wrapper.text()).toContain('延长实例有效期')
    expect(wrapper.text()).toContain('第 2 次提交命中 Flag')
  })

  it('应该支持包含百分号的班级名路由参数', async () => {
    routeMock.params.className = '100% 班级'

    mount(TeacherStudentAnalysis, {
      global: {
        stubs: {
          SkillRadar: true,
        },
      },
    })

    await flushPromises()

    expect(teacherApiMocks.getClassStudents).toHaveBeenCalledWith('100% 班级')
  })
})
