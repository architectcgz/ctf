import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import PlatformInstanceManagement from '../InstanceManage.vue'
import adminInstanceManageSource from '../InstanceManage.vue?raw'
import instanceManageHeroPanelSource from '@/components/platform/instance/InstanceManageHeroPanel.vue?raw'
import instanceManageWorkspacePanelSource from '@/components/platform/instance/InstanceManageWorkspacePanel.vue?raw'
import platformInstanceManagementModelSource from '@/features/platform-instance-management/model/usePlatformInstanceManagementPage.ts?raw'

const pushMock = vi.fn()

const instanceAccessApiMocks = vi.hoisted(() => ({
  getInstanceDirectoryByRole: vi.fn(),
  destroyManagedInstanceByRole: vi.fn(),
}))

const confirmMock = vi.hoisted(() => vi.fn())

vi.mock('vue-router', async () => {
  const actual = await vi.importActual<typeof import('vue-router')>('vue-router')
  return {
    ...actual,
    useRouter: () => ({ push: pushMock }),
  }
})

vi.mock('@/api/instances', () => instanceAccessApiMocks)
vi.mock('@/composables/useDestructiveConfirm', () => ({
  confirmDestructiveAction: confirmMock,
}))

describe('PlatformInstanceManagement', () => {
  beforeEach(() => {
    pushMock.mockReset()
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
    expect(adminInstanceManageSource).toContain(
      "usePlatformInstanceManagementPage } from '@/features/platform-instance-management'"
    )
    expect(platformInstanceManagementModelSource).toContain('InstanceDirectoryItem')
    expect(platformInstanceManagementModelSource).toContain("from '@/api/instances'")
    expect(platformInstanceManagementModelSource).not.toContain('TeacherInstanceItem')
    expect(platformInstanceManagementModelSource).not.toContain("from '@/api/admin'")
    expect(platformInstanceManagementModelSource).not.toContain('getPlatformInstances')
    expect(platformInstanceManagementModelSource).not.toContain('destroyPlatformInstance')
    expect(adminInstanceManageSource).not.toContain("from '@/api/teacher'")
    expect(adminInstanceManageSource).not.toContain("from '@/composables/useDestructiveConfirm'")
    expect(adminInstanceManageSource).not.toContain("from '@/api/admin'")
    expect(adminInstanceManageSource).not.toContain("from '@/composables/useAdminDestructiveConfirm'")
    expect(adminInstanceManageSource).toContain(
      "import InstanceManageWorkspacePanel from '@/components/platform/instance/InstanceManageWorkspacePanel.vue'"
    )
    expect(adminInstanceManageSource).toContain(
      "import InstanceManageHeroPanel from '@/components/platform/instance/InstanceManageHeroPanel.vue'"
    )
    expect(adminInstanceManageSource).toContain('<InstanceManageHeroPanel')
    expect(adminInstanceManageSource).toContain('<InstanceManageWorkspacePanel')
    expect(adminInstanceManageSource).toContain(
      'class="workspace-shell journal-shell journal-shell-admin journal-hero admin-instance-manage-shell"'
    )
    expect(instanceManageHeroPanelSource).toContain('返回概览')
    expect(instanceManageHeroPanelSource).toContain('刷新列表')
    expect(instanceManageHeroPanelSource).toContain('class="header-btn header-btn--primary"')
    expect(instanceManageHeroPanelSource).toContain(
      'class="admin-summary-grid admin-instance-manage-shell__summary progress-strip metric-panel-grid metric-panel-default-surface metric-panel-workspace-surface"'
    )
    expect(instanceManageWorkspacePanelSource).toContain("from '@/components/common/WorkspaceDataTable.vue'")
    expect(instanceManageWorkspacePanelSource).toContain(
      "from '@/components/common/WorkspaceDirectoryPagination.vue'"
    )
    expect(instanceManageWorkspacePanelSource).toContain(
      "from '@/components/common/WorkspaceDirectoryToolbar.vue'"
    )
    expect(instanceManageWorkspacePanelSource).toContain('<WorkspaceDataTable')
    expect(instanceManageWorkspacePanelSource).toContain('<WorkspaceDirectoryToolbar')
    expect(instanceManageWorkspacePanelSource).toContain('<WorkspaceDirectoryPagination')
    expect(instanceManageWorkspacePanelSource).toContain('search-placeholder="检索实例、题目、用户或访问地址..."')
    expect(instanceManageWorkspacePanelSource).toContain('filter-panel-title="实例筛选"')
    expect(instanceManageWorkspacePanelSource).toContain("label: '班级'")
    expect(instanceManageWorkspacePanelSource).toContain('class="instance-user-link"')
    expect(instanceManageWorkspacePanelSource).toContain('class="instance-status-pill"')
    expect(instanceManageWorkspacePanelSource).toContain('class="ui-btn ui-btn--danger ui-btn--xs"')
    expect(adminInstanceManageSource).not.toContain('bg-green-100 text-green-700')
    expect(adminInstanceManageSource).not.toContain('bg-slate-100 text-slate-600')

    const wrapper = mount(PlatformInstanceManagement)
    await flushPromises()

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
    const wrapper = mount(PlatformInstanceManagement)
    await flushPromises()

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
    const wrapper = mount(PlatformInstanceManagement)
    await flushPromises()

    await wrapper.get('.instance-user-link').trigger('click')

    expect(pushMock).toHaveBeenCalledWith({
      name: 'PlatformStudentAnalysis',
      params: { className: 'Class A', studentId: 'stu-1' },
    })
  })

  it('应支持销毁实例并更新列表', async () => {
    const wrapper = mount(PlatformInstanceManagement)
    await flushPromises()

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
    const wrapper = mount(PlatformInstanceManagement)
    await flushPromises()

    await wrapper
      .findAll('button')
      .find((node) => node.text().includes('返回概览'))
      ?.trigger('click')

    expect(pushMock).toHaveBeenCalledWith({ name: 'PlatformOverview' })
  })
})
