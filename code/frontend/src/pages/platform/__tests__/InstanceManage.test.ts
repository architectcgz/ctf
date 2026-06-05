import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createMemoryHistory, createRouter } from 'vue-router'

import instancePresentationSource from '@/entities/instance/model/presentation.ts?raw'
import PlatformInstanceManagement from '@/pages/platform/InstanceManageRoutePage.vue'
import adminInstanceManageSource from '@/pages/platform/InstanceManageRoutePage.vue?raw'
import platformInstanceManagementPageSource from '@/features/platform/instance-management/ui/PlatformInstanceManagementPage.vue?raw'
import instanceManageHeroPanelSource from '@/features/platform/instance-management/ui/InstanceManageHeroPanel.vue?raw'
import instanceManageWorkspacePanelSource from '@/features/platform/instance-management/ui/InstanceManageWorkspacePanel.vue?raw'
import platformInstanceManagementModelSource from '@/features/platform/instance-management/model/usePlatformInstanceManagementPage.ts?raw'
import managedInstanceDirectorySource from '@/features/managed-instance-directory/model/useManagedInstanceDirectory.ts?raw'
import managedInstanceDestroyWorkflowSource from '@/features/managed-instance-workflow/model/useManagedInstanceDestroyAction.ts?raw'

const instanceAccessApiMocks = vi.hoisted(() => ({
  getInstanceDirectoryByRole: vi.fn(),
  destroyManagedInstanceByRole: vi.fn(),
}))

const confirmMock = vi.hoisted(() => vi.fn())

vi.mock('@/api/instances', () => instanceAccessApiMocks)
vi.mock('@/shared/model/common/useDestructiveConfirm', () => ({
  confirmDestructiveAction: confirmMock,
}))

function createTestRouter() {
  return createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/platform/instances', component: PlatformInstanceManagement },
      {
        path: '/platform/overview',
        name: 'PlatformOverview',
        component: { template: '<div>platform overview</div>' },
      },
      {
        path: '/platform/students/:className/:studentId',
        name: 'PlatformStudentAnalysis',
        component: { template: '<div>platform student analysis</div>' },
      },
    ],
  })
}

async function mountPage() {
  const router = createTestRouter()
  await router.push('/platform/instances')
  await router.isReady()

  const wrapper = mount(PlatformInstanceManagement, {
    global: {
      plugins: [router],
    },
  })

  await flushPromises()
  return { wrapper, router }
}

