import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { flushPromises, mount } from '@vue/test-utils'
import { useAuthStore } from '@/stores/auth'

import TeacherClassStudents from '../TeacherClassStudents.vue'
import teacherClassStudentsSource from '../TeacherClassStudents.vue?raw'
import classStudentsPageSource from '@/components/teacher/class-management/ClassStudentsPage.vue?raw'

const ElTable = { template: '<div><slot /></div>' }
const ElTableColumn = { template: '<div><slot /></div>' }
const ElButton = { template: '<button><slot /></button>' }

const pushMock = vi.fn()
const replaceMock = vi.fn()
const routeMock = {
  params: {
    className: 'Class A',
  },
  query: {
    panel: 'students',
  },
}

const teacherApiMocks = vi.hoisted(() => ({
  getClasses: vi.fn(),
  getClassStudents: vi.fn(),
  getClassReview: vi.fn(),
  getClassSummary: vi.fn(),
  getClassTrend: vi.fn(),
  getStudentRecommendations: vi.fn(),
}))

vi.mock('vue-router', async () => {
  const actual = await vi.importActual<typeof import('vue-router')>('vue-router')
  return {
    ...actual,
    useRouter: () => ({ push: pushMock, replace: replaceMock }),
    useRoute: () => routeMock,
  }
})

vi.mock('@/api/teacher', () => teacherApiMocks)

function deferred<T>() {
  let resolve!: (value: T | PromiseLike<T>) => void
  const promise = new Promise<T>((nextResolve) => {
    resolve = nextResolve
  })
  return { promise, resolve }
}

