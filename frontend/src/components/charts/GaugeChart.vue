<script setup lang="ts">
import { computed } from 'vue'
import type { EChartsOption } from 'echarts'
import { use } from 'echarts/core'
import { GaugeChart as EChartsGaugeChart } from 'echarts/charts'
import { CanvasRenderer } from 'echarts/renderers'
import VChart from 'vue-echarts'

use([EChartsGaugeChart, CanvasRenderer])

const props = withDefaults(defineProps<{
  value: number
  min?: number
  max?: number
  name?: string
}>(), {
  min: 0,
  max: 100,
  name: '完成度',
})

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
        color: '#e6edf3',
      },
      title: {
        color: '#8b949e',
      },
      data: [{ value: props.value, name: props.name }],
    },
  ],
}))
</script>

<template>
  <VChart class="h-72 w-full" :option="option" autoresize />
</template>
