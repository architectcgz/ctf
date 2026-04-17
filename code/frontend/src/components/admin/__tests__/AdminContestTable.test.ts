import { afterEach, describe, expect, it } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import AdminContestTable from '../contest/AdminContestTable.vue'
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

describe('AdminContestTable', () => {
  afterEach(() => {
    document.body.innerHTML = ''
  })

  it('应将编辑和导出结果收纳进更多菜单，并通过浮层渲染', async () => {
    const wrapper = mount(AdminContestTable, {
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

    const teleportedMenu = document.body.querySelector('.contest-row-menu')
    expect(teleportedMenu).not.toBeNull()
    expect(wrapper.find('.workspace-directory-list .contest-row-menu').exists()).toBe(false)
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
