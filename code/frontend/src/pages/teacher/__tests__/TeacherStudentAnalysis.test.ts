import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { reactive } from 'vue'

import TeacherStudentAnalysis from '@/pages/teacher/TeacherStudentAnalysisRoutePage.vue'
import teacherStudentAnalysisSource from '@/pages/teacher/TeacherStudentAnalysisRoutePage.vue?raw'
import studentAnalysisNavigationSource from '@/features/teaching/student-analysis-workspace/model/useStudentAnalysisNavigation.ts?raw'
import studentAnalysisPageModelSource from '@/features/teaching/student-analysis-workspace/model/useStudentAnalysisPage.ts?raw'
import studentAnalysisReviewQuerySyncSource from '@/features/teaching/student-analysis-workspace/model/useStudentAnalysisReviewQuerySync.ts?raw'
import studentAnalysisPageSource from '@/features/teaching/student-analysis-workspace/ui/StudentAnalysisPage.vue?raw'
import studentAnalysisWorkspacePageSource from '@/features/teaching/student-analysis-workspace/ui/StudentAnalysisWorkspacePage.vue?raw'
import studentAnalysisWorkspaceContentSource from '@/features/teaching/student-analysis-workspace/ui/StudentAnalysisWorkspaceContent.vue?raw'
import studentAnalysisWorkspaceTabsSource from '@/features/teaching/student-analysis-workspace/ui/StudentAnalysisWorkspaceTabs.vue?raw'
import studentAnalysisWorkspaceTabsHelperSource from '@/features/teaching/student-analysis-workspace/ui/studentAnalysisWorkspaceTabs.ts?raw'
import studentAnalysisOverviewHeroPanelSource from '@/features/teaching/student-analysis-workspace/ui/StudentAnalysisOverviewHeroPanel.vue?raw'
import studentInsightLoadingSurfaceSource from '@/features/teaching/student-analysis-shared/ui/StudentInsightLoadingSurface.vue?raw'
import studentInsightPanelSource from '@/features/teaching/student-analysis-workspace/ui/StudentInsightPanel.vue?raw'
import studentInsightPrimarySectionsSource from '@/features/teaching/student-analysis-workspace/ui/StudentInsightPrimarySections.vue?raw'
import studentInsightTimelineSectionSource from '@/features/teaching/student-analysis-workspace/ui/StudentInsightTimelineSection.vue?raw'
import studentInsightAttackSessionsSectionSource from '@/features/teaching/student-analysis-review/ui/StudentInsightAttackSessionsSection.vue?raw'
import studentInsightOverviewSectionSource from '@/features/teaching/student-analysis-workspace/ui/StudentInsightOverviewSection.vue?raw'
import studentInsightRecommendationsSectionSource from '@/features/teaching/student-analysis-workspace/ui/StudentInsightRecommendationsSection.vue?raw'
import studentInsightReviewSectionsSource from '@/features/teaching/student-analysis-review/ui/StudentInsightReviewSections.vue?raw'
import studentInsightStateSurfaceSource from '@/features/teaching/student-analysis-shared/ui/StudentInsightStateSurface.vue?raw'
import writeupsSectionSource from '@/features/teaching/student-analysis-review/ui/StudentInsightWriteupsSection.vue?raw'
import studentReviewWorkspaceSource from '@/features/teaching/student-analysis-review/ui/StudentReviewWorkspace.vue?raw'
import { useAuthStore } from '@/stores/auth'
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

