import { defineComponent, h, nextTick, ref } from 'vue'
import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import ContestWorkbenchStageRail from '../contest/ContestWorkbenchStageRail.vue'

type ContestWorkbenchStageKey =
  | 'basics'
  | 'pool'
  | 'awd-config'
  | 'preflight'
  | 'operations'

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
    { key: 'operations', label: '轮次运行' },
  ]
}

describe('ContestWorkbenchStageRail', () => {
  it('应该按传入阶段渲染按钮并高亮当前阶段', async () => {
    const selectStage = vi.fn()
    const wrapper = mount(ContestWorkbenchStageRail, {
      props: {
        stages: buildStages(),
        activeStage: 'pool',
        selectStage,
      },
    })

    expect(wrapper.findAll('[role="tab"]')).toHaveLength(5)
    expect(wrapper.get('[role="tab"][aria-selected="true"]').text()).toContain('题目池')

    const operationsTab = wrapper
      .findAll('[role="tab"]')
      .find((node) => node.text().includes('轮次运行'))

    expect(operationsTab).toBeDefined()

    await operationsTab!.trigger('click')

    expect(selectStage).toHaveBeenCalledWith('operations')
  })

  it('应该跳过 disabled 阶段并保持 roving tabindex', async () => {
    const stages = buildStages().map((stage) =>
      stage.key === 'awd-config' ? { ...stage, disabled: true } : stage
    )
    const selectStage = vi.fn()
    const activeStage = ref<ContestWorkbenchStageKey>('pool')

    const wrapper = mount(
      defineComponent({
        setup() {
          return () =>
            h(ContestWorkbenchStageRail, {
              stages,
              activeStage: activeStage.value,
              selectStage: (stage: ContestWorkbenchStageKey) => {
                activeStage.value = stage
                selectStage(stage)
              },
            })
        },
      })
    )

    const poolTab = wrapper.get('[role="tab"][aria-selected="true"]')

    expect(poolTab.text()).toContain('题目池')
    expect(poolTab.attributes('tabindex')).toBe('0')

    await poolTab.trigger('keydown', { key: 'ArrowRight' })
    await nextTick()

    expect(selectStage).toHaveBeenCalledWith('preflight')
    expect(wrapper.get('[role="tab"][aria-selected="true"]').text()).toContain('赛前检查')

    const rovingTabs = wrapper.findAll('[role="tab"]')
    const preflightTab = rovingTabs.find((node) => node.text().includes('赛前检查'))
    const disabledTab = rovingTabs.find((node) => node.text().includes('AWD 配置'))

    expect(preflightTab?.attributes('tabindex')).toBe('0')
    expect(disabledTab?.attributes('tabindex')).toBe('-1')
  })
})
