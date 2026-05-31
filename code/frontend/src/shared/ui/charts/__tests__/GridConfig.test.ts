import { describe, expect, it } from 'vitest'
import lineChartSource from '../LineChart.vue?raw'
import barChartSource from '../BarChart.vue?raw'

describe('chart grid config', () => {
  it('uses ECharts 6 outerBounds config in line chart', () => {
    expect(lineChartSource).not.toContain('containLabel: true')
    expect(lineChartSource).toContain("outerBoundsMode: 'same'")
    expect(lineChartSource).toContain("outerBoundsContain: 'axisLabel'")
  })

  it('uses ECharts 6 outerBounds config in bar chart', () => {
    expect(barChartSource).not.toContain('containLabel: true')
    expect(barChartSource).toContain("outerBoundsMode: 'same'")
    expect(barChartSource).toContain("outerBoundsContain: 'axisLabel'")
  })
})