async function openWorkspaceTab(
  wrapper: ReturnType<typeof mount>,
  tabKey: 'overview' | 'recommendations' | 'writeups' | 'evidence' | 'training-records'
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

    await openWorkspaceTab(wrapper, 'training-records')
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

  it('应兼容旧的 timeline query 并归一到 training-records', async () => {
    routeMock.query = { panel: 'timeline' }

    const wrapper = mount(TeacherStudentAnalysis, {
      global: {
        stubs: {
          SkillRadar: true,
          ClassReportExportDialog: reportDialogStub,
        },
      },
    })

    await flushPromises()

    expect(replaceMock).toHaveBeenCalledWith({
      query: { panel: 'training-records' },
    })
    expect(wrapper.find('#student-tab-training-records').exists()).toBe(true)
    expect(wrapper.text()).toContain('训练记录')
    expect(wrapper.text()).toContain('web-1')
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
    expect(teacherStudentAnalysisSource).toContain('StudentAnalysisWorkspacePage')
    expect(teacherStudentAnalysisSource).toContain("from '@/features/teaching/student-analysis-workspace'")
    expect(teacherStudentAnalysisSource).not.toContain('useStudentAnalysisPage')
    expect(teacherStudentAnalysisSource).not.toContain('StudentAnalysisPage')
    expect(teacherStudentAnalysisSource).not.toContain(
      "import { ClassReportExportDialog } from '@/features/teaching/class-report-export'"
    )
    expect(studentAnalysisWorkspacePageSource).toContain(
      "import { ClassReportExportDialog } from '@/features/teaching/class-report-export'"
    )
    expect(studentAnalysisWorkspacePageSource).toContain('useStudentAnalysisPage')
    expect(studentAnalysisWorkspacePageSource).toContain('StudentAnalysisPage')
    expect(teacherStudentAnalysisSource).not.toContain('resolveClassManagementRouteName')
    expect(teacherStudentAnalysisSource).not.toContain('resolveClassStudentsRouteName')
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
    expect(studentAnalysisNavigationSource).not.toContain("from 'vue-router'")
    expect(studentAnalysisReviewQuerySyncSource).not.toContain("from 'vue-router'")
    expect(studentAnalysisPageModelSource).toContain(
      "import { useRouteNavigationTransport } from '@/shared/model/navigation/useRouteNavigationTransport'"
    )
    expect(studentAnalysisPageModelSource).toContain(
      "import { useRouteQueryTransport } from '@/shared/model/navigation/useRouteQueryTransport'"
    )
    expect(studentAnalysisPageModelSource).toContain(
      "import { useRouteQueryTabs } from '@/shared/model/navigation/useRouteQueryTabs'"
    )
    expect(studentAnalysisPageModelSource).toContain(
      'const { params, query, replaceQuery } = useRouteQueryTransport()'
    )
    expect(studentAnalysisPageModelSource).toContain(
      'const { push } = useRouteNavigationTransport()'
    )
    expect(studentAnalysisPageModelSource).not.toContain("from 'vue-router'")
    expect(studentAnalysisPageModelSource).not.toContain('useRoute(')
    expect(studentAnalysisPageModelSource).not.toContain('useRouter(')
    expect(studentAnalysisNavigationSource).toContain("from './studentAnalysisRoutes'")
    expect(studentAnalysisNavigationSource).toContain('studentAnalysisClassStudentsRoute')
    expect(studentAnalysisNavigationSource).toContain('studentAnalysisChallengeDetailRoute')
    expect(studentAnalysisNavigationSource).toContain('studentAnalysisReviewArchiveRoute')
    expect(studentAnalysisPageModelSource).not.toContain('useTeacherReviewWorkspace')
    expect(studentAnalysisPageModelSource).not.toContain('useTeacherSubmissionReviewFlows')
    expect(studentAnalysisPageModelSource).not.toContain('openClassManagement')
    expect(studentAnalysisPageModelSource).not.toContain('selectClass')
    expect(studentAnalysisPageModelSource).not.toContain('selectStudent')
    expect(studentAnalysisPageModelSource).toContain('activeTab: activeWorkspaceTab')
    expect(studentAnalysisPageModelSource).toContain('selectTab: selectWorkspaceTab')
  })

  it('路由页应提供可供 Transition 动画使用的单一元素根节点', () => {
    expect(teacherStudentAnalysisSource).not.toContain('<section class="teacher-route-root">')
    expect(teacherStudentAnalysisSource).toMatch(
      /<template>\s*<StudentAnalysisWorkspacePage route-root-class="teacher-route-root" \/>\s*<\/template>/s
    )
    expect(studentAnalysisWorkspacePageSource).toContain(':class="routeRootClass"')
    expect(studentAnalysisWorkspacePageSource).toMatch(
      /<template>\s*<section :class="routeRootClass">[\s\S]*<StudentAnalysisPage[\s\S]*<ClassReportExportDialog[\s\S]*<\/section>\s*<\/template>/s
    )
  })

  it('学员分析头部应只保留姓名标题，不重复显示英文 eyebrow 和用户名 chip', () => {
    expect(studentAnalysisPageSource).not.toContain('Student Analysis')
    expect(studentAnalysisPageSource).not.toContain('teacher-student-chip')
    expect(studentAnalysisPageSource).not.toContain('teacher-eyebrow-row')
    expect(studentAnalysisPageSource).not.toContain('useUrlSyncedTabs')
    expect(studentAnalysisPageSource).toContain('StudentAnalysisWorkspaceTabs')
    expect(studentAnalysisPageSource).toContain('StudentAnalysisWorkspaceContent')
    expect(studentAnalysisPageSource).not.toContain("from '@/shared/lib/keyboard/useTabKeyboardNavigation'")
    expect(studentAnalysisPageSource).not.toContain('StudentAnalysisOverviewHeroPanel')
    expect(studentAnalysisPageSource).not.toContain('StudentInsightPanel')
    expect(studentAnalysisPageSource).not.toContain('classes: ClassDirectoryItem[]')
    expect(studentAnalysisPageSource).toContain('selectedStudent: StudentDirectoryItem | null')
    expect(studentAnalysisPageSource).not.toContain('TeacherStudentItem')
    expect(studentAnalysisPageSource).not.toContain('selectedClassName: string')
    expect(studentAnalysisPageSource).not.toContain('selectedStudentId: string')
    expect(studentAnalysisPageSource).not.toContain('loadingClasses: boolean')
    expect(studentAnalysisPageSource).not.toContain('loadingStudents: boolean')
    expect(studentAnalysisPageSource).not.toContain('openClassManagement: []')
    expect(studentAnalysisPageSource).not.toContain('selectClass: [className: string]')
    expect(studentAnalysisPageSource).not.toContain('selectStudent: [studentId: string]')
    expect(studentAnalysisPageSource).not.toContain('<span>已做题目数</span>')
    expect(studentAnalysisWorkspaceTabsSource).toContain(
      "from '@/shared/lib/keyboard/useTabKeyboardNavigation'"
    )
    expect(studentAnalysisWorkspaceTabsSource).toContain('selectWorkspaceTab: [tab: StudentAnalysisWorkspaceTab]')
    expect(studentAnalysisWorkspaceTabsHelperSource).toContain("label: '学员画像'")
    expect(studentAnalysisWorkspaceTabsHelperSource).toContain("label: '证据链'")
    expect(studentAnalysisWorkspaceContentSource).toContain('StudentAnalysisOverviewHeroPanel')
    expect(studentAnalysisWorkspaceContentSource).toContain('StudentInsightPanel')
    expect(studentAnalysisWorkspaceContentSource).toContain(':loading="loadingDetails"')
    expect(studentAnalysisOverviewHeroPanelSource).toContain(
      "{{ selectedStudent?.name || selectedStudent?.username || '学员分析' }}"
    )
    expect(studentAnalysisOverviewHeroPanelSource).toContain('loading?: boolean')
    expect(studentAnalysisOverviewHeroPanelSource).toContain('v-if="loading"')
    expect(studentAnalysisOverviewHeroPanelSource).toContain('summary-card summary-card--loading progress-card metric-panel-card')
    expect(studentAnalysisOverviewHeroPanelSource).toContain("'summary-card-loading-value'")
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
    expect(studentInsightPanelSource).toContain('StudentInsightPrimarySections')
    expect(studentInsightPanelSource).toContain('StudentInsightReviewSections')
    expect(studentInsightPanelSource).toContain(
      "from '@/features/teaching/student-analysis-review'"
    )
    expect(studentInsightPanelSource).not.toContain('StudentInsightOverviewSection')
    expect(studentInsightPanelSource).not.toContain('StudentInsightRecommendationsSection')
    expect(studentInsightPanelSource).not.toContain('StudentInsightAttackSessionsSection')
    expect(studentInsightPanelSource).not.toContain('TeacherStudentReviewWorkspace')
    expect(studentInsightPrimarySectionsSource).toContain('StudentInsightOverviewSection')
    expect(studentInsightPrimarySectionsSource).toContain('StudentInsightRecommendationsSection')
    expect(studentInsightPrimarySectionsSource).toContain('StudentInsightTimelineSection')
    expect(studentInsightOverviewSectionSource).toContain('<SkillRadar :scores="radarScores" />')
    expect(studentInsightRecommendationsSectionSource).toContain(
      'class="insight-recommendation-list workspace-glass-region workspace-directory-list"'
    )
    expect(studentInsightRecommendationsSectionSource).toMatch(
      /<StudentInsightStateSurface[\s\S]*class="insight-recommendation-list workspace-glass-region workspace-directory-list"[\s\S]*<template #loading>[\s\S]*<template #empty>[\s\S]*<template #default>/s
    )
    expect(studentInsightPrimarySectionsSource).toContain(':loading="recommendationsLoading"')
    // merged: loading 和 loaded 共用同一个 StudentInsightPrimarySections 实例
    expect(studentInsightPanelSource).toContain(':recommendations-loading="loading"')
    expect(studentInsightPanelSource).toContain(':loading="loading"')
    // 不再包含独立的分支骨架
    expect(studentInsightPanelSource).not.toContain('insight-loading-shell')
    expect(studentInsightPanelSource).not.toContain('insight-skeleton-line')
    expect(studentInsightPanelSource).not.toContain('insight-skeleton-block')
    expect(studentInsightOverviewSectionSource).not.toContain('StudentInsightLoadingSurface')
    expect(studentInsightOverviewSectionSource).toContain('insight-overview-loading-radar')
    expect(studentInsightOverviewSectionSource).toContain('student-insight-skeleton-panel')
    expect(studentInsightOverviewSectionSource).toContain('class="insight-dimension-frame mt-4"')
    expect(studentInsightTimelineSectionSource).toContain('TrainingTimelineContent')
    expect(studentInsightTimelineSectionSource).toContain(':loading="loading"')
    expect(studentInsightTimelineSectionSource).not.toContain('StudentInsightLoadingSurface')
    expect(studentInsightTimelineSectionSource).not.toContain('<SectionCard')
    expect(studentInsightTimelineSectionSource).not.toContain('insight-timeline-loading-hero')
    expect(studentInsightTimelineSectionSource).not.toContain('insight-timeline-loading-list')
    expect(studentInsightTimelineSectionSource).not.toContain('insight-timeline-loading-row')
    expect(studentInsightTimelineSectionSource).not.toContain('workspace-glass-metric-surface')
    expect(studentInsightTimelineSectionSource).not.toContain('workspace-glass-region')
    expect(studentInsightRecommendationsSectionSource).toContain(
      '<style src="@/features/teaching/student-analysis-shared/ui/studentInsightSurface.css"></style>'
    )
    expect(studentInsightStateSurfaceSource).toContain('surface?: \'glass\' | \'plain\'')
    expect(studentInsightLoadingSurfaceSource).toContain('student-insight-glass-surface')
    expect(studentInsightRecommendationsSectionSource).toContain(
      '<div class="insight-recommendation-skeleton-head">'
    )
    expect(studentInsightRecommendationsSectionSource).toContain(
      'class="insight-recommendation-skeleton-row"'
    )
    expect(studentInsightRecommendationsSectionSource).toContain(
      'class="insight-recommendation-skeleton-pills"'
    )
    expect(studentInsightRecommendationsSectionSource).toMatch(
      /\.insight-recommendation-skeleton-row\s*\{[\s\S]*grid-template-columns:\s*minmax\(0,\s*1fr\)\s*auto\s*auto;[\s\S]*border-bottom:\s*1px solid var\(--workspace-directory-row-divider\);/s
    )
    expect(studentInsightRecommendationsSectionSource).toContain('student-insight-skeleton-line')
    expect(studentInsightStateSurfaceSource).toContain('student-insight-state-surface--loading')
    expect(studentInsightLoadingSurfaceSource).toContain('student-insight-glass-surface')
  })

  it('复盘区 section 应由 student-analysis-review feature 承接共享工作台组件', () => {
    expect(studentInsightReviewSectionsSource).toContain('StudentInsightWriteupsSection')
    expect(studentInsightReviewSectionsSource).toContain('StudentInsightManualReviewSection')
    expect(studentInsightReviewSectionsSource).toContain('StudentInsightAttackSessionsSection')
    expect(studentInsightAttackSessionsSectionSource).toContain(
      "import StudentReviewWorkspace from './StudentReviewWorkspace.vue'"
    )
    expect(studentInsightAttackSessionsSectionSource).not.toContain('TeacherStudentReviewWorkspace')
    expect(writeupsSectionSource).toContain('StudentInsightStateSurface')
    expect(writeupsSectionSource).toContain('writeup-loading-metrics')
    expect(writeupsSectionSource).toContain('writeup-loading-head')
    expect(writeupsSectionSource).toContain('writeup-loading-rows')
    expect(studentInsightReviewSectionsSource).toContain(':loading="writeupPaginationLoading"')
    expect(studentInsightAttackSessionsSectionSource).toContain('StudentInsightStateSurface')
    expect(studentInsightAttackSessionsSectionSource).toContain('evidence-loading-filters')
    expect(studentInsightAttackSessionsSectionSource).toContain('evidence-loading-summary')
    expect(studentInsightAttackSessionsSectionSource).toContain('evidence-loading-sessions')
    expect(writeupsSectionSource).not.toContain('writeup-glass')
    expect(studentInsightAttackSessionsSectionSource).not.toContain('evidence-glass')
    expect(studentAnalysisReviewQuerySyncSource).toContain("from '@/api/contracts'")
    expect(studentAnalysisReviewQuerySyncSource).not.toContain("from '@/api/teacher'")
    expect(studentAnalysisReviewQuerySyncSource).not.toContain("from '@/api/teaching'")
    expect(studentReviewWorkspaceSource).toContain("from '@/api/contracts'")
    expect(studentReviewWorkspaceSource).not.toContain("from '@/api/teacher'")
    expect(studentReviewWorkspaceSource).not.toContain("from '@/api/teaching'")
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
    replaceMock.mockClear()

    let selects = wrapper.findAll('select')
    await selects[1].setValue('awd')
    await flushPromises()
    selects = wrapper.findAll('select')
    await selects[2].setValue('failed')
    await flushPromises()

    expect(teachingApiMocks.getStudentAttackSessions).toHaveBeenNthCalledWith(1, 'stu-1', {
      with_events: true,
      limit: 20,
      offset: 0,
      mode: 'awd',
    })
    expect(teachingApiMocks.getStudentAttackSessions).toHaveBeenLastCalledWith('stu-1', {
      with_events: true,
      limit: 20,
      offset: 0,
      mode: 'awd',
      result: 'failed',
    })
    expect(teachingApiMocks.getStudentEvidence).toHaveBeenCalledTimes(1)
    expect(replaceMock).toHaveBeenNthCalledWith(1, {
      query: {
        panel: 'evidence',
        reviewMode: 'awd',
        reviewResult: undefined,
        reviewChallengeId: undefined,
      },
    })
    expect(replaceMock).toHaveBeenLastCalledWith({
      query: {
        panel: 'evidence',
        reviewMode: 'awd',
        reviewResult: 'failed',
        reviewChallengeId: undefined,
      },
    })
  })

  it('应从 panel query 恢复当前标签页，并在切换标签时保留已有复盘 query', async () => {
    routeMock.query = {
      panel: 'writeups',
      reviewMode: 'awd',
      reviewChallengeId: '11',
    }

    const wrapper = mount(TeacherStudentAnalysis, {
      global: {
        stubs: {
          SkillRadar: true,
          ClassReportExportDialog: reportDialogStub,
        },
      },
    })

    await flushPromises()

    expect(wrapper.text()).toContain('题解列表')
    expect(wrapper.text()).not.toContain('复盘工作台')

    replaceMock.mockClear()
    await wrapper.get('#student-tab-evidence').trigger('click')
    await flushPromises()

    expect(replaceMock).toHaveBeenCalledWith({
      query: {
        panel: 'evidence',
        reviewMode: 'awd',
        reviewChallengeId: '11',
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
    replaceMock.mockClear()

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
        panel: 'evidence',
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
    expect(wrapper.find('#student-tab-training-records').exists()).toBe(true)
    expect(studentAnalysisPageSource).toMatch(/class="[^"]*\bworkspace-shell\b[^"]*"/)
    expect(studentAnalysisPageSource).toMatch(
      /class="[^"]*\bstudent-analysis-shell\b[^"]*\bflex\b[^"]*\bmin-h-full\b[^"]*\bflex-1\b[^"]*\bflex-col\b[^"]*"/
    )
    expect(studentAnalysisPageSource).not.toContain('class="workspace-topbar"')
    expect(studentAnalysisPageSource).toContain('class="content-pane"')
    expect(studentAnalysisPageSource).toContain('StudentAnalysisWorkspaceTabs')
    expect(studentAnalysisWorkspaceTabsSource).toContain('class="workspace-tabbar top-tabs"')
    expect(studentAnalysisPageSource).toMatch(
      /<div\s+class="[^"]*\bworkspace-shell\b[^"]*\bworkspace-shell--plain\b[^"]*"\s*>[\s\S]*<StudentAnalysisWorkspaceTabs[\s\S]*<main class="content-pane">/s
    )
    expect(studentAnalysisPageSource).toMatch(
      /\.content-pane\s*\{[\s\S]*flex:\s*1 1 auto;[\s\S]*align-content:\s*start;/s
    )
    expect(studentAnalysisPageSource).toContain('workspace-shell--plain')
    expect(studentAnalysisPageSource).toContain('--workspace-shell-bg:')
    expect(studentAnalysisPageSource).toContain('--workspace-panel:')
    expect(studentAnalysisPageSource).toContain('--workspace-shadow-shell:')
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
