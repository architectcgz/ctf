<template>
  <div class="min-h-screen bg-base text-text-primary">
    <div class="mx-auto flex min-h-screen max-w-md items-center px-4">
      <ElCard class="w-full" shadow="never">
        <template #header>
          <div class="text-base font-semibold">注册</div>
        </template>

        <ElForm :model="form" label-position="top" @submit.prevent="onSubmit">
          <ElFormItem label="用户名">
            <ElInput v-model="form.username" autocomplete="username" />
          </ElFormItem>
          <ElFormItem label="密码">
            <ElInput v-model="form.password" type="password" autocomplete="new-password" show-password />
          </ElFormItem>
          <ElFormItem label="班级（可选）">
            <ElInput v-model="form.class_name" />
          </ElFormItem>

          <div class="mt-4 flex items-center justify-between">
            <RouterLink class="text-sm text-primary hover:underline" to="/login">去登录</RouterLink>
            <ElButton type="primary" :loading="loading" @click="onSubmit">注册</ElButton>
          </div>
        </ElForm>
      </ElCard>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { RouterLink } from 'vue-router'

import { useAuth } from '@/composables/useAuth'

const { register } = useAuth()

const loading = ref(false)
const form = reactive({ username: '', password: '', class_name: '' })

async function onSubmit() {
  if (!form.username || !form.password) return
  loading.value = true
  try {
    await register({ username: form.username, password: form.password, class_name: form.class_name || undefined })
  } finally {
    loading.value = false
  }
}
</script>

