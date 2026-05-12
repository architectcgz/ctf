import { nextTick } from 'vue'
import { mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import LineChart from '../LineChart.vue'
import lineChartSource from '../LineChart.vue?raw'
import barChartSource from '../BarChart.vue?raw'
import gaugeChartSource from '../GaugeChart.vue?raw'
import radarChartSource from '../RadarChart.vue?raw'

vi.mock('vue-echarts', () => ({
  default: {
    name: 'VChart',
    template: '<div data-testid="echart-instance" />',
  },
}))

class ResizeObserverStub {
  static instances: ResizeObserverStub[] = []

  private readonly callback: ResizeObserverCallback

  observe = vi.fn()
  disconnect = vi.fn()

  constructor(callback: ResizeObserverCallback) {
    this.callback = callback
    ResizeObserverStub.instances.push(this)
  }

  trigger(target: Element) {
    this.callback(
      [
        {
          target,
          contentRect: {
            width: (target as HTMLElement).clientWidth,
            height: (target as HTMLElement).clientHeight,
          },
        } as ResizeObserverEntry,
      ],
      this as unknown as ResizeObserver
    )
  }
}

describe('ECharts mount gate', () => {
  beforeEach(() => {
    ResizeObserverStub.instances = []
    vi.stubGlobal('ResizeObserver', ResizeObserverStub)
  })

  afterEach(() => {
    document.body.innerHTML = ''
  })

  it('容器尺寸为 0 时不应先挂载折线图，尺寸可用后再初始化', async () => {
    const wrapper = mount(LineChart, {
      attachTo: document.body,
      props: {
        categories: ['05-01', '05-02'],
        series: [{ name: '训练事件', data: [2, 4] }],
      },
    })

    expect(wrapper.find('[data-testid="echart-instance"]').exists()).toBe(false)

    const chartContainer = wrapper.element as HTMLElement
    Object.defineProperty(chartContainer, 'clientWidth', {
      configurable: true,
      value: 640,
    })
    Object.defineProperty(chartContainer, 'clientHeight', {
      configurable: true,
      value: 320,
    })

    ResizeObserverStub.instances[0]?.trigger(chartContainer)
    await nextTick()

    expect(wrapper.find('[data-testid="echart-instance"]').exists()).toBe(true)
  })

  it('共享图表组件应统一使用尺寸门槛，避免在隐藏容器里初始化 ECharts', () => {
    expect(lineChartSource).toContain('useEChartsMountGate')
    expect(barChartSource).toContain('useEChartsMountGate')
    expect(gaugeChartSource).toContain('useEChartsMountGate')
    expect(radarChartSource).toContain('useEChartsMountGate')
  })
})
