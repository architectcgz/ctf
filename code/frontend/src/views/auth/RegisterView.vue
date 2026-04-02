<template>
  <AuthEntryShell
    panel-eyebrow="New Account"
    panel-title="创建账号"
    panel-description="填写基础信息后进入平台。"
  >
    <ElForm
      class="auth-register-form"
      :model="form"
      label-position="top"
      @submit.prevent="onSubmit"
    >
      <ElFormItem label="用户名">
        <ElInput
          v-model="form.username"
          autocomplete="username"
          size="large"
        />
      </ElFormItem>
      <ElFormItem label="密码">
        <ElInput
          v-model="form.password"
          type="password"
          autocomplete="new-password"
          show-password
          size="large"
        />
      </ElFormItem>
      <ElFormItem label="班级（可选）">
        <ElInput
          v-model="form.class_name"
          size="large"
        />
      </ElFormItem>

      <ElButton
        class="auth-register-form__submit"
        type="primary"
        size="large"
        :loading="loading"
        @click="onSubmit"
      >
        注册
      </ElButton>
    </ElForm>

    <template #footer>
      <div class="auth-register-form__footer">
        已有账号，
        <RouterLink
          class="auth-register-form__link"
          to="/login"
        >
          去登录
        </RouterLink>
      </div>
    </template>
  </AuthEntryShell>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { RouterLink } from 'vue-router'

import AuthEntryShell from '@/components/auth/AuthEntryShell.vue'
import { useAuth } from '@/composables/useAuth'

const { register } = useAuth()

const loading = ref(false)
const form = reactive({ username: '', password: '', class_name: '' })

async function onSubmit() {
  if (!form.username || !form.password) return
  loading.value = true
  try {
    await register({
      username: form.username,
      password: form.password,
      class_name: form.class_name.trim() || undefined,
    })
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.auth-register-form__submit {
  width: 100%;
  margin-top: 0.5rem;
}

.auth-register-form__footer {
  color: var(--color-text-secondary);
  font-size: 0.88rem;
  line-height: 1.7;
}

.auth-register-form__link {
  color: var(--color-primary-hover);
  font-weight: 600;
}
</style>
