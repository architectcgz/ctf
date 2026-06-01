import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { flushPromises, mount } from '@vue/test-utils'
import { createMemoryHistory, createRouter } from 'vue-router'

import instancePresentationSource from '@/entities/instance/model/presentation.ts?raw'
import InstanceManagement from '@/pages/teacher/InstanceManagementRoutePage.vue'
import instanceManagementViewSource from '@/pages/teacher/InstanceManagementRoutePage.vue?raw'
import appRouteLinkSource from '@/shared/ui/navigation/AppRouteLink.vue?raw'
import instanceManagementSourceBase from '@/features/teacher/instances/ui/TeacherInstanceManagementPage.vue?raw'
import teacherInstanceManagementPageModelSource from '@/features/teacher/instances/model/useInstanceManagementPage.ts?raw'
import teacherInstanceDirectorySectionSource from '@/features/teacher/instances/ui/TeacherInstanceDirectorySection.vue?raw'
import teacherInstanceHeroPanelSource from '@/features/teacher/instances/ui/TeacherInstanceHeroPanel.vue?raw'
import teacherInstancesHookSource from '@/features/teacher/instances/model/useInstances.ts?raw'
import managedInstanceDirectorySource from '@/features/managed-instance-directory/model/useManagedInstanceDirectory.ts?raw'
import managedInstanceDestroyWorkflowSource from '@/features/managed-instance-workflow/model/useManagedInstanceDestroyAction.ts?raw'
import { useAuthStore } from '@/stores/auth'

const instanceManagementSource = [
  instanceManagementSourceBase,
  teacherInstanceHeroPanelSource,
  teacherInstanceDirectorySectionSource,
].join('\n')

const ElTable = { template: '<div><slot /></div>' }
const ElTableColumn = { template: '<div><slot /></div>' }
const ElButton = { template: '<button><slot /></button>' }

const teacherApiMocks = vi.hoisted(() => ({
  getClasses: vi.fn(),
}))

const instanceAccessApiMocks = vi.hoisted(() => ({
  getInstanceDirectoryByRole: vi.fn(),
  destroyManagedInstanceByRole: vi.fn(),
}))

const confirmMock = vi.hoisted(() => vi.fn())

vi.mock('@/api/teacher', () => teacherApiMocks)
vi.mock('@/api/instances', () => instanceAccessApiMocks)
vi.mock('@/shared/model/common/useDestructiveConfirm', () => ({
  confirmDestructiveAction: confirmMock,
}))

let pinia: ReturnType<typeof createPinia>

function createTestRouter() {
  return createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/academy/instances', component: InstanceManagement },
      {
        path: '/academy/dashboard',
        name: 'TeacherDashboard',
        component: { template: '<div>teacher dashboard</div>' },
      },
      {
        path: '/platform/overview',
        name: 'PlatformOverview',
        component: { template: '<div>platform overview</div>' },
      },
    ],
  })
}

async function mountPage() {
  const router = createTestRouter()
  await router.push('/academy/instances')
  await router.isReady()

  const wrapper = mount(InstanceManagement, {
    global: {
      plugins: [pinia, router],
      components: {
        ElTable,
        ElTableColumn,
        ElButton,
      },
    },
  })
  await flushPromises()
  return { wrapper, router }
}

