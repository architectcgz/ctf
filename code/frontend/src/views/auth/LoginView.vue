<template>
  <AuthEntryShell
    panel-eyebrow="Account Access"
    panel-title="登录平台"
    panel-description="使用你的训练账号进入工作台。"
  >
    <ElForm
      class="auth-login-form"
      :model="form"
      label-position="top"
      @submit.prevent="onSubmit"
    >
      <ElFormItem label="用户名">
        <ElInput
          ref="usernameInput"
          v-model="form.username"
          autocomplete="username"
          size="large"
          @keyup.enter="onSubmit"
        />
      </ElFormItem>
      <ElFormItem label="密码">
        <ElInput
          ref="passwordInput"
          v-model="form.password"
          type="password"
          autocomplete="current-password"
          show-password
          size="large"
        />
      </ElFormItem>

      <ElButton
        class="auth-login-form__submit"
        type="primary"
        size="large"
        :loading="loading"
        @click="onSubmit"
      >
        登录
      </ElButton>
    </ElForm>

    <template #footer>
      <div class="auth-login-form__footer">
        没有账号？
        <RouterLink
          class="auth-login-form__link"
          to="/register"
        >
          创建新账号
        </RouterLink>
      </div>
    </template>
  </AuthEntryShell>
</template>

<script setup lang="ts">
import { reactive, ref, useTemplateRef } from 'vue'
import { RouterLink, useRoute } from 'vue-router'

import AuthEntryShell from '@/components/auth/AuthEntryShell.vue'
import { sanitizeRedirectPath } from '@/router/guards'
import { useAuth } from '@/composables/useAuth'

const { login } = useAuth()
const route = useRoute()

const loading = ref(false)
const form = reactive({ username: '', password: '' })
const usernameInput = useTemplateRef<{ input?: HTMLInputElement }>('usernameInput')
const passwordInput = useTemplateRef<{ input?: HTMLInputElement }>('passwordInput')

async function onSubmit() {
  form.username = usernameInput.value?.input?.value || form.username
  form.password = passwordInput.value?.input?.value || form.password

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

<style scoped>
.auth-login-form__submit {
  width: 100%;
  margin-top: 0.5rem;
}

.auth-login-form__footer {
  color: var(--color-text-secondary);
  line-height: 1.7;
  font-size: 0.88rem;
}

.auth-login-form__link {
  color: var(--color-primary-hover);
  font-weight: 600;
}
</style>