describe('PlatformInstanceManagement', () => {
  beforeEach(() => {
    instanceAccessApiMocks.getInstanceDirectoryByRole.mockReset()
    instanceAccessApiMocks.destroyManagedInstanceByRole.mockReset()
    confirmMock.mockReset()

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
        {
          id: 'inst-2',
          student_id: 'stu-2',
          student_name: 'Bob',
          student_username: 'bob',
          student_no: 'S-1002',
          class_name: 'Class B',
          challenge_id: 'challenge-2',
          challenge_title: 'Pwn Stack 201',
          status: 'expired',
          access_url: '',
          expires_at: '2026-03-09T09:00:00Z',
          remaining_time: 0,
          extend_count: 0,
          max_extends: 3,
          created_at: '2026-03-09T08:30:00Z',
        },
      ],
      total: 2,
      page: 1,
      page_size: 15,
      summary: {
        total_count: 2,
        running_count: 1,
        expiring_soon_count: 0,
        warning_count: 1,
      },
    })
    instanceAccessApiMocks.destroyManagedInstanceByRole.mockResolvedValue(undefined)
    confirmMock.mockResolvedValue(true)
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('应保留当前后台实例页样式并复用 admin 实例目录接口 owner', async () => {
    expect(adminInstanceManageSource).toContain("from '@/features/platform/instance-management'")
    expect(adminInstanceManageSource).toContain('PlatformInstanceManagementPage')
    expect(adminInstanceManageSource).not.toContain('usePlatformInstanceManagementPage')
    expect(adminInstanceManageSource).not.toContain('InstanceManageHeroPanel')
    expect(adminInstanceManageSource).not.toContain('InstanceManageWorkspacePanel')
    expect(platformInstanceManagementModelSource).toContain('InstanceDirectoryItem')
    expect(platformInstanceManagementModelSource).toContain(
      "from '@/features/managed-instance-directory'"
    )
    expect(platformInstanceManagementModelSource).toContain(
      "from '@/features/managed-instance-workflow'"
    )
    expect(platformInstanceManagementModelSource).not.toContain("from '@/api/instances'")
    expect(platformInstanceManagementModelSource).toContain("from '@/entities/instance'")
    expect(platformInstanceManagementModelSource).not.toContain("from 'vue-router'")
    expect(platformInstanceManagementModelSource).toContain('const overviewRoute = platformOverviewRoute')
    expect(platformInstanceManagementModelSource).toContain('function buildStudentRoute')
    expect(platformInstanceManagementModelSource).not.toContain('confirmDestructiveAction')
    expect(platformInstanceManagementModelSource).not.toContain('TeacherInstanceItem')
    expect(platformInstanceManagementModelSource).not.toContain("from '@/api/admin'")
    expect(platformInstanceManagementModelSource).not.toContain('getPlatformInstances')
    expect(platformInstanceManagementModelSource).not.toContain('destroyPlatformInstance')
    expect(managedInstanceDirectorySource).toContain('getInstanceDirectoryByRole')
    expect(managedInstanceDirectorySource).toContain('useAbortController')
    expect(managedInstanceDirectorySource).toContain('scheduleSearch')
    expect(managedInstanceDestroyWorkflowSource).toContain('destroyManagedInstanceByRole')
    expect(adminInstanceManageSource).not.toContain("from '@/api/teacher'")
    expect(adminInstanceManageSource).not.toContain("from '@/shared/model/common/useDestructiveConfirm'")
    expect(adminInstanceManageSource).not.toContain("from '@/api/admin'")
    expect(adminInstanceManageSource).not.toContain("from '@/composables/useAdminDestructiveConfirm'")
    expect(platformInstanceManagementPageSource).toContain('usePlatformInstanceManagementPage')
    expect(platformInstanceManagementPageSource).toContain('InstanceManageWorkspacePanel')
    expect(platformInstanceManagementPageSource).toContain('InstanceManageHeroPanel')
    expect(platformInstanceManagementPageSource).not.toContain("from '@/api/instances'")
    expect(instanceManageHeroPanelSource).toContain("from '@/shared/ui/navigation/AppRouteLink.vue'")
    expect(instanceManageHeroPanelSource).toContain('返回概览')
    expect(instanceManageHeroPanelSource).toContain('刷新列表')
    expect(instanceManageWorkspacePanelSource).toContain("from '@/shared/ui/common/WorkspaceDataTable.vue'")
    expect(instanceManageWorkspacePanelSource).toContain(
      "from '@/shared/ui/common/WorkspaceDirectoryPagination.vue'"
    )
    expect(instanceManageWorkspacePanelSource).toContain(
      "from '@/shared/ui/common/WorkspaceDirectoryToolbar.vue'"
    )
    expect(instanceManageWorkspacePanelSource).toContain("from '@/entities/instance'")
    expect(instancePresentationSource).toContain('getInstanceStatusLabel')
    expect(instancePresentationSource).toContain('getInstanceStudentDisplayName')
    expect(instanceManageWorkspacePanelSource).toContain(
      "from '@/shared/ui/navigation/AppRouteLink.vue'"
    )
    expect(instanceManageWorkspacePanelSource).toContain('getInstanceStatusPillClass')

    const { wrapper } = await mountPage()

    expect(instanceAccessApiMocks.getInstanceDirectoryByRole).toHaveBeenCalledWith(
      'admin',
      {
        class_name: undefined,
        keyword: undefined,
        student_no: undefined,
        status: undefined,
        page: 1,
        page_size: 15,
      },
      expect.objectContaining({
        signal: expect.any(AbortSignal),
      })
    )
    expect(wrapper.text()).toContain('实例管理')
    expect(wrapper.text()).toContain('Web SQLi 101')
    expect(wrapper.text()).toContain('Alice')
    expect(wrapper.text()).toContain('Class A')
    expect(wrapper.text()).toContain('http://127.0.0.1:30001')
    expect(wrapper.text()).toContain('运行中')
    expect(wrapper.text()).toContain('已过期')
    expect(wrapper.text()).toContain('销毁')
    expect(wrapper.text()).toContain('共 2 个实例')
  })

  it('应支持按实例关键词筛选目录', async () => {
    vi.useFakeTimers()
    const { wrapper } = await mountPage()

    instanceAccessApiMocks.getInstanceDirectoryByRole.mockResolvedValue({
      list: [
        {
          id: 'inst-2',
          student_id: 'stu-2',
          student_name: 'Bob',
          student_username: 'bob',
          student_no: 'S-1002',
          class_name: 'Class B',
          challenge_id: 'challenge-2',
          challenge_title: 'Pwn Stack 201',
          status: 'expired',
          access_url: '',
          expires_at: '2026-03-09T09:00:00Z',
          remaining_time: 0,
          extend_count: 0,
          max_extends: 3,
          created_at: '2026-03-09T08:30:00Z',
        },
      ],
      total: 1,
      page: 1,
      page_size: 15,
      summary: {
        total_count: 1,
        running_count: 0,
        expiring_soon_count: 0,
        warning_count: 1,
      },
    })

    await wrapper.get('.workspace-directory-toolbar__search-input').setValue('Pwn')
    vi.advanceTimersByTime(250)
    await flushPromises()

    expect(instanceAccessApiMocks.getInstanceDirectoryByRole).toHaveBeenLastCalledWith(
      'admin',
      {
        class_name: undefined,
        keyword: 'Pwn',
        student_no: undefined,
        status: undefined,
        page: 1,
        page_size: 15,
      },
      expect.objectContaining({
        signal: expect.any(AbortSignal),
      })
    )
    expect(wrapper.text()).not.toContain('Web SQLi 101')
    expect(wrapper.text()).toContain('Pwn Stack 201')
    expect(wrapper.text()).toContain('共 1 个实例')
  })

  it('应支持点击所属用户进入学生分析页', async () => {
    const { wrapper, router } = await mountPage()

    await wrapper.get('.instance-user-link').trigger('click')
    await flushPromises()

    expect(router.currentRoute.value.name).toBe('PlatformStudentAnalysis')
    expect(router.currentRoute.value.params).toEqual({ className: 'Class A', studentId: 'stu-1' })
  })

  it('应支持销毁实例并更新列表', async () => {
    const { wrapper } = await mountPage()

    instanceAccessApiMocks.getInstanceDirectoryByRole.mockResolvedValue({
      list: [
        {
          id: 'inst-2',
          student_id: 'stu-2',
          student_name: 'Bob',
          student_username: 'bob',
          student_no: 'S-1002',
          class_name: 'Class B',
          challenge_id: 'challenge-2',
          challenge_title: 'Pwn Stack 201',
          status: 'expired',
          access_url: '',
          expires_at: '2026-03-09T09:00:00Z',
          remaining_time: 0,
          extend_count: 0,
          max_extends: 3,
          created_at: '2026-03-09T08:30:00Z',
        },
      ],
      total: 1,
      page: 1,
      page_size: 15,
      summary: {
        total_count: 1,
        running_count: 0,
        expiring_soon_count: 0,
        warning_count: 1,
      },
    })

    await wrapper
      .findAll('button')
      .find((node) => node.text().includes('销毁'))
      ?.trigger('click')
    await flushPromises()

    expect(confirmMock).toHaveBeenCalled()
    expect(instanceAccessApiMocks.destroyManagedInstanceByRole).toHaveBeenCalledWith(
      'admin',
      'inst-1'
    )
    expect(instanceAccessApiMocks.getInstanceDirectoryByRole).toHaveBeenLastCalledWith(
      'admin',
      {
        class_name: undefined,
        keyword: undefined,
        student_no: undefined,
        status: undefined,
        page: 1,
        page_size: 15,
      },
      expect.objectContaining({
        signal: expect.any(AbortSignal),
      })
    )
    expect(wrapper.text()).not.toContain('Web SQLi 101')
    expect(wrapper.text()).toContain('Pwn Stack 201')
  })

  it('应支持返回概览页', async () => {
    const { wrapper, router } = await mountPage()
    const overviewLink = wrapper.findAll('a').find((node) => node.text().includes('返回概览'))

    expect(overviewLink).toBeTruthy()

    await overviewLink!.trigger('click')
    await flushPromises()

    expect(router.currentRoute.value.name).toBe('PlatformOverview')
  })
})
