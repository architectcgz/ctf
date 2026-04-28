import { afterEach, describe, expect, it } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import PlatformContestTable from '../contest/PlatformContestTable.vue'
import adminContestTableSource from '../contest/PlatformContestTable.vue?raw'
import workspaceDataTableSource from '@/components/common/WorkspaceDataTable.vue?raw'
import type { ContestDetailData } from '@/api/contracts'

function buildContest(overrides: Partial<ContestDetailData> = {}): ContestDetailData {
  return {
    id: 'awd-running',
    title: '2026 AWD 联赛',
    description: '运行中赛事',
    mode: 'awd',
    status: 'running',
    starts_at: '2026-04-15T09:00:00.000Z',
    ends_at: '2026-04-15T18:00:00.000Z',
    ...overrides,
  }
}

describe('PlatformContestTable', () => {
  afterEach(() => {
    document.body.innerHTML = ''
  })

  it('更多菜单应收敛到共享 action menu primitive，而不是继续维护赛事私有浮层', () => {
    expect(adminContestTableSource).toContain("from '@/components/common/menus/CActionMenu.vue'")
    expect(adminContestTableSource).not.toContain('<Teleport to="body">')
    expect(adminContestTableSource).not.toContain('class="contest-row-menu__title">Management</div>')
  })

  it('赛事页不应继续保留页面私有菜单 token 和 light/dark 分支', () => {
    expect(adminContestTableSource).not.toContain('--contest-action-surface')
    expect(adminContestTableSource).not.toContain('--contest-action-line')
    expect(adminContestTableSource).not.toContain(":global([data-theme='dark']) .contest-row-menu")
    expect(adminContestTableSource).not.toContain(":global([data-theme='dark']) .contest-row-menu-button")
  })

  it('分页壳层应通过语义类承接弱文本色，而不是继续在模板里内联主题 utility', () => {
    expect(adminContestTableSource).toContain('contest-pagination-tone')
    expect(adminContestTableSource).not.toContain('text-[var(--color-text-muted)]')
  })

  it('竞赛目录字号应与平台审计列表使用同一组目录 token', () => {
    expect(adminContestTableSource).toContain("from '@/components/common/WorkspaceDataTable.vue'")
    expect(adminContestTableSource).toContain('<WorkspaceDataTable')
    expect(adminContestTableSource).toContain('class="contest-directory workspace-directory-list"')
    expect(workspaceDataTableSource).toContain('font-size: var(--font-size-11);')
    expect(adminContestTableSource).toContain('font-size: var(--font-size-14);')
    expect(adminContestTableSource).toContain('font-size: var(--font-size-13);')
    expect(adminContestTableSource).toContain('--ui-badge-size: var(--font-size-11);')
    expect(adminContestTableSource).toContain('border-left: 1px solid var(--workspace-table-line);')
    expect(adminContestTableSource).not.toContain('font-size: var(--font-size-1-00);')
    expect(adminContestTableSource).not.toContain('font-size: var(--font-size-0-875);')
    expect(adminContestTableSource).not.toContain('font-size: var(--font-size-0-90);')
    expect(workspaceDataTableSource).not.toContain('font-size: 0.6875rem;')
    expect(adminContestTableSource).not.toContain('--ui-badge-size: var(--font-size-0-78);')
  })

  it('行内操作应直接提供编辑入口，更多菜单不再承载导出结果', async () => {
    const wrapper = mount(PlatformContestTable, {
      attachTo: document.body,
      props: {
        contests: [buildContest()],
        page: 1,
        pageSize: 20,
        total: 1,
      },
    })

    expect(wrapper.get('#contest-open-workbench-awd-running').text()).toBe('运维')
    expect(wrapper.findAll('.contest-action').length).toBe(2)
    expect(wrapper.get('#contest-row-edit-awd-running').text()).toContain('编辑')

    await wrapper.get('#contest-row-edit-awd-running').trigger('click')
    await flushPromises()

    expect(wrapper.emitted('edit')?.[0]?.[0]).toMatchObject({ id: 'awd-running' })

    await wrapper.get('#contest-row-more-awd-running').trigger('click')
    await flushPromises()

    const teleportedMenu = document.body.querySelector('[data-action-menu-panel]')
    expect(teleportedMenu).not.toBeNull()
    expect(wrapper.find('.workspace-directory-list [data-action-menu-panel]').exists()).toBe(false)
    const exportButton = document.body.querySelector<HTMLButtonElement>(
      '#contest-row-menu-export-awd-running'
    )
    expect(document.body.querySelector('#contest-row-menu-edit-awd-running')).toBeNull()
    expect(exportButton).toBeNull()

    wrapper.unmount()
  })

  it('行内操作应使用固定槽位，避免不同按钮数量导致编辑和更多入口错位', () => {
    expect(adminContestTableSource).toContain('contestTableColumns')
    expect(adminContestTableSource).toContain("{ key: 'actions', label: '操作', widthClass: 'w-[14rem]', align: 'right' as const }")
    expect(adminContestTableSource).toContain('class="ui-row-actions contest-table__actions ui-row-actions--fixed"')
    expect(adminContestTableSource).toContain('ui-row-action--main')
    expect(adminContestTableSource).toContain('ui-row-action--default')
    expect(adminContestTableSource).toContain('ui-row-action--menu')
    expect(adminContestTableSource).toContain('--ui-row-action-main-width: 4.25rem;')
    expect(adminContestTableSource).not.toContain('Swords')
    expect(adminContestTableSource).not.toContain('--contest-directory-action-column')
    expect(adminContestTableSource).not.toContain('--contest-action-workbench-width')
    expect(adminContestTableSource).not.toContain('--contest-action-button-width')
    expect(adminContestTableSource).not.toContain('--contest-action-menu-width')
    expect(adminContestTableSource).not.toContain('minmax(11rem, 11rem)')
  })

  it('未结束竞赛的更多菜单应提供发布通知入口并抛出 announce 事件', async () => {
    const wrapper = mount(PlatformContestTable, {
      attachTo: document.body,
      props: {
        contests: [buildContest()],
        page: 1,
        pageSize: 20,
        total: 1,
      },
    })

    await wrapper.get('#contest-row-more-awd-running').trigger('click')
    await flushPromises()

    const announceButton = document.body.querySelector<HTMLButtonElement>(
      '#contest-row-menu-announce-awd-running'
    )
    expect(announceButton?.textContent).toContain('发布通知')

    announceButton?.click()
    await flushPromises()

    expect(wrapper.emitted('announce')?.[0]?.[0]).toMatchObject({ id: 'awd-running' })

    wrapper.unmount()
  })

  it('已结束竞赛不显示更多菜单，但仍可进入运维与编辑', async () => {
    const wrapper = mount(PlatformContestTable, {
      attachTo: document.body,
      props: {
        contests: [buildContest({ id: 'contest-ended', status: 'ended' })],
        page: 1,
        pageSize: 20,
        total: 1,
      },
    })

    expect(wrapper.get('#contest-open-workbench-contest-ended').text()).toBe('运维')
    expect(wrapper.get('#contest-row-edit-contest-ended').text()).toContain('编辑')
    expect(wrapper.find('#contest-row-more-contest-ended').exists()).toBe(false)

    await wrapper.get('#contest-open-workbench-contest-ended').trigger('click')
    await wrapper.get('#contest-row-edit-contest-ended').trigger('click')
    await flushPromises()

    expect(wrapper.emitted('workbench')?.[0]?.[0]).toMatchObject({ id: 'contest-ended' })
    expect(wrapper.emitted('edit')?.[0]?.[0]).toMatchObject({ id: 'contest-ended' })

    wrapper.unmount()
  })
})
