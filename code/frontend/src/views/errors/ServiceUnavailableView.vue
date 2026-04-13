<script setup lang="ts">
import { ArrowLeft, RotateCcw, ServerOff } from 'lucide-vue-next'

import ErrorStatusShell from '@/components/errors/ErrorStatusShell.vue'
</script>

<template>
  <ErrorStatusShell
    status-code="503"
    kicker="503 · Service Unavailable"
    title="服务暂时不可用"
    description="系统正在维护或暂时承压，当前无法完成请求。你可以先刷新当前页面确认是否恢复；若仍不可用，先返回上一页，必要时再回工作台继续其他操作。"
    accent="warning"
    :icon="ServerOff"
    :primary-icon="RotateCcw"
    :secondary-icon="ArrowLeft"
    primary-action="reload"
    secondary-action="back"
  />
</template>
