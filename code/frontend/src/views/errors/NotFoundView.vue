<script setup lang="ts">
import { ArrowLeft, Home, MapPinned } from 'lucide-vue-next'

import ErrorStatusShell from '@/components/errors/ErrorStatusShell.vue'
</script>

<template>
  <ErrorStatusShell
    status-code="404"
    kicker="404 · Not Found"
    title="页面不存在"
    description="你访问的地址不存在，可能已被移动或删除。建议先返回当前工作台，或回到上一页重新确认入口。"
    accent="warning"
    :icon="MapPinned"
    :primary-icon="Home"
    :secondary-icon="ArrowLeft"
    secondary-action="back"
  />
</template>
