<script setup lang="ts">
import { ArrowLeft, RotateCcw, ServerCrash } from 'lucide-vue-next'

import ErrorStatusShell from '@/components/errors/ErrorStatusShell.vue'
</script>

<template>
  <ErrorStatusShell
    status-code="500"
    kicker="500 · Internal Server Error"
    title="系统内部错误"
    description="服务端在处理当前请求时发生异常。你可以先刷新当前页面重试；若仍未恢复，先返回上一页，必要时再回工作台继续其他操作。"
    accent="danger"
    :icon="ServerCrash"
    :primary-icon="RotateCcw"
    :secondary-icon="ArrowLeft"
    primary-action="reload"
    secondary-action="back"
  />
</template>
