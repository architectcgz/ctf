<script setup lang="ts">
import { computed } from 'vue'
import type { EChartsOption } from 'echarts'
import { use } from 'echarts/core'
import { GaugeChart as EChartsGaugeChart } from 'echarts/charts'
import { CanvasRenderer } from 'echarts/renderers'
import VChart from 'vue-echarts'
import { useEChartsMountGate } from '@/components/charts/echartsMountGate'

use([EChartsGaugeChart, CanvasRenderer])

const props = withDefaults(
  defineProps<{
    value: number
    min?: number
    max?: number
    name?: string
  }>(),
  {
    min: 0,
    max: 100,
    name: '完成度',
  }
)
const { containerRef, isChartReady } = useEChartsMountGate()

function cssVar(name: string): string {
  return getComputedStyle(document.documentElement).getPropertyValue(name).trim()
}

const option = computed<EChartsOption>(() => ({
  series: [
    {
      type: 'gauge',
      min: props.min,
      max: props.max,
      progress: { show: true, width: 16 },
      axisLine: { lineStyle: { width: 16 } },
      detail: {
        valueAnimation: true,
        formatter: '{value}',
        color: cssVar('--color-text-primary'),
      },
      title: {
        color: cssVar('--color-text-secondary'),
      },
      data: [{ value: props.value, name: props.name }],
    },
  ],
}))
</script>

<template>
  <div ref="containerRef" class="h-72 w-full">
    <VChart
      v-if="isChartReady"
      class="h-full w-full"
      :option="option"
      autoresize
    />
  </div>
</template>
