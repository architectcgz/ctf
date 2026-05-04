import { computed, nextTick, ref } from 'vue'
import { describe, expect, it } from 'vitest'

import type { AWDDefenseServiceCard } from './awdDefensePresentation'
import { useAwdDefenseServiceSelection } from './useAwdDefenseServiceSelection'

function card(serviceId: string): AWDDefenseServiceCard {
  return {
    serviceId,
    challengeId: `challenge-${serviceId}`,
    title: serviceId,
    riskLevel: 'stable',
    riskReasons: [],
    serviceStatusLabel: '正常',
    instanceStatusLabel: '平台代理已就绪',
    canOpenService: true,
    canRequestSSH: true,
    canRestart: true,
  }
}

describe('useAwdDefenseServiceSelection', () => {
  it('默认选择当前排序后的第一个服务', () => {
    const services = ref([card('critical'), card('stable')])
    const { selectedServiceId } = useAwdDefenseServiceSelection(computed(() => services.value))

    expect(selectedServiceId.value).toBe('critical')
  })

  it('刷新后保留仍然存在的手动选择', async () => {
    const services = ref([card('critical'), card('stable')])
    const { selectedServiceId, selectService } = useAwdDefenseServiceSelection(
      computed(() => services.value)
    )

    selectService('stable')
    services.value = [card('stable'), card('critical')]
    await nextTick()

    expect(selectedServiceId.value).toBe('stable')
  })

  it('选中服务消失后回退到第一个服务', async () => {
    const services = ref([card('critical'), card('stable')])
    const { selectedServiceId, selectService } = useAwdDefenseServiceSelection(
      computed(() => services.value)
    )

    selectService('stable')
    services.value = [card('critical')]
    await nextTick()

    expect(selectedServiceId.value).toBe('critical')
  })
})