describe('TeacherClassStudents', () => {
  const reportDialogStub = {
    name: 'TeacherClassReportExportDialog',
    props: ['modelValue', 'defaultClassName', 'defaultFromDate', 'defaultToDate'],
    template:
      '<div data-testid="class-report-dialog" :data-open="String(modelValue)" :data-default-class-name="defaultClassName || \'\'" :data-default-from-date="defaultFromDate || \'\'" :data-default-to-date="defaultToDate || \'\'" />',
  }

  beforeEach(() => {
    setActivePinia(createPinia())
    localStorage.clear()
    pushMock.mockReset()
    replaceMock.mockReset()
    routeMock.params.className = 'Class A'
    routeMock.query.panel = 'students'
    delete routeMock.query.from_date
    delete routeMock.query.to_date
    teacherApiMocks.getClasses.mockReset()
    teacherApiMocks.getClassStudents.mockReset()
    teacherApiMocks.getClassReview.mockReset()
    teacherApiMocks.getClassSummary.mockReset()
    teacherApiMocks.getClassTrend.mockReset()
    teacherApiMocks.getStudentRecommendations.mockReset()

    teacherApiMocks.getClasses.mockResolvedValue([{ name: 'Class A', student_count: 2 }])
    teacherApiMocks.getClassStudents.mockResolvedValue([
      {
        id: 'stu-1',
        username: 'alice',
        name: 'Alice Zhang',
        solved_count: 3,
        total_score: 280,
        recent_event_count: 0,
        weak_dimension: 'crypto',
      },
      {
        id: 'stu-2',
        username: 'bob',
        solved_count: 1,
        total_score: 100,
        recent_event_count: 2,
        weak_dimension: 'pwn',
      },
    ])
    teacherApiMocks.getStudentRecommendations.mockResolvedValue({
      weak_dimensions: [
        {
          dimension: 'crypto',
          label: '密码',
          severity: 'warning',
          confidence: 0.84,
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
    teacherApiMocks.getClassReview.mockResolvedValue({
      class_name: 'Class A',
      items: [
        {
          code: 'activity_risk',
          severity: 'warning',
          summary: '班级活跃度需要补强',
          evidence: 'Class A 近 7 天活跃率为 50%，适合通过定向训练把低活跃学生重新拉回训练节奏。',
        },
        {
          code: 'weak_dimension_cluster',
          severity: 'attention',
          summary: '优先补薄弱维度',
          evidence: 'crypto 是当前最集中的薄弱项，涉及 1 名学生，建议本周统一布置该维度基础题。',
          dimension: 'crypto',
          students: [{ id: 'stu-1', username: 'alice' }],
          recommendation: {
            challenge_id: '12',
            title: 'crypto-lab',
            category: 'crypto',
            difficulty: 'medium',
            summary: '针对薄弱维度：密码',
            evidence: '当前密码维度已经形成高置信度薄弱信号。',
          },
        },
        {
          code: 'focus_students',
          severity: 'attention',
          summary: '先跟进重点学生',
          evidence: '建议教师先跟进 alice，并优先布置推荐题做补强训练。',
          students: [{ id: 'stu-1', username: 'alice' }],
        },
      ],
    })
    teacherApiMocks.getClassSummary.mockResolvedValue({
      class_name: 'Class A',
      student_count: 2,
      average_solved: 2,
      active_student_count: 1,
      active_rate: 50,
      recent_event_count: 6,
    })
    teacherApiMocks.getClassTrend.mockResolvedValue({
      class_name: 'Class A',
      points: [
        { date: '2026-03-05', active_student_count: 1, event_count: 2, solve_count: 1 },
        { date: '2026-03-06', active_student_count: 1, event_count: 3, solve_count: 1 },
      ],
    })

    const authStore = useAuthStore()
    authStore.setAuth({
      id: 'teacher-1',
      username: 'teacher',
      role: 'teacher',
      class_name: 'Class A',
    })
  })

  it('应该展示班级学生列表并支持进入学员分析页', async () => {
    const wrapper = mount(TeacherClassStudents, {
      global: {
        components: {
          ElTable,
          ElTableColumn,
          ElButton,
        },
        stubs: {
          LineChart: true,
          TeacherClassReportExportDialog: reportDialogStub,
        },
      },
    })

    await flushPromises()
    await flushPromises()

    expect(wrapper.text()).toContain('学生列表')
    expect(wrapper.find('.teacher-topbar').exists()).toBe(true)
    expect(wrapper.find('.teacher-summary').exists()).toBe(true)
    expect(wrapper.find('.teacher-directory-shell').exists()).toBe(true)
    expect(wrapper.find('[role="tablist"]').exists()).toBe(true)
    expect(wrapper.find('#class-tab-overview').exists()).toBe(true)
    expect(wrapper.find('#class-tab-trend').exists()).toBe(true)
    expect(wrapper.find('#class-tab-students').exists()).toBe(true)
    expect(wrapper.find('#class-tab-review').exists()).toBe(true)
    expect(wrapper.find('#class-tab-insight').exists()).toBe(true)
    expect(wrapper.find('#class-tab-action').exists()).toBe(true)
    expect(wrapper.text()).toContain('alice')
    expect(wrapper.text()).toContain('bob')
    expect(wrapper.text()).toContain('50%')
    expect(wrapper.text()).toContain('280')
    expect(wrapper.text()).toContain('crypto')
    expect(wrapper.text()).toContain('教学复盘结论')
    expect(wrapper.text()).toContain('班级活跃度需要补强')
    expect(wrapper.text()).toContain('优先补薄弱维度')
    expect(wrapper.text()).toContain('先跟进重点学生')
    expect(wrapper.text()).toContain('班级 Top 学生')
    expect(wrapper.text()).toContain('薄弱维度分布')
    expect(wrapper.text()).toContain('优先介入学生')
    expect(wrapper.text()).toContain('建议训练题')
    expect(wrapper.text()).toContain('推荐训练题')
    expect(wrapper.text()).toContain('crypto-lab')
    await wrapper.find('#class-tab-students').trigger('click')
    const studentsPanel = wrapper.get('#class-students')
    expect(studentsPanel.find('table').exists()).toBe(true)
    expect(studentsPanel.findAll('tbody tr')).toHaveLength(2)
    expect(studentsPanel.text()).toContain('学号')
    expect(studentsPanel.text()).toContain('学生名称')
    expect(studentsPanel.text()).toContain('昵称')
    expect(studentsPanel.text()).toContain('做题数 / 得分数')
    expect(studentsPanel.text()).not.toContain('切换班级')
    expect(studentsPanel.text()).not.toContain('返回列表')
    expect(studentsPanel.find('select').exists()).toBe(false)
    expect(studentsPanel.find('input[placeholder="输入学号精确查询"]').exists()).toBe(true)
    expect(wrapper.find('.teacher-directory-row-title').attributes('title')).toBe('Alice Zhang')
    expect(wrapper.find('.teacher-directory-row-points').attributes('title')).toBe('alice')
    expect(teacherApiMocks.getClassReview).toHaveBeenCalledWith('Class A')
    expect(teacherApiMocks.getStudentRecommendations).toHaveBeenCalledWith('stu-1')

    wrapper.findComponent({ name: 'ClassStudentsPage' }).vm.$emit('openStudent', 'stu-1')

    expect(pushMock).toHaveBeenCalledWith({
      name: 'TeacherStudentAnalysis',
      params: { className: 'Class A', studentId: 'stu-1' },
    })
  })

  it('没有推荐题时应在教师复盘与介入卡片中展示友好提示', async () => {
    teacherApiMocks.getStudentRecommendations.mockResolvedValue({
      weak_dimensions: [],
      challenges: [],
    })
    teacherApiMocks.getClassReview.mockResolvedValue({
      class_name: 'Class A',
      items: [
        {
          code: 'activity_risk',
          severity: 'warning',
          summary: '班级活跃度需要补强',
          evidence: 'Class A 近 7 天活跃率为 50%，需要尽快补回训练节奏。',
          action: '本周先联系低活跃学生，确认是否存在卡点。',
        },
        {
          code: 'retry_cost_high',
          severity: 'warning',
          summary: '部分学生试错成本偏高',
          evidence: 'alice 最近连续错误提交偏多，需要先回放关键利用链路。',
          action: '按验证思路再提交的节奏继续训练。',
          students: [{ id: 'stu-1', username: 'alice' }],
        },
      ],
    })

    const wrapper = mount(TeacherClassStudents, {
      global: {
        components: {
          ElTable,
          ElTableColumn,
          ElButton,
        },
        stubs: {
          LineChart: true,
          TeacherClassReportExportDialog: reportDialogStub,
        },
      },
    })

    await flushPromises()
    await flushPromises()

    expect(wrapper.text()).toContain('当前暂无直接匹配的推荐题，可先跟进这名学生最近的训练卡点。')
    expect(wrapper.text()).toContain('当前没有直接匹配的推荐题，可先按本条结论安排训练。')
    expect(wrapper.text()).not.toContain('crypto-lab')
  })

  it('推荐题加载失败时不应把教师介入卡片伪装成空推荐', async () => {
    teacherApiMocks.getStudentRecommendations.mockRejectedValue(new Error('boom'))

    const wrapper = mount(TeacherClassStudents, {
      global: {
        components: {
          ElTable,
          ElTableColumn,
          ElButton,
        },
        stubs: {
          LineChart: true,
          TeacherClassReportExportDialog: reportDialogStub,
        },
      },
    })

    await flushPromises()
    await flushPromises()

    expect(wrapper.text()).toContain('推荐题暂时没有加载成功，请稍后再试。')
    expect(wrapper.text()).not.toContain(
      '当前暂无直接匹配的推荐题，可先跟进这名学生最近的训练卡点。'
    )
  })

  it('路由页应仅负责组合，不直接依赖教师接口实现', () => {
    expect(teacherClassStudentsSource).toContain('useTeacherClassStudentsPage')
    expect(teacherClassStudentsSource).not.toContain("from '@/api/teacher'")
  })

  it('路由页应提供可供 Transition 动画使用的单一元素根节点', () => {
    expect(teacherClassStudentsSource).toContain('class="teacher-route-root"')
    expect(teacherClassStudentsSource).toMatch(
      /<template>\s*<section class="teacher-route-root">[\s\S]*<ClassStudentsPage[\s\S]*<TeacherClassReportExportDialog[\s\S]*<\/section>\s*<\/template>/s
    )
  })

  it('班级学生薄弱项应复用题目分类胶囊色，并先归一化后判断分类值', () => {
    expect(classStudentsPageSource).toContain('ChallengeCategoryPill')
    expect(classStudentsPageSource).toContain('toChallengeCategory(student.weak_dimension)')
  })

  it('班级详情页应采用与教学概览一致的顶部 tabs 壳层结构，并去掉页面内重复顶栏', () => {
    expect(classStudentsPageSource).toMatch(/class="[^"]*\bworkspace-shell\b[^"]*"/)
    expect(classStudentsPageSource).toMatch(/class="[^"]*\bteacher-management-shell\b[^"]*"/)
    expect(classStudentsPageSource).not.toContain('class="workspace-topbar"')
    expect(classStudentsPageSource).toContain('class="top-tabs"')
    expect(classStudentsPageSource).toContain('class="content-pane"')
    expect(classStudentsPageSource).toContain('WorkspaceDataTable')
    expect(classStudentsPageSource).toContain('class="teacher-topbar class-overview-topbar"')
    expect(classStudentsPageSource).toContain('class="teacher-summary class-overview-summary"')
    expect(classStudentsPageSource).toContain(
      'class="teacher-summary-grid progress-strip metric-panel-grid metric-panel-default-surface"'
    )
    expect(classStudentsPageSource).toContain('class="progress-card metric-panel-card"')
    expect(classStudentsPageSource).toMatch(
      /<div class="[^"]*\bworkspace-shell\b[^"]*">[\s\S]*<nav class="top-tabs"[\s\S]*<main class="content-pane">/s
    )
    expect(classStudentsPageSource).toMatch(
      /class="teacher-directory-row-title"[\s\S]*:title="\(row as ClassStudentDirectoryRow\)\.name"/s
    )
    expect(classStudentsPageSource).toMatch(
      /class="teacher-directory-row-points"[\s\S]*:title="\(row as ClassStudentDirectoryRow\)\.username"/s
    )
    expect(classStudentsPageSource).toMatch(
      /\.teacher-directory-row-title\s*\{[^}]*overflow:\s*hidden;[^}]*text-overflow:\s*ellipsis;[^}]*white-space:\s*nowrap;/s
    )
    expect(classStudentsPageSource).toMatch(
      /\.teacher-directory-row-points\s*\{[^}]*overflow:\s*hidden;[^}]*text-overflow:\s*ellipsis;[^}]*white-space:\s*nowrap;/s
    )
    expect(classStudentsPageSource).not.toContain('class="teacher-section-head"')
    expect(classStudentsPageSource).not.toContain('切换班级')
    expect(classStudentsPageSource).not.toContain("emit('selectClass'")
    expect(classStudentsPageSource).toContain(
      'class="teacher-directory-shell workspace-directory-list"'
    )
    expect(classStudentsPageSource).toContain('class="teacher-student-directory-table"')
    expect(classStudentsPageSource).toContain("label: '做题数 / 得分数'")
    expect(classStudentsPageSource).not.toContain('class="teacher-directory-row-metrics"')
    expect(classStudentsPageSource).not.toMatch(/\.teacher-directory-row-metrics\s*\{/)
    expect(classStudentsPageSource).toContain('当前班级学生总数')
    expect(classStudentsPageSource).toContain('当前班级人均完成题目数')
    expect(classStudentsPageSource).toMatch(/\.class-overview-topbar\s*\{[^}]*border-bottom:\s*0;/s)
    expect(classStudentsPageSource).toMatch(
      /\.class-overview-summary\s*\{[^}]*padding:\s*0;[^}]*border-bottom:\s*0;/s
    )
    expect(classStudentsPageSource).toContain('<span>班级人数</span>')
    expect(classStudentsPageSource).toContain('<Users class="h-4 w-4" />')
    expect(classStudentsPageSource).toContain('<span>平均解题</span>')
    expect(classStudentsPageSource).toContain('<Target class="h-4 w-4" />')
    expect(classStudentsPageSource).toContain('班级训练时间段')
    expect(classStudentsPageSource).toContain('开始日期')
    expect(classStudentsPageSource).toContain('结束日期')
    expect(classStudentsPageSource).toContain('应用时间段')
    expect(classStudentsPageSource).toContain('恢复默认')
    expect(classStudentsPageSource).toContain('<span>当前窗口活跃率</span>')
    expect(classStudentsPageSource).not.toContain('<span>近 7 天活跃率</span>')
    expect(classStudentsPageSource).toContain('<Activity class="h-4 w-4" />')
  })

  it('应该保留已解码的班级名并使用原值请求学生列表', async () => {
    routeMock.params.className = '100% 班级'

    mount(TeacherClassStudents, {
      global: {
        components: {
          ElTable,
          ElTableColumn,
          ElButton,
        },
        stubs: {
          LineChart: true,
          TeacherClassReportExportDialog: reportDialogStub,
        },
      },
    })

    await flushPromises()

    expect(teacherApiMocks.getClassStudents).toHaveBeenCalledWith('100% 班级', {
      student_no: undefined,
    })
    expect(teacherApiMocks.getClassReview).toHaveBeenCalledWith('100% 班级')
    expect(teacherApiMocks.getClassSummary).toHaveBeenCalledWith('100% 班级')
    expect(teacherApiMocks.getClassTrend).toHaveBeenCalledWith('100% 班级')
  })

  it('时间段 query 应驱动班级概览请求，并在应用后回写路由与导出弹窗上下文', async () => {
    routeMock.query.from_date = '2026-03-01'
    routeMock.query.to_date = '2026-03-07'

    const wrapper = mount(TeacherClassStudents, {
      global: {
        components: {
          ElTable,
          ElTableColumn,
          ElButton,
        },
        stubs: {
          LineChart: true,
          TeacherClassReportExportDialog: reportDialogStub,
        },
      },
    })

    await flushPromises()
    await flushPromises()

    expect(teacherApiMocks.getClassReview).toHaveBeenCalledWith('Class A', {
      from_date: '2026-03-01',
      to_date: '2026-03-07',
    })
    expect(teacherApiMocks.getClassSummary).toHaveBeenCalledWith('Class A', {
      from_date: '2026-03-01',
      to_date: '2026-03-07',
    })
    expect(teacherApiMocks.getClassTrend).toHaveBeenCalledWith('Class A', {
      from_date: '2026-03-01',
      to_date: '2026-03-07',
    })

    const reportDialog = wrapper.get('[data-testid="class-report-dialog"]')
    expect(reportDialog.attributes('data-default-from-date')).toBe('2026-03-01')
    expect(reportDialog.attributes('data-default-to-date')).toBe('2026-03-07')

    const dateInputs = wrapper.findAll('input[type="date"]')
    expect(dateInputs).toHaveLength(2)
    await dateInputs[0].setValue('2026-03-05')
    await dateInputs[1].setValue('2026-03-09')

    await wrapper
      .findAll('button')
      .find((button) => button.text().includes('应用时间段'))
      ?.trigger('click')

    expect(replaceMock).toHaveBeenCalledWith({
      query: {
        panel: 'students',
        from_date: '2026-03-05',
        to_date: '2026-03-09',
      },
    })
  })

  it('管理员从班级详情返回班级管理时应回到后台班级页', async () => {
    const authStore = useAuthStore()
    authStore.setAuth({
      id: 'admin-1',
      username: 'admin',
      role: 'admin',
      class_name: 'Class A',
    })

    const wrapper = mount(TeacherClassStudents, {
      global: {
        components: {
          ElTable,
          ElTableColumn,
          ElButton,
        },
        stubs: {
          LineChart: true,
          TeacherClassReportExportDialog: reportDialogStub,
        },
      },
    })

    await flushPromises()
    await flushPromises()

    wrapper.findComponent({ name: 'ClassStudentsPage' }).vm.$emit('openClassManagement')

    expect(pushMock).toHaveBeenCalledWith({ name: 'PlatformClassManagement' })
  })

  it('管理员在班级详情内查看学生和返回概览时应停留在后台路由', async () => {
    const authStore = useAuthStore()
    authStore.setAuth({
      id: 'admin-1',
      username: 'admin',
      role: 'admin',
      class_name: 'Class A',
    })

    const wrapper = mount(TeacherClassStudents, {
      global: {
        components: {
          ElTable,
          ElTableColumn,
          ElButton,
        },
        stubs: {
          LineChart: true,
          TeacherClassReportExportDialog: reportDialogStub,
        },
      },
    })

    await flushPromises()
    await flushPromises()

    wrapper.findComponent({ name: 'ClassStudentsPage' }).vm.$emit('openStudent', 'stu-1')
    wrapper.findComponent({ name: 'ClassStudentsPage' }).vm.$emit('openDashboard')
    expect(pushMock).toHaveBeenCalledWith({
      name: 'PlatformStudentAnalysis',
      params: { className: 'Class A', studentId: 'stu-1' },
    })
    expect(pushMock).toHaveBeenCalledWith({ name: 'PlatformOverview' })
  })

  it('应该忽略过期学号搜索请求的返回结果', async () => {
    const slowRequest = deferred<
      Array<{
        id: string
        username: string
        name?: string
        solved_count?: number
        total_score?: number
        recent_event_count?: number
        weak_dimension?: string
      }>
    >()
    const fastRequest = deferred<
      Array<{
        id: string
        username: string
        name?: string
        solved_count?: number
        total_score?: number
        recent_event_count?: number
        weak_dimension?: string
      }>
    >()

    teacherApiMocks.getClassStudents.mockReset()
    teacherApiMocks.getClassReview.mockReset()
    teacherApiMocks.getStudentRecommendations.mockReset()
    teacherApiMocks.getClassSummary.mockReset()
    teacherApiMocks.getClassTrend.mockReset()
    teacherApiMocks.getClassStudents
      .mockResolvedValueOnce([
        {
          id: 'stu-1',
          username: 'alice',
          name: 'Alice Zhang',
          solved_count: 3,
          total_score: 280,
          recent_event_count: 0,
          weak_dimension: 'crypto',
        },
        {
          id: 'stu-2',
          username: 'bob',
          solved_count: 1,
          total_score: 100,
          recent_event_count: 2,
          weak_dimension: 'pwn',
        },
      ])
      .mockImplementationOnce(() => slowRequest.promise)
      .mockImplementationOnce(() => fastRequest.promise)
    teacherApiMocks.getClassSummary.mockResolvedValue({
      class_name: 'Class A',
      student_count: 2,
      average_solved: 2,
      active_student_count: 1,
      active_rate: 50,
      recent_event_count: 6,
    })
    teacherApiMocks.getClassTrend.mockResolvedValue({
      class_name: 'Class A',
      points: [
        { date: '2026-03-05', active_student_count: 1, event_count: 2, solve_count: 1 },
        { date: '2026-03-06', active_student_count: 1, event_count: 3, solve_count: 1 },
      ],
    })
    teacherApiMocks.getStudentRecommendations.mockResolvedValue({
      weak_dimensions: [
        {
          dimension: 'crypto',
          label: '密码',
          severity: 'warning',
          confidence: 0.84,
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
    teacherApiMocks.getClassReview.mockResolvedValue({
      class_name: 'Class A',
      items: [
        {
          code: 'activity_risk',
          severity: 'warning',
          summary: '班级活跃度需要补强',
          evidence: 'Class A 近 7 天活跃率为 50%，适合通过定向训练把低活跃学生重新拉回训练节奏。',
        },
      ],
    })

    const wrapper = mount(TeacherClassStudents, {
      global: {
        components: {
          ElTable,
          ElTableColumn,
          ElButton,
        },
        stubs: {
          LineChart: true,
          TeacherClassReportExportDialog: reportDialogStub,
        },
      },
    })

    await flushPromises()
    await wrapper.find('#class-tab-students').trigger('click')

    const studentNoInput = wrapper.find('input[placeholder="输入学号精确查询"]')
    await studentNoInput.setValue('20260001')
    await studentNoInput.setValue('20260002')

    fastRequest.resolve([
      {
        id: 'stu-1',
        username: 'alice',
        name: 'Alice Zhang',
        solved_count: 3,
        total_score: 280,
        recent_event_count: 0,
        weak_dimension: 'crypto',
      },
    ])
    await flushPromises()

    slowRequest.resolve([
      {
        id: 'stu-2',
        username: 'bob',
        solved_count: 1,
        total_score: 100,
        recent_event_count: 2,
        weak_dimension: 'pwn',
      },
    ])
    await flushPromises()

    expect(wrapper.text()).toContain('alice')
    expect(wrapper.text()).not.toContain('bob')
  })

  it('按学号筛选时应该只刷新学生列表，不重复请求班级概览数据', async () => {
    const wrapper = mount(TeacherClassStudents, {
      global: {
        components: {
          ElTable,
          ElTableColumn,
          ElButton,
        },
        stubs: {
          LineChart: true,
          TeacherClassReportExportDialog: reportDialogStub,
        },
      },
    })

    await flushPromises()
    await flushPromises()

    await wrapper.find('#class-tab-students').trigger('click')

    expect(wrapper.text()).not.toContain('学生筛选')

    const studentNoInput = wrapper.find('input[placeholder="输入学号精确查询"]')
    expect(studentNoInput.exists()).toBe(true)

    await studentNoInput.setValue('20260001')
    await flushPromises()

    expect(wrapper.find('.teacher-filter-reset').exists()).toBe(true)
    expect(wrapper.find('.teacher-filter-clear').exists()).toBe(true)

    await wrapper.find('.teacher-filter-clear').trigger('click')
    await flushPromises()

    expect((studentNoInput.element as HTMLInputElement).value).toBe('')
    expect(wrapper.find('.teacher-filter-reset').exists()).toBe(false)

    expect(teacherApiMocks.getClassStudents).toHaveBeenCalledWith('Class A', {
      student_no: '20260001',
    })
    expect(teacherApiMocks.getClassReview).toHaveBeenCalledTimes(1)
    expect(teacherApiMocks.getClassSummary).toHaveBeenCalledTimes(1)
    expect(teacherApiMocks.getClassTrend).toHaveBeenCalledTimes(1)
  })

  it('点击导出班级报告时应打开当前班级上下文对话框', async () => {
    const wrapper = mount(TeacherClassStudents, {
      global: {
        components: {
          ElTable,
          ElTableColumn,
          ElButton,
        },
        stubs: {
          LineChart: true,
          TeacherClassReportExportDialog: reportDialogStub,
        },
      },
    })

    await flushPromises()
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
})
