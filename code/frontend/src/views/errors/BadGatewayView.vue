<script setup lang="ts">
import { ArrowLeft, Network, RotateCcw } from 'lucide-vue-next'

import ErrorStatusShell from '@/components/errors/ErrorStatusShell.vue'
</script>

<template>
  <ErrorStatusShell
    status-code="502"
    kicker="502 · Bad Gateway"
    title="上游服务响应异常"
    description="当前服务依赖的上游节点返回了无效响应。你可以先刷新当前页面重试；若问题持续，先返回上一页，必要时再回工作台查看其他内容。"
    accent="danger"
    :icon="Network"
    :primary-icon="RotateCcw"
    :secondary-icon="ArrowLeft"
    primary-action="reload"
    secondary-action="back"
  />
</template>
