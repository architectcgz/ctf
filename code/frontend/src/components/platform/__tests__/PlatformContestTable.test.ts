import { afterEach, describe, expect, it } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import PlatformContestTable from '../contest/PlatformContestTable.vue'
import adminContestTableSource from '../contest/PlatformContestTable.vue?raw'
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

  it('应将编辑和导出结果收纳进更多菜单，并通过浮层渲染', async () => {
    const wrapper = mount(PlatformContestTable, {
      attachTo: document.body,
      props: {
        contests: [buildContest()],
        page: 1,
        pageSize: 20,
        total: 1,
      },
    })

    expect(wrapper.text()).toContain('进入 AWD 赛区')
    expect(wrapper.findAll('.contest-action').length).toBe(1)

    await wrapper.get('#contest-row-more-awd-running').trigger('click')
    await flushPromises()

    const teleportedMenu = document.body.querySelector('[data-action-menu-panel]')
    expect(teleportedMenu).not.toBeNull()
    expect(wrapper.find('.workspace-directory-list [data-action-menu-panel]').exists()).toBe(false)
    const editButton = document.body.querySelector<HTMLButtonElement>(
      '#contest-row-menu-edit-awd-running'
    )
    const exportButton = document.body.querySelector<HTMLButtonElement>(
      '#contest-row-menu-export-awd-running'
    )
    expect(editButton?.textContent).toContain('编辑')
    expect(exportButton?.textContent).toContain('导出结果')

    editButton?.click()
    await flushPromises()
    expect(wrapper.emitted('edit')?.[0]?.[0]).toMatchObject({ id: 'awd-running' })

    await wrapper.get('#contest-row-more-awd-running').trigger('click')
    await flushPromises()
    document.body.querySelector<HTMLButtonElement>('#contest-row-menu-export-awd-running')?.click()
    await flushPromises()
    expect(wrapper.emitted('export')?.[0]?.[0]).toMatchObject({ id: 'awd-running' })

    wrapper.unmount()
  })
})