describe('InstanceManagement', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    pinia = createPinia()
    setActivePinia(pinia)
    localStorage.clear()
    Object.values(teacherApiMocks).forEach((mock) => mock.mockReset())
    Object.values(instanceAccessApiMocks).forEach((mock) => mock.mockReset())

    teacherApiMocks.getClasses.mockResolvedValue([{ name: 'Class A', student_count: 1 }])
    instanceAccessApiMocks.getInstanceDirectoryByRole.mockResolvedValue({
      list: [
        {
          id: 'inst-1',
          student_id: 'stu-1',
          student_name: 'Alice',
          student_username: 'alice',
          student_no: 'S-1001',
          class_name: 'Class A',
          challenge_id: 'challenge-1',
          challenge_title: 'Web SQLi 101',
          status: 'running',
          access_url: 'http://127.0.0.1:30001',
          expires_at: '2026-03-09T10:30:00Z',
          remaining_time: 1200,
          extend_count: 1,
          max_extends: 3,
          created_at: '2026-03-09T09:30:00Z',
        },
      ],
      total: 1,
      page: 1,
      page_size: 20,
      summary: {
        total_count: 1,
        running_count: 1,
        expiring_soon_count: 0,
        warning_count: 0,
      },
    })
    instanceAccessApiMocks.destroyManagedInstanceByRole.mockResolvedValue(undefined)
    confirmMock.mockReset()
    confirmMock.mockResolvedValue(true)

    const authStore = useAuthStore()
    authStore.setAuth({
      id: 'teacher-1',
      username: 'teacher',
      role: 'teacher',
      class_name: 'Class A',
    })
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('应该按教师所属班级加载实例', async () => {
    const { wrapper } = await mountPage()

    expect(instanceAccessApiMocks.getInstanceDirectoryByRole).toHaveBeenCalledWith(
      'teacher',
      {
        class_name: 'Class A',
        keyword: undefined,
        student_no: undefined,
        page: 1,
        page_size: 20,
      },
      expect.objectContaining({
        signal: expect.any(AbortSignal),
      })
    )
    expect(wrapper.findAll('.progress-card.metric-panel-card')).toHaveLength(3)
    expect(wrapper.find('.workspace-data-table').exists()).toBe(true)
    expect(wrapper.find('.teacher-directory-shell.workspace-directory-list').exists()).toBe(true)
    expect(wrapper.find('.teacher-instance-list.workspace-directory-list').exists()).toBe(false)
    expect(wrapper.text()).toContain('Web SQLi 101')
    expect(wrapper.text()).toContain('@alice')
    expect(wrapper.findAll('.teacher-instance-primary-text')[0].attributes('title')).toBe('Alice')
    expect(wrapper.findAll('.teacher-instance-primary-text')[1].attributes('title')).toBe(
      'Web SQLi 101'
    )
    const mutedCells = wrapper.findAll('.teacher-instance-muted-text')
    expect(mutedCells[2].text()).toBe('1 / 3')
    expect(mutedCells[3].text()).toBe('00:20:00')
    expect(wrapper.find('.teacher-instance-url-text').text()).toContain('http://127.0.0.1:30001')
    expect(wrapper.text()).not.toContain('重置筛选')
    expect(wrapper.findAll('button').some((node) => node.text().includes('查询实例'))).toBe(false)
    expect(wrapper.text()).not.toContain('实例筛选')
    expect(wrapper.text()).not.toContain('支持按班级、用户名或学号关键字筛选，也可用学号精确筛选。')
  })

  it('路由页应仅负责组合，不直接依赖页面级流程细节', () => {
    expect(instanceManagementViewSource).toContain("from '@/features/teacher/instances'")
    expect(instanceManagementViewSource).toContain('TeacherInstanceManagementPage')
    expect(instanceManagementViewSource).toContain('useTeacherInstanceManagementPage')
    expect(instanceManagementViewSource).not.toContain('useInstanceManagementPage')
    expect(instanceManagementViewSource).not.toContain('confirmDestructiveAction')
    expect(instanceManagementViewSource).not.toContain('resolveTeachingDashboardRouteName')
    expect(teacherInstancesHookSource).toContain("from '@/features/managed-instance-directory'")
    expect(teacherInstancesHookSource).toContain("from '@/features/managed-instance-workflow'")
    expect(teacherInstancesHookSource).toContain('export function useTeacherInstanceDirectoryState()')
    expect(teacherInstancesHookSource).not.toContain('export function useInstances()')
    expect(teacherInstancesHookSource).not.toContain("from '@/api/instances'")
    expect(teacherInstancesHookSource).not.toContain('getTeacherInstances')
    expect(teacherInstancesHookSource).not.toContain('destroyTeacherInstance')
    expect(teacherInstancesHookSource).toContain("reportFrontendError('加载教师实例管理页失败:', err)")
    expect(teacherInstancesHookSource).toContain("reportFrontendError('教师销毁实例失败:', err)")
    expect(teacherInstancesHookSource).not.toContain("console.error('加载教师实例")
    expect(teacherInstanceManagementPageModelSource).toContain(
      'export function useTeacherInstanceManagementPage()'
    )
    expect(teacherInstanceManagementPageModelSource).toContain('useTeacherInstanceDirectoryState()')
    expect(teacherInstanceManagementPageModelSource).not.toContain('useInstances()')
    expect(teacherInstanceManagementPageModelSource).not.toContain('confirmDestructiveAction')
    expect(teacherInstanceManagementPageModelSource).not.toContain("from 'vue-router'")
    expect(teacherInstanceManagementPageModelSource).toContain(
      'const dashboardRoute = teacherInstanceDashboardRoute(authStore.user?.role)'
    )
    expect(managedInstanceDestroyWorkflowSource).toContain('destroyManagedInstanceByRole')
    expect(managedInstanceDestroyWorkflowSource).toContain('confirmDestructiveAction')
    expect(managedInstanceDirectorySource).toContain('getInstanceDirectoryByRole')
    expect(managedInstanceDirectorySource).toContain('useAbortController')
    expect(managedInstanceDirectorySource).toContain('scheduleSearch')
    expect(teacherInstanceHeroPanelSource).toContain("from '@/shared/ui/navigation/AppRouteLink.vue'")
    expect(teacherInstanceHeroPanelSource).toContain('<AppRouteLink')
    expect(appRouteLinkSource).toContain("from 'vue-router'")
  })

  it('应该支持输入后自动筛选并销毁实例', async () => {
    const { wrapper } = await mountPage()

    const inputs = wrapper.findAll('input')
    await inputs[0].setValue('ali')
    await inputs[1].setValue('S-1001')
    expect(instanceAccessApiMocks.getInstanceDirectoryByRole).toHaveBeenCalledTimes(1)
    vi.advanceTimersByTime(250)
    await flushPromises()

    expect(instanceAccessApiMocks.getInstanceDirectoryByRole).toHaveBeenLastCalledWith(
      'teacher',
      {
        class_name: 'Class A',
        keyword: 'ali',
        student_no: 'S-1001',
        page: 1,
        page_size: 20,
      },
      expect.objectContaining({
        signal: expect.any(AbortSignal),
      })
    )

    instanceAccessApiMocks.getInstanceDirectoryByRole.mockResolvedValue({
      list: [],
      total: 0,
      page: 1,
      page_size: 20,
      summary: {
        total_count: 0,
        running_count: 0,
        expiring_soon_count: 0,
        warning_count: 0,
      },
    })

    await wrapper.find('[data-instance-id="inst-1"]').trigger('click')
    await flushPromises()

    expect(confirmMock).toHaveBeenCalled()
    expect(instanceAccessApiMocks.destroyManagedInstanceByRole).toHaveBeenCalledWith(
      'teacher',
      'inst-1'
    )
    expect(wrapper.text()).not.toContain('Web SQLi 101')
  })

  it('取消危险确认后不应继续销毁实例', async () => {
    confirmMock.mockResolvedValue(false)

    const { wrapper } = await mountPage()

    await wrapper.find('[data-instance-id="inst-1"]').trigger('click')
    await flushPromises()

    expect(confirmMock).toHaveBeenCalled()
    expect(instanceAccessApiMocks.destroyManagedInstanceByRole).not.toHaveBeenCalled()
    expect(wrapper.text()).toContain('Web SQLi 101')
  })

  it('管理员从实例管理返回概览时应回到后台概览页', async () => {
    const authStore = useAuthStore()
    authStore.setAuth({
      id: 'admin-1',
      username: 'admin',
      role: 'admin',
      class_name: 'Class A',
    })

    const { wrapper, router } = await mountPage()
    const dashboardLink = wrapper.findAll('a').find((node) => node.text().includes('返回教学概览'))

    expect(dashboardLink).toBeTruthy()

    await dashboardLink!.trigger('click')
    await flushPromises()

    expect(router.currentRoute.value.name).toBe('PlatformOverview')
  })

  it('应该支持实例目录分页切换', async () => {
    instanceAccessApiMocks.getInstanceDirectoryByRole.mockImplementation(
      async (
        role: string | null | undefined,
        params?: { page?: number; page_size?: number }
      ) => {
        expect(role).toBe('teacher')
        const requestedPage = params?.page ?? 1
        if (requestedPage === 2) {
          return {
            list: [
              {
                id: 'inst-21',
                student_id: 'stu-21',
                student_name: 'Student 21',
                student_username: 'student-21',
                student_no: 'S-0021',
                class_name: 'Class A',
                challenge_id: 'challenge-21',
                challenge_title: 'Challenge 21',
                status: 'running',
                access_url: 'http://127.0.0.1:30021',
                expires_at: '2026-03-09T10:30:00Z',
                remaining_time: 1200,
                extend_count: 1,
                max_extends: 3,
                created_at: '2026-03-09T09:30:00Z',
              },
            ],
            total: 21,
            page: 2,
            page_size: 20,
            summary: {
              total_count: 21,
              running_count: 21,
              expiring_soon_count: 0,
              warning_count: 0,
            },
          }
        }

        return {
          list: [
            ...Array.from({ length: 20 }, (_, index) => ({
              id: `inst-${index + 1}`,
              student_id: `stu-${index + 1}`,
              student_name: `Student ${index + 1}`,
              student_username: `student-${index + 1}`,
              student_no: `S-${String(index + 1).padStart(4, '0')}`,
              class_name: 'Class A',
              challenge_id: `challenge-${index + 1}`,
              challenge_title: `Challenge ${index + 1}`,
              status: 'running',
              access_url: `http://127.0.0.1:30${String(index + 1).padStart(3, '0')}`,
              expires_at: '2026-03-09T10:30:00Z',
              remaining_time: 1200,
              extend_count: 1,
              max_extends: 3,
              created_at: '2026-03-09T09:30:00Z',
            })),
          ],
          total: 21,
          page: 1,
          page_size: 20,
          summary: {
            total_count: 21,
            running_count: 21,
            expiring_soon_count: 0,
            warning_count: 0,
          },
        }
      }
    )

    const { wrapper } = await mountPage()

    expect(wrapper.findAll('.teacher-directory-row')).toHaveLength(20)
    expect(wrapper.find('.teacher-directory-pagination').text()).toContain('共 21 条实例')
    expect(wrapper.find('.teacher-directory-pagination').text()).toContain('1 / 2')
    expect(wrapper.text()).toContain('Challenge 20')
    expect(wrapper.text()).not.toContain('Challenge 21')

    const paginationButtons = wrapper.findAll('.page-pagination-controls__button')
    await paginationButtons[1].trigger('click')
    await flushPromises()

    expect(instanceAccessApiMocks.getInstanceDirectoryByRole).toHaveBeenLastCalledWith(
      'teacher',
      {
        class_name: 'Class A',
        keyword: undefined,
        student_no: undefined,
        page: 2,
        page_size: 20,
      },
      expect.objectContaining({
        signal: expect.any(AbortSignal),
      })
    )
    expect(wrapper.findAll('.teacher-directory-row')).toHaveLength(1)
    expect(wrapper.find('.teacher-directory-pagination').text()).toContain('2 / 2')
    expect(wrapper.text()).toContain('Challenge 21')
    expect(wrapper.text()).not.toContain('Challenge 20')
  })

  it('应该为教师实例列表长文本保留省略样式与完整提示', () => {
    expect(instanceManagementSource).toContain(
      'class="workspace-directory-section teacher-directory-section"'
    )
    expect(instanceManagementSource).toContain('class="list-heading"')
    expect(instanceManagementSource).toContain(
      'class="teacher-summary metric-panel-default-surface"'
    )
    expect(instanceManagementSource).toContain(
      'class="teacher-summary-grid progress-strip metric-panel-grid metric-panel-default-surface"'
    )
    expect(instanceManagementSource).toContain('class="progress-card metric-panel-card"')
    expect(instanceManagementSource).not.toContain(
      'class="teacher-summary-item progress-card metric-panel-card"'
    )
    expect(instanceManagementSource).toContain('class="progress-card-label metric-panel-label"')
    expect(instanceManagementSource).not.toContain(
      'class="teacher-summary-label progress-card-label metric-panel-label"'
    )
    expect(instanceManagementSource).toContain('class="progress-card-value metric-panel-value"')
    expect(instanceManagementSource).not.toContain(
      'class="teacher-summary-value progress-card-value metric-panel-value"'
    )
    expect(instanceManagementSource).toContain('class="progress-card-hint metric-panel-helper"')
    expect(instanceManagementSource).not.toContain(
      'class="teacher-summary-helper progress-card-hint metric-panel-helper"'
    )
    expect(instanceManagementSource).toContain('<span>当前可见</span>')
    expect(instanceManagementSource).toContain('<Eye class="h-4 w-4" />')
    expect(instanceManagementSource).toContain('<span>运行中</span>')
    expect(instanceManagementSource).toContain('<Activity class="h-4 w-4" />')
    expect(instanceManagementSource).toContain('<span>即将到期</span>')
    expect(instanceManagementSource).toContain('<Clock3 class="h-4 w-4" />')
    expect(instanceManagementSource).not.toContain('teacher-controls-title')
    expect(instanceManagementSource).not.toContain('teacher-controls-copy')
    expect(instanceManagementSource).toContain(
      "from '@/shared/ui/common/WorkspaceDataTable.vue'"
    )
    expect(teacherInstanceDirectorySectionSource).toContain("from '@/entities/instance'")
    expect(instancePresentationSource).toContain('getInstanceStudentDisplayName')
    expect(instancePresentationSource).toContain('getInstanceStatusPillClass')
    expect(instanceManagementSource).toContain(
      "from '@/shared/ui/common/WorkspaceDirectoryPagination.vue'"
    )
    expect(instanceManagementSource).toContain('<WorkspaceDataTable')
    expect(instanceManagementSource).toContain('<WorkspaceDirectoryPagination')
    expect(instanceManagementSource).not.toContain("from '@/shared/ui/common/PagePaginationControls.vue'")
    expect(instanceManagementSource).toContain('用户关键字')
    expect(instanceManagementSource).toContain('按用户名或学号搜索')
    expect(instanceManagementSource).toContain("label: '学生'")
    expect(instanceManagementSource).toContain("label: '题目'")
    expect(instanceManagementSource).not.toContain('<span>学生 / 题目</span>')
    expect(instanceManagementSource).toContain("label: '创建时间'")
    expect(instanceManagementSource).toContain("label: '到期时间'")
    expect(instanceManagementSource).toContain("label: '延期'")
    expect(instanceManagementSource).toContain("label: '剩余时间'")
    expect(instanceManagementSource).toContain("label: '访问地址'")
    expect(instanceManagementSource).not.toContain('class="teacher-directory-head"')
    expect(instanceManagementSource).not.toContain('<div class="teacher-directory-head">')
    expect(instanceManagementSource).toContain('class="teacher-instance-primary-text"')
    expect(instanceManagementSource).toContain('class="teacher-instance-secondary-text"')
    expect(instanceManagementSource).toContain('class="teacher-instance-muted-text"')
    expect(instanceManagementSource).toContain('class="teacher-instance-url-text"')
    expect(instanceManagementSource).toContain('class="teacher-directory-pagination"')
    expect(instanceManagementSource).not.toContain('重置筛选')
    expect(instanceManagementSource).not.toContain('查询实例')
    expect(instanceManagementSource).not.toContain('创建于 {{ formatDateTime(item.created_at) }}')
    expect(instanceManagementSource).not.toContain('到期 {{ formatDateTime(item.expires_at) }}')
    expect(instanceManagementSource).not.toContain(
      '延期 {{ item.extend_count }} / {{ item.max_extends }}'
    )
    expect(instanceManagementSource).not.toContain(
      '剩余 {{ formatRemainingTime(item.remaining_time) }}'
    )
    expect(instanceManagementSource).not.toContain(
      'class="teacher-directory-chip teacher-directory-chip-muted"'
    )
    expect(instanceManagementSource).not.toContain('class="teacher-directory-row-metrics"')
    expect(instanceManagementSource).toContain('instance-status-pill--running')
    expect(instanceManagementSource).toContain('class="instance-status-pill"')
    expect(instanceManagementSource).toContain('class="ui-btn ui-btn--danger ui-btn--xs teacher-instance-danger-action"')
    expect(instanceManagementSource).not.toContain('teacher-row-btn')
    expect(instanceManagementSource).not.toContain('销毁实例')
    expect(instanceManagementSource).not.toContain('border-[var(--color-success)]')
    expect(instanceManagementSource).not.toContain('bg-[var(--color-success)]')
    expect(instanceManagementSource).not.toContain('text-[var(--color-success)]')
    expect(instanceManagementSource).not.toContain('border-[var(--color-primary)]')
    expect(instanceManagementSource).not.toContain('bg-[var(--color-primary)]')
    expect(instanceManagementSource).not.toContain('text-[var(--color-primary)]')
    expect(instanceManagementSource).toContain(
      'class="teacher-directory-shell workspace-directory-list"'
    )
    expect(instanceManagementSource).not.toContain(
      'class="teacher-instance-list workspace-directory-list"'
    )
    expect(instanceManagementSource).toMatch(
      /class="teacher-instance-primary-text"[\s\S]*:title="getInstanceStudentDisplayName\(\s*row as InstanceDirectoryItem\s*\)"/s
    )
    expect(instanceManagementSource).toMatch(
      /class="teacher-instance-primary-text"[\s\S]*:title="\(row as InstanceDirectoryItem\)\.challenge_title"/s
    )
    expect(instanceManagementSource).toContain('InstanceDirectoryItem')
    expect(instanceManagementSource).not.toContain('TeacherInstanceItem')
    expect(instanceManagementSource).toMatch(/class="teacher-instance-secondary-text"[\s\S]*:title="/s)
    expect(instanceManagementSource).toContain('getInstanceStudentIdentityLabel')
    expect(instanceManagementSource).toContain('getInstanceStudentSecondaryLabel')
    expect(instanceManagementSource).toContain('getInstanceStatusPillClass')
    expect(instanceManagementSource).toContain('formatInstanceRemainingTime')
    expect(instanceManagementSource).toMatch(
      /\.teacher-instance-primary-text,\s*[\s\S]*\.teacher-instance-url-text\s*\{[^}]*overflow:\s*hidden;[^}]*text-overflow:\s*ellipsis;[^}]*white-space:\s*nowrap;/s
    )
    expect(instanceManagementSource).toMatch(
      /\.teacher-instance-url-text\s*\{[^}]*overflow-wrap:\s*anywhere;[^}]*word-break:\s*break-word;/s
    )
  })
})
