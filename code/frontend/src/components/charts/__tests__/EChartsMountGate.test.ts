import { nextTick } from 'vue'
import { mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import RadarChart from '../RadarChart.vue'
import LineChart from '../LineChart.vue'
import lineChartSource from '../LineChart.vue?raw'
import barChartSource from '../BarChart.vue?raw'
import gaugeChartSource from '../GaugeChart.vue?raw'
import radarChartSource from '../RadarChart.vue?raw'
import { useTheme } from '@/composables/useTheme'

vi.mock('vue-echarts', () => ({
  default: {
    name: 'VChart',
    props: {
      option: {
        type: Object,
        default: null,
      },
      autoresize: {
        type: Boolean,
        default: false,
      },
    },
    template: '<div data-testid="echart-instance" />',
  },
}))

interface RadarOption {
  radar?: {
    axisName?: {
      color?: string
    }
  }
}

const mockedThemeColors: Record<'dark' | 'light', Record<string, string>> = {
  dark: {
    '--color-text-primary': 'dark-text-primary',
    '--color-text-secondary': 'dark-text-secondary',
    '--color-border-default': 'dark-border-default',
    '--color-border-subtle': 'dark-border-subtle',
    '--color-primary': 'dark-primary',
    '--color-primary-hover': 'dark-primary-hover',
  },
  light: {
    '--color-text-primary': 'light-text-primary',
    '--color-text-secondary': 'light-text-secondary',
    '--color-border-default': 'light-border-default',
    '--color-border-subtle': 'light-border-subtle',
    '--color-primary': 'light-primary',
    '--color-primary-hover': 'light-primary-hover',
  },
}

function resolveMockThemeColor(name: string): string {
  const themeName = document.documentElement.getAttribute('data-theme') === 'light' ? 'light' : 'dark'
  return mockedThemeColors[themeName][name] ?? ''
}

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
    expect(lineChartSource).toContain('void theme.value')
    expect(barChartSource).toContain('void theme.value')
    expect(gaugeChartSource).toContain('void theme.value')
    expect(radarChartSource).toContain('void theme.value')
  })

  it('雷达图在切换主题后应重新取用当前主题的轴标签颜色', async () => {
    localStorage.clear()
    const { initTheme, toggleTheme, setTheme } = useTheme()
    initTheme()
    const getComputedStyleSpy = vi.spyOn(window, 'getComputedStyle').mockImplementation(() => ({
      getPropertyValue: (name: string) => resolveMockThemeColor(name),
    } as CSSStyleDeclaration))

    try {
      const wrapper = mount(RadarChart, {
        attachTo: document.body,
        props: {
          indicators: [
            { name: 'Web', max: 100 },
            { name: 'Crypto', max: 100 },
          ],
          values: [82, 55],
        },
      })

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

      const chart = wrapper.getComponent({ name: 'VChart' })
      const darkAxisColor = (chart.props('option') as RadarOption).radar?.axisName?.color

      expect(darkAxisColor).toBe(resolveMockThemeColor('--color-text-primary'))

      toggleTheme()
      await nextTick()

      const lightAxisColor = (chart.props('option') as RadarOption).radar?.axisName?.color

      expect(lightAxisColor).toBe(resolveMockThemeColor('--color-text-primary'))
      expect(lightAxisColor).not.toBe(darkAxisColor)
    } finally {
      getComputedStyleSpy.mockRestore()
      setTheme('dark')
      await nextTick()
    }
  })
})
