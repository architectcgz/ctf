<script setup lang="ts">
import { ArrowLeft, Gauge, Home } from 'lucide-vue-next'

import ErrorStatusShell from '@/components/errors/ErrorStatusShell.vue'
</script>

<template>
  <ErrorStatusShell
    status-code="429"
    kicker="429 · Too Many Requests"
    title="请求过于频繁"
    description="系统在短时间内收到了过多请求。建议稍后再试，或先返回上一页避免重复触发当前操作。"
    accent="warning"
    :icon="Gauge"
    :primary-icon="Home"
    :secondary-icon="ArrowLeft"
    secondary-action="back"
  />
</template>
