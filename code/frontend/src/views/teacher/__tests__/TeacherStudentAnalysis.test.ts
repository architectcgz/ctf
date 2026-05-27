import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { reactive } from 'vue'

import TeacherStudentAnalysis from '../TeacherStudentAnalysis.vue'
import teacherStudentAnalysisSource from '../TeacherStudentAnalysis.vue?raw'
import studentAnalysisPageModelSource from '@/features/student-analysis-workspace/model/useStudentAnalysisPage.ts?raw'
import studentAnalysisPageSource from '@/components/teacher/class-management/StudentAnalysisPage.vue?raw'
import studentAnalysisOverviewHeroPanelSource from '@/components/teacher/class-management/StudentAnalysisOverviewHeroPanel.vue?raw'
import studentInsightPanelSource from '@/components/teacher/StudentInsightPanel.vue?raw'
import studentInsightAttackSessionsSectionSource from '@/components/teacher/student-insight/StudentInsightAttackSessionsSection.vue?raw'
import studentInsightOverviewSectionSource from '@/components/teacher/student-insight/StudentInsightOverviewSection.vue?raw'
import studentInsightRecommendationsSectionSource from '@/components/teacher/student-insight/StudentInsightRecommendationsSection.vue?raw'
import { useAuthStore } from '@/stores/auth'
import {
  reportDialogStub,
  resetStudentAnalysisRouteTestState,
  type StudentAnalysisRouteTestState,
  type StudentAnalysisTeachingApiMocks,
} from '@/views/__tests__/studentAnalysisRouteTestSupport'

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

async function openWorkspaceTab(
  wrapper: ReturnType<typeof mount>,
  tabKey: 'overview' | 'recommendations' | 'writeups' | 'evidence' | 'timeline'
): Promise<void> {
  await wrapper.get(`#student-tab-${tabKey}`).trigger('click')
  await flushPromises()
}

