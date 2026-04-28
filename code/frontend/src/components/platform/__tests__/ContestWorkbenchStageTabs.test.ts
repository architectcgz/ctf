import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import ContestWorkbenchStageTabs from '../contest/ContestWorkbenchStageTabs.vue'

type ContestWorkbenchStageKey =
  | 'basics'
  | 'pool'
  | 'awd-config'
  | 'preflight'

interface ContestWorkbenchStage {
  key: ContestWorkbenchStageKey
  label: string
  disabled?: boolean
}

function buildStages(): ContestWorkbenchStage[] {
  return [
    { key: 'basics', label: '基础信息' },
    { key: 'pool', label: '题目池' },
    { key: 'awd-config', label: 'AWD 配置' },
    { key: 'preflight', label: '赛前检查' },
  ]
}

describe('ContestWorkbenchStageTabs', () => {
  it('应该按传入阶段渲染按钮并高亮当前阶段', async () => {
    const selectStage = vi.fn()
    const wrapper = mount(ContestWorkbenchStageTabs, {
      props: {
        stages: buildStages(),
        activeStage: 'pool',
        selectStage,
      },
    })

    expect(wrapper.findAll('[role="tablist"]')).toHaveLength(1)
    expect(wrapper.get('[role="tablist"]').attributes('aria-label')).toBe('竞赛工作台阶段切换')
    expect(wrapper.findAll('.top-tab')).toHaveLength(4)
    expect(wrapper.get('.top-tab.active').text()).toContain('题目池')

    const preflightTab = wrapper
      .findAll('.top-tab')
      .find((node) => node.text().includes('赛前检查'))

    expect(preflightTab).toBeDefined()

    await preflightTab!.trigger('click')

    expect(selectStage).toHaveBeenCalledWith('preflight')
  })
})
