<script setup lang="ts">
import { ArrowLeft, LogIn, ShieldEllipsis } from 'lucide-vue-next'

import ErrorStatusShell from '@/components/errors/ErrorStatusShell.vue'
</script>

<template>
  <ErrorStatusShell
    status-code="401"
    kicker="401 · Unauthorized"
    title="登录状态已失效"
    description="当前请求需要有效身份认证。请重新登录后再继续访问对应页面。"
    accent="warning"
    :icon="ShieldEllipsis"
    :primary-icon="LogIn"
    :secondary-icon="ArrowLeft"
    primary-to="/login"
    primary-label="返回登录页"
    secondary-action="back"
  />
</template>