describe('TeacherStudentAnalysis', () => {
  beforeEach(() => {
    resetStudentAnalysisRouteTestState({
      pushMock,
      replaceMock,
      routeMock,
      teachingApiMocks,
      role: 'teacher',
    })
  })

  it('应该展示当前学员分析内容', async () => {
    const wrapper = mount(TeacherStudentAnalysis, {
      global: {
        stubs: {
          SkillRadar: true,
          ClassReportExportDialog: reportDialogStub,
        },
      },
    })

    await flushPromises()

    expect(wrapper.text()).toContain('alice')
    expect(wrapper.text()).toContain('50%')
    await openWorkspaceTab(wrapper, 'recommendations')
    expect(wrapper.text()).toContain('crypto-lab')

    await openWorkspaceTab(wrapper, 'timeline')
    expect(wrapper.text()).toContain('web-1')
    expect(wrapper.text()).toContain('查看题目详情')
    expect(wrapper.text()).toContain('解锁第 1 级提示')
    expect(wrapper.text()).toContain('访问攻击目标')
    expect(wrapper.text()).toContain('延长实例有效期')
    expect(wrapper.text()).toContain('第 2 次提交命中 Flag')

    await openWorkspaceTab(wrapper, 'evidence')
    expect(wrapper.text()).toContain('复盘工作台')
    expect(wrapper.text()).toContain('会话数')
    expect(wrapper.text()).toContain('事件数')
    expect(wrapper.text()).toContain('实操请求')
    expect(wrapper.text()).toContain('POST /login')

    await openWorkspaceTab(wrapper, 'writeups')
    expect(wrapper.text()).toContain('题解列表')
    expect(wrapper.text()).toContain('misc-essay')
    expect(wrapper.text()).toContain('从回显到 flag')
    expect(wrapper.text()).toContain('社区题解状态')
    expect(wrapper.text()).toContain('审核状态')
    expect(wrapper.text()).toContain('查看审核')
    expect(wrapper.text()).toContain('推荐题解')
    expect(wrapper.text()).toContain('已公开')
    expect(wrapper.text()).toContain('取消推荐')

    expect(teachingApiMocks.getStudentEvidence).toHaveBeenCalledWith('stu-1', {})
    expect(teachingApiMocks.getStudentAttackSessions).toHaveBeenCalledWith('stu-1', {
      with_events: true,
      limit: 20,
      offset: 0,
    })
    expect(teachingApiMocks.getTeacherWriteupSubmissions).toHaveBeenCalledWith({
      student_id: 'stu-1',
      submission_status: 'published',
      page: 1,
      page_size: 6,
    })
    expect(teachingApiMocks.getTeacherManualReviewSubmissions).toHaveBeenCalledWith({
      student_id: 'stu-1',
      page_size: 6,
    })
  })

  it('班级列表接口失败不应阻断学员分析加载', async () => {
    teachingApiMocks.getClasses.mockRejectedValue(new Error('班级列表失败'))

    const wrapper = mount(TeacherStudentAnalysis, {
      global: {
        stubs: {
          SkillRadar: true,
          ClassReportExportDialog: reportDialogStub,
        },
      },
    })

    await flushPromises()

    expect(teachingApiMocks.getClasses).not.toHaveBeenCalled()
    expect(teachingApiMocks.getClassStudents).toHaveBeenCalledWith('Class A')
    expect(wrapper.text()).not.toContain('学员分析加载失败')
    expect(wrapper.text()).toContain('alice')
    expect(wrapper.text()).toContain('50%')
  })

  it('应只渲染当前激活 tab 的详情内容，切换后再显示对应区块', async () => {
    const wrapper = mount(TeacherStudentAnalysis, {
      global: {
        stubs: {
          SkillRadar: true,
          ClassReportExportDialog: reportDialogStub,
        },
      },
    })

    await flushPromises()

    expect(wrapper.text()).not.toContain('crypto-lab')
    expect(wrapper.text()).not.toContain('题解列表')
    expect(wrapper.text()).not.toContain('POST /login')

    await openWorkspaceTab(wrapper, 'recommendations')
    expect(wrapper.text()).toContain('crypto-lab')
    expect(wrapper.text()).not.toContain('题解列表')
    expect(wrapper.text()).not.toContain('POST /login')

    await openWorkspaceTab(wrapper, 'writeups')
    expect(wrapper.text()).toContain('题解列表')
    expect(wrapper.text()).not.toContain('crypto-lab')
    expect(wrapper.text()).not.toContain('POST /login')

    await openWorkspaceTab(wrapper, 'evidence')
    expect(wrapper.text()).toContain('POST /login')
    expect(wrapper.text()).not.toContain('crypto-lab')
    expect(wrapper.text()).not.toContain('题解列表')
  })

  it('路由页应仅负责组合，不直接处理路由解析逻辑', () => {
    expect(teacherStudentAnalysisSource).toContain('useStudentAnalysisPage')
    expect(teacherStudentAnalysisSource).toContain(
      "import { StudentAnalysisPage } from '@/components/class-management'"
    )
    expect(teacherStudentAnalysisSource).toContain(
      "import { ClassReportExportDialog } from '@/components/teacher/reports'"
    )
    expect(teacherStudentAnalysisSource).not.toContain('resolveClassManagementRouteName')
    expect(teacherStudentAnalysisSource).not.toContain('resolveClassStudentsRouteName')
    expect(teacherStudentAnalysisSource).not.toContain(
      '@/components/teacher/class-management/StudentAnalysisPage.vue'
    )
    expect(teacherStudentAnalysisSource).not.toContain('ClassReportExportDialog.vue')
    expect(teacherStudentAnalysisSource).not.toContain(':classes="classes"')
    expect(teacherStudentAnalysisSource).not.toContain(':students="students"')
    expect(teacherStudentAnalysisSource).not.toContain(':selected-class-name="selectedClassName"')
    expect(teacherStudentAnalysisSource).not.toContain(':selected-student-id="selectedStudentId"')
    expect(teacherStudentAnalysisSource).not.toContain(':loading-classes="loadingClasses"')
    expect(teacherStudentAnalysisSource).not.toContain(':loading-students="loadingStudents"')
    expect(teacherStudentAnalysisSource).not.toContain('@open-class-management="openClassManagement"')
    expect(teacherStudentAnalysisSource).not.toContain('@select-class="selectClass"')
    expect(teacherStudentAnalysisSource).not.toContain('@select-student="selectStudent"')
    expect(studentAnalysisPageModelSource).toContain('useReviewWorkspace()')
    expect(studentAnalysisPageModelSource).toContain('useSubmissionReviewFlows({')
    expect(studentAnalysisPageModelSource).not.toContain('useTeacherReviewWorkspace')
    expect(studentAnalysisPageModelSource).not.toContain('useTeacherSubmissionReviewFlows')
    expect(studentAnalysisPageModelSource).not.toContain('openClassManagement')
    expect(studentAnalysisPageModelSource).not.toContain('selectClass')
    expect(studentAnalysisPageModelSource).not.toContain('selectStudent')
  })

  it('路由页应提供可供 Transition 动画使用的单一元素根节点', () => {
    expect(teacherStudentAnalysisSource).toContain('class="teacher-route-root"')
    expect(teacherStudentAnalysisSource).toMatch(
      /<template>\s*<section class="teacher-route-root">[\s\S]*<StudentAnalysisPage[\s\S]*<ClassReportExportDialog[\s\S]*<\/section>\s*<\/template>/s
    )
  })

  it('学员分析头部应只保留姓名标题，不重复显示英文 eyebrow 和用户名 chip', () => {
    expect(studentAnalysisPageSource).not.toContain('Student Analysis')
    expect(studentAnalysisPageSource).not.toContain('teacher-student-chip')
    expect(studentAnalysisPageSource).not.toContain('teacher-eyebrow-row')
    expect(studentAnalysisPageSource).toContain('StudentAnalysisOverviewHeroPanel')
    expect(studentAnalysisPageSource).not.toContain('classes: ClassDirectoryItem[]')
    expect(studentAnalysisPageSource).not.toContain('students: TeacherStudentItem[]')
    expect(studentAnalysisPageSource).not.toContain('selectedClassName: string')
    expect(studentAnalysisPageSource).not.toContain('selectedStudentId: string')
    expect(studentAnalysisPageSource).not.toContain('loadingClasses: boolean')
    expect(studentAnalysisPageSource).not.toContain('loadingStudents: boolean')
    expect(studentAnalysisPageSource).not.toContain('openClassManagement: []')
    expect(studentAnalysisPageSource).not.toContain('selectClass: [className: string]')
    expect(studentAnalysisPageSource).not.toContain('selectStudent: [studentId: string]')
    expect(studentAnalysisPageSource).not.toContain('<span>已做题目数</span>')
    expect(studentAnalysisOverviewHeroPanelSource).toContain(
      "{{ selectedStudent?.name || selectedStudent?.username || '学员分析' }}"
    )
    expect(studentAnalysisOverviewHeroPanelSource).toContain('<span>已做题目数</span>')
    expect(studentAnalysisOverviewHeroPanelSource).toContain('<CheckCircle class="h-4 w-4" />')
    expect(studentAnalysisOverviewHeroPanelSource).toContain('<span>完成率</span>')
    expect(studentAnalysisOverviewHeroPanelSource).toContain('<Trophy class="h-4 w-4" />')
    expect(studentAnalysisOverviewHeroPanelSource).toContain('<span>薄弱维度</span>')
    expect(studentAnalysisOverviewHeroPanelSource).toContain('<AlertTriangle class="h-4 w-4" />')
    expect(studentAnalysisOverviewHeroPanelSource).toContain('导出班级报告')
    expect(studentAnalysisOverviewHeroPanelSource).toContain('完整复盘页')
  })

  it('学员详情面板应通过 section 组件装配复盘区，而不是直接依赖 review workspace widget', () => {
    expect(studentInsightPanelSource).toContain('StudentInsightOverviewSection')
    expect(studentInsightPanelSource).toContain('StudentInsightRecommendationsSection')
    expect(studentInsightPanelSource).toContain('StudentInsightAttackSessionsSection')
    expect(studentInsightPanelSource).not.toContain('TeacherStudentReviewWorkspace')
    expect(studentInsightOverviewSectionSource).toContain('<SkillRadar :scores="radarScores" />')
    expect(studentInsightRecommendationsSectionSource).toContain(
      'class="insight-recommendation-list workspace-directory-list"'
    )
  })

  it('复盘区 section 应通过共享 widget 公共出口消费中性符号', () => {
    expect(studentInsightAttackSessionsSectionSource).toContain(
      "import { StudentReviewWorkspace } from '@/widgets/teacher-student-review-workspace'"
    )
    expect(studentInsightAttackSessionsSectionSource).not.toContain('TeacherStudentReviewWorkspace')
  })

  it('应该支持隐藏社区题解', async () => {
    const wrapper = mount(TeacherStudentAnalysis, {
      global: {
        stubs: {
          SkillRadar: true,
          ClassReportExportDialog: reportDialogStub,
        },
      },
    })

    await flushPromises()
    await openWorkspaceTab(wrapper, 'writeups')

    const hideButton = wrapper
      .findAll('button')
      .find((button) => button.text().includes('隐藏题解'))
    expect(hideButton).toBeDefined()

    await hideButton?.trigger('click')
    await flushPromises()

    expect(teachingApiMocks.hideTeacherCommunityWriteup).toHaveBeenCalledWith('writeup-1')
    expect(teachingApiMocks.getTeacherWriteupSubmissions).toHaveBeenCalledTimes(2)
  })

  it('应在题解列表内打开并处理人工审核提交', async () => {
    const wrapper = mount(TeacherStudentAnalysis, {
      global: {
        stubs: {
          SkillRadar: true,
          ClassReportExportDialog: reportDialogStub,
        },
      },
    })

    await flushPromises()
    await openWorkspaceTab(wrapper, 'writeups')

    const reviewButton = wrapper
      .findAll('button')
      .find((button) => button.text().includes('查看审核'))

    expect(reviewButton).toBeDefined()

    await reviewButton?.trigger('click')
    await flushPromises()

    expect(teachingApiMocks.getTeacherManualReviewSubmission).toHaveBeenCalledWith('manual-1')
    expect(wrapper.text()).toContain('完整答案正文')

    const approveButton = wrapper
      .findAll('button')
      .find((button) => button.text().includes('审核通过'))

    expect(approveButton).toBeDefined()

    await approveButton?.trigger('click')
    await flushPromises()

    expect(teachingApiMocks.reviewTeacherManualReviewSubmission).toHaveBeenCalledWith('manual-1', {
      review_status: 'approved',
      review_comment: undefined,
    })
  })

  it('应该支持包含百分号的班级名路由参数', async () => {
    routeMock.params.className = '100% 班级'

    mount(TeacherStudentAnalysis, {
      global: {
        stubs: {
          SkillRadar: true,
          ClassReportExportDialog: reportDialogStub,
        },
      },
    })

    await flushPromises()

    expect(teachingApiMocks.getClassStudents).toHaveBeenCalledWith('100% 班级')
  })

  it('应该支持跳转到完整复盘页', async () => {
    const wrapper = mount(TeacherStudentAnalysis, {
      global: {
        stubs: {
          SkillRadar: true,
          ClassReportExportDialog: reportDialogStub,
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

  it('切换复盘筛选时应只刷新攻击会话查询', async () => {
    const wrapper = mount(TeacherStudentAnalysis, {
      global: {
        stubs: {
          SkillRadar: true,
          ClassReportExportDialog: reportDialogStub,
        },
      },
    })

    await flushPromises()
    teachingApiMocks.getStudentAttackSessions.mockClear()
    await openWorkspaceTab(wrapper, 'evidence')

    const selects = wrapper.findAll('select')
    await selects[1].setValue('awd')
    await flushPromises()
    await selects[2].setValue('failed')
    await flushPromises()

    expect(teachingApiMocks.getStudentAttackSessions).toHaveBeenNthCalledWith(1, 'stu-1', {
      with_events: true,
      limit: 20,
      offset: 0,
      mode: 'awd',
    })
    expect(teachingApiMocks.getStudentAttackSessions).toHaveBeenNthCalledWith(2, 'stu-1', {
      with_events: true,
      limit: 20,
      offset: 0,
      mode: 'awd',
      result: 'failed',
    })
    expect(teachingApiMocks.getStudentEvidence).toHaveBeenCalledTimes(1)
    expect(replaceMock).toHaveBeenNthCalledWith(1, {
      query: {
        reviewMode: 'awd',
        reviewResult: undefined,
        reviewChallengeId: undefined,
      },
    })
    expect(replaceMock).toHaveBeenNthCalledWith(2, {
      query: {
        reviewMode: 'awd',
        reviewResult: 'failed',
        reviewChallengeId: undefined,
      },
    })
  })

  it('切换题目筛选时应刷新证据和攻击会话，并同步到路由 query', async () => {
    const wrapper = mount(TeacherStudentAnalysis, {
      global: {
        stubs: {
          SkillRadar: true,
          ClassReportExportDialog: reportDialogStub,
        },
      },
    })

    await flushPromises()
    teachingApiMocks.getStudentAttackSessions.mockClear()
    teachingApiMocks.getStudentEvidence.mockClear()
    await openWorkspaceTab(wrapper, 'evidence')

    const selects = wrapper.findAll('select')
    await selects[0].setValue('11')
    await flushPromises()

    expect(teachingApiMocks.getStudentEvidence).toHaveBeenCalledWith('stu-1', {
      challenge_id: '11',
    })
    expect(teachingApiMocks.getStudentAttackSessions).toHaveBeenCalledWith('stu-1', {
      with_events: true,
      limit: 20,
      offset: 0,
      challenge_id: '11',
    })
    expect(replaceMock).toHaveBeenCalledWith({
      query: {
        reviewMode: undefined,
        reviewResult: undefined,
        reviewChallengeId: '11',
      },
    })
  })

  it('路由 query 回退到新的复盘筛选时应只刷新复盘区', async () => {
    mount(TeacherStudentAnalysis, {
      global: {
        stubs: {
          SkillRadar: true,
          ClassReportExportDialog: reportDialogStub,
        },
      },
    })

    await flushPromises()
    teachingApiMocks.getStudentEvidence.mockClear()
    teachingApiMocks.getStudentAttackSessions.mockClear()
    teachingApiMocks.getStudentProgress.mockClear()
    teachingApiMocks.getStudentTimeline.mockClear()

    routeMock.query = {
      reviewMode: 'awd',
      reviewResult: 'failed',
    }
    await flushPromises()

    expect(teachingApiMocks.getStudentAttackSessions).toHaveBeenCalledWith('stu-1', {
      with_events: true,
      limit: 20,
      offset: 0,
      mode: 'awd',
      result: 'failed',
    })
    expect(teachingApiMocks.getStudentEvidence).not.toHaveBeenCalled()
    expect(teachingApiMocks.getStudentProgress).not.toHaveBeenCalled()
    expect(teachingApiMocks.getStudentTimeline).not.toHaveBeenCalled()
  })

  it('应采用顶部 tabs 工作区壳层而不是把所有内容堆叠在主页面，并去掉页面内重复顶栏', async () => {
    const wrapper = mount(TeacherStudentAnalysis, {
      global: {
        stubs: {
          SkillRadar: true,
          ClassReportExportDialog: reportDialogStub,
        },
      },
    })

    await flushPromises()

    expect(wrapper.find('[role="tablist"]').exists()).toBe(true)
    expect(wrapper.find('#student-tab-overview').exists()).toBe(true)
    expect(wrapper.find('#student-tab-recommendations').exists()).toBe(true)
    expect(wrapper.find('#student-tab-writeups').exists()).toBe(true)
    expect(wrapper.find('#student-tab-manual-review').exists()).toBe(false)
    expect(wrapper.find('#student-tab-evidence').exists()).toBe(true)
    expect(wrapper.find('#student-tab-timeline').exists()).toBe(true)
    expect(studentAnalysisPageSource).toMatch(/class="[^"]*\bworkspace-shell\b[^"]*"/)
    expect(studentAnalysisPageSource).not.toContain('class="workspace-topbar"')
    expect(studentAnalysisPageSource).toContain('class="workspace-tabbar top-tabs"')
    expect(studentAnalysisPageSource).toContain('class="content-pane"')
    expect(studentAnalysisPageSource).toMatch(
      /<div class="[^"]*\bworkspace-shell\b[^"]*">[\s\S]*<nav class="workspace-tabbar top-tabs"[\s\S]*<main class="content-pane">/s
    )
  })

  it('点击导出班级报告时应打开当前班级上下文对话框', async () => {
    const wrapper = mount(TeacherStudentAnalysis, {
      global: {
        stubs: {
          SkillRadar: true,
          ClassReportExportDialog: reportDialogStub,
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
    teachingApiMocks.exportStudentReviewArchive.mockRejectedValue(new Error('导出失败'))

    const wrapper = mount(TeacherStudentAnalysis, {
      global: {
        stubs: {
          ClassReportExportDialog: reportDialogStub,
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

    expect(teachingApiMocks.exportStudentReviewArchive).toHaveBeenCalledWith('stu-1', {
      format: 'json',
    })
  })

  it('管理员在学员分析内继续切换学生链路时应停留在后台路由', async () => {
    const authStore = useAuthStore()
    authStore.setAuth({
      id: 'admin-1',
      username: 'admin',
      role: 'admin',
      class_name: 'Class A',
    })

    const wrapper = mount(TeacherStudentAnalysis, {
      global: {
        stubs: {
          SkillRadar: true,
          ClassReportExportDialog: reportDialogStub,
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
