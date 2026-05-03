import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { flushPromises, mount } from '@vue/test-utils'

import TeacherStudentAnalysis from '../TeacherStudentAnalysis.vue'
import teacherStudentAnalysisSource from '../TeacherStudentAnalysis.vue?raw'
import studentAnalysisPageSource from '@/components/teacher/class-management/StudentAnalysisPage.vue?raw'
import { useAuthStore } from '@/stores/auth'

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
  getStudentEvidence: vi.fn(),
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
    useRouter: () => ({ push: pushMock }),
    useRoute: () => routeMock,
  }
})

vi.mock('@/api/teacher', () => teacherApiMocks)

describe('TeacherStudentAnalysis', () => {
  const reportDialogStub = {
    name: 'TeacherClassReportExportDialog',
    props: ['modelValue', 'defaultClassName'],
    template:
      '<div data-testid="class-report-dialog" :data-open="String(modelValue)" :data-default-class-name="defaultClassName || \'\'" />',
  }

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
      {
        challenge_id: '12',
        title: 'crypto-lab',
        category: 'crypto',
        difficulty: 'medium',
        reason: '针对薄弱维度：密码',
      },
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
    teacherApiMocks.getStudentEvidence.mockResolvedValue({
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
    teacherApiMocks.getTeacherWriteupSubmissions.mockResolvedValue({
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
    teacherApiMocks.getTeacherManualReviewSubmissions.mockResolvedValue({
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
    teacherApiMocks.getTeacherManualReviewSubmission.mockResolvedValue({
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
    teacherApiMocks.reviewTeacherManualReviewSubmission.mockResolvedValue({
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
    teacherApiMocks.recommendTeacherCommunityWriteup.mockResolvedValue({
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
    teacherApiMocks.unrecommendTeacherCommunityWriteup.mockResolvedValue({
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
    teacherApiMocks.hideTeacherCommunityWriteup.mockResolvedValue({
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
    teacherApiMocks.restoreTeacherCommunityWriteup.mockResolvedValue({
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
    teacherApiMocks.exportStudentReviewArchive.mockResolvedValue({
      report_id: 'report-1',
      status: 'processing',
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

  it('应该展示当前学员分析内容', async () => {
    const wrapper = mount(TeacherStudentAnalysis, {
      global: {
        stubs: {
          SkillRadar: true,
          TeacherClassReportExportDialog: reportDialogStub,
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
    expect(wrapper.text()).toContain('攻防证据链')
    expect(wrapper.text()).toContain('人工审核题')
    expect(wrapper.text()).toContain('misc-essay')
    expect(wrapper.text()).toContain('从回显到 flag')
    expect(wrapper.text()).toContain('总事件数')
    expect(wrapper.text()).toContain('5')
    expect(wrapper.text()).toContain('利用请求')
    expect(wrapper.text()).toContain('POST /login')
    expect(wrapper.text()).toContain('社区题解状态')
    expect(wrapper.text()).toContain('推荐题解')
    expect(wrapper.text()).toContain('已公开')
    expect(wrapper.text()).toContain('取消推荐')

    expect(teacherApiMocks.getStudentEvidence).toHaveBeenCalledWith('stu-1')
    expect(teacherApiMocks.getTeacherWriteupSubmissions).toHaveBeenCalledWith({
      student_id: 'stu-1',
      submission_status: 'published',
      page: 1,
      page_size: 6,
    })
    expect(teacherApiMocks.getTeacherManualReviewSubmissions).toHaveBeenCalledWith({
      student_id: 'stu-1',
      page_size: 6,
    })
  })

  it('路由页应仅负责组合，不直接处理路由解析逻辑', () => {
    expect(teacherStudentAnalysisSource).toContain('useTeacherStudentAnalysisPage')
    expect(teacherStudentAnalysisSource).not.toContain('resolveClassManagementRouteName')
    expect(teacherStudentAnalysisSource).not.toContain('resolveClassStudentsRouteName')
  })

  it('学员分析头部应只保留姓名标题，不重复显示英文 eyebrow 和用户名 chip', () => {
    expect(studentAnalysisPageSource).not.toContain('Student Analysis')
    expect(studentAnalysisPageSource).not.toContain('teacher-student-chip')
    expect(studentAnalysisPageSource).not.toContain('teacher-eyebrow-row')
    expect(studentAnalysisPageSource).toContain(
      "{{ selectedStudent?.name || selectedStudent?.username || '学员分析' }}"
    )
  })

  it('应该支持隐藏社区题解', async () => {
    const wrapper = mount(TeacherStudentAnalysis, {
      global: {
        stubs: {
          SkillRadar: true,
          TeacherClassReportExportDialog: reportDialogStub,
        },
      },
    })

    await flushPromises()

    const hideButton = wrapper
      .findAll('button')
      .find((button) => button.text().includes('隐藏题解'))
    expect(hideButton).toBeDefined()

    await hideButton?.trigger('click')
    await flushPromises()

    expect(teacherApiMocks.hideTeacherCommunityWriteup).toHaveBeenCalledWith('writeup-1')
    expect(teacherApiMocks.getTeacherWriteupSubmissions).toHaveBeenCalledTimes(2)
  })

  it('应该支持包含百分号的班级名路由参数', async () => {
    routeMock.params.className = '100% 班级'

    mount(TeacherStudentAnalysis, {
      global: {
        stubs: {
          SkillRadar: true,
          TeacherClassReportExportDialog: reportDialogStub,
        },
      },
    })

    await flushPromises()

    expect(teacherApiMocks.getClassStudents).toHaveBeenCalledWith('100% 班级')
  })

  it('应该支持跳转到完整复盘页', async () => {
    const wrapper = mount(TeacherStudentAnalysis, {
      global: {
        stubs: {
          SkillRadar: true,
          TeacherClassReportExportDialog: reportDialogStub,
        },
      },
    })

    await flushPromises()

    const reviewButton = wrapper
      .findAll('button')
      .find((button) => button.text().includes('完整复盘页'))

    expect(reviewButton).toBeDefined()

    await reviewButton?.trigger('click')

    expect(pushMock).toHaveBeenCalledWith({
      name: 'TeacherStudentReviewArchive',
      params: {
        className: 'Class A',
        studentId: 'stu-1',
      },
    })
  })

  it('应采用顶部 tabs 工作区壳层而不是把所有内容堆叠在主页面，并去掉页面内重复顶栏', async () => {
    const wrapper = mount(TeacherStudentAnalysis, {
      global: {
        stubs: {
          SkillRadar: true,
          TeacherClassReportExportDialog: reportDialogStub,
        },
      },
    })

    await flushPromises()

    expect(wrapper.find('[role="tablist"]').exists()).toBe(true)
    expect(wrapper.find('#student-tab-overview').exists()).toBe(true)
    expect(wrapper.find('#student-tab-recommendations').exists()).toBe(true)
    expect(wrapper.find('#student-tab-writeups').exists()).toBe(true)
    expect(wrapper.find('#student-tab-evidence').exists()).toBe(true)
    expect(wrapper.find('#student-tab-timeline').exists()).toBe(true)
    expect(studentAnalysisPageSource).toMatch(/class="[^"]*\bworkspace-shell\b[^"]*"/)
    expect(studentAnalysisPageSource).not.toContain('class="workspace-topbar"')
    expect(studentAnalysisPageSource).toContain('class="top-tabs"')
    expect(studentAnalysisPageSource).toContain('class="content-pane"')
    expect(studentAnalysisPageSource).toMatch(
      /<div class="[^"]*\bworkspace-shell\b[^"]*">[\s\S]*<nav class="top-tabs"[\s\S]*<main class="content-pane">/s
    )
  })

  it('点击导出班级报告时应打开当前班级上下文对话框', async () => {
    const wrapper = mount(TeacherStudentAnalysis, {
      global: {
        stubs: {
          SkillRadar: true,
          TeacherClassReportExportDialog: reportDialogStub,
        },
      },
    })

    await flushPromises()

    await wrapper
      .findAll('button')
      .find((button) => button.text().includes('导出班级报告'))
      ?.trigger('click')
    await flushPromises()

    const dialog = wrapper.get('[data-testid="class-report-dialog"]')
    expect(dialog.attributes('data-open')).toBe('true')
    expect(dialog.attributes('data-default-class-name')).toBe('Class A')
    expect(pushMock).not.toHaveBeenCalledWith({ name: 'TeacherAWDReviewIndex' })
  })

  it('导出复盘归档失败时不应抛到全局错误页', async () => {
    teacherApiMocks.exportStudentReviewArchive.mockRejectedValue(new Error('导出失败'))

    const wrapper = mount(TeacherStudentAnalysis, {
      global: {
        stubs: {
          TeacherClassReportExportDialog: reportDialogStub,
          StudentAnalysisPage: {
            name: 'StudentAnalysisPage',
            template:
              '<button id="export-review-archive" type="button" @click="$emit(\'exportReviewArchive\')">导出复盘归档</button>',
          },
        },
      },
    })

    await flushPromises()

    await expect(wrapper.get('#export-review-archive').trigger('click')).resolves.toBeUndefined()
    await flushPromises()

    expect(teacherApiMocks.exportStudentReviewArchive).toHaveBeenCalledWith('stu-1', {
      format: 'json',
    })
  })

  it('管理员从学员分析返回班级管理时应回到后台班级页', async () => {
    const authStore = useAuthStore()
    authStore.setAuth(
      {
        id: 'admin-1',
        username: 'admin',
        role: 'admin',
        class_name: 'Class A',
      })

    const wrapper = mount(TeacherStudentAnalysis, {
      global: {
        stubs: {
          SkillRadar: true,
          TeacherClassReportExportDialog: reportDialogStub,
        },
      },
    })

    await flushPromises()

    wrapper.findComponent({ name: 'StudentAnalysisPage' }).vm.$emit('openClassManagement')

    expect(pushMock).toHaveBeenCalledWith({ name: 'PlatformClassManagement' })
  })

  it('管理员在学员分析内继续切换学生链路时应停留在后台路由', async () => {
    const authStore = useAuthStore()
    authStore.setAuth(
      {
        id: 'admin-1',
        username: 'admin',
        role: 'admin',
        class_name: 'Class A',
      })

    const wrapper = mount(TeacherStudentAnalysis, {
      global: {
        stubs: {
          SkillRadar: true,
          TeacherClassReportExportDialog: reportDialogStub,
        },
      },
    })

    await flushPromises()

    wrapper.findComponent({ name: 'StudentAnalysisPage' }).vm.$emit('openClassStudents')
    wrapper.findComponent({ name: 'StudentAnalysisPage' }).vm.$emit('openReviewArchive')

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
  })
})
