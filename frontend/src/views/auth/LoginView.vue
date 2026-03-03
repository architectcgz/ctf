<template>
  <div class="min-h-screen bg-base text-text-primary">
    <div class="mx-auto flex min-h-screen max-w-md items-center px-4">
      <ElCard class="w-full" shadow="never">
        <template #header>
          <div class="text-base font-semibold">登录</div>
        </template>

        <ElForm :model="form" label-position="top" @submit.prevent="onSubmit">
          <ElFormItem label="用户名">
            <ElInput v-model="form.username" autocomplete="username" />
          </ElFormItem>
          <ElFormItem label="密码">
            <ElInput v-model="form.password" type="password" autocomplete="current-password" show-password />
          </ElFormItem>

          <div class="mt-4 flex items-center justify-between">
            <RouterLink class="text-sm text-primary hover:underline" to="/register">去注册</RouterLink>
            <ElButton type="primary" :loading="loading" @click="onSubmit">登录</ElButton>
          </div>
        </ElForm>
      </ElCard>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { RouterLink, useRoute } from 'vue-router'

import { useAuth } from '@/composables/useAuth'

const { login } = useAuth()
const route = useRoute()

const loading = ref(false)
const form = reactive({ username: '', password: '' })

function sanitizeRedirectPath(input: unknown): string {
  if (typeof input !== 'string') return '/dashboard'
  if (!input.startsWith('/')) return '/dashboard'
  if (input.startsWith('//')) return '/dashboard'
  return input
}

async function onSubmit() {
  if (!form.username || !form.password) return
  loading.value = true
  try {
    const redirectTo = sanitizeRedirectPath(route.query.redirect)
    await login({ username: form.username, password: form.password }, redirectTo)
  } finally {
    loading.value = false
  }
}
</script>

