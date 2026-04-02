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

    <div
      v-if="casLoading || casStatus?.enabled"
      class="auth-login-form__cas"
    >
      <div class="auth-login-form__cas-eyebrow">Campus SSO</div>
      <div class="auth-login-form__cas-title">CAS 统一认证</div>
      <p class="auth-login-form__cas-copy">
        {{
          casLoading
            ? '正在检查学校统一认证入口状态...'
            : casReady
              ? '当前环境已启用 CAS 登录，可使用校园统一认证直接进入平台。'
              : 'CAS 已启用但配置未完成，当前环境暂时无法直接跳转。'
        }}
      </p>
      <ElButton
        class="auth-login-form__cas-button"
        size="large"
        :disabled="!casReady"
        :loading="casRedirecting"
        @click="onCASLogin"
      >
        使用 CAS 统一认证登录
      </ElButton>
    </div>

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
import { onMounted, reactive, ref, useTemplateRef } from 'vue'
import { RouterLink, useRoute } from 'vue-router'

import AuthEntryShell from '@/components/auth/AuthEntryShell.vue'
import { sanitizeRedirectPath } from '@/router/guards'
import { useAuth } from '@/composables/useAuth'
import { useCASAuth } from '@/composables/useCASAuth'

const { login } = useAuth()
const {
  casStatus,
  casLoading,
  casReady,
  casRedirecting,
  fetchCASStatus,
  beginCASLogin,
} = useCASAuth()
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

async function onCASLogin() {
  const redirectTo = sanitizeRedirectPath(route.query.redirect)
  await beginCASLogin(redirectTo === '/' ? '/dashboard' : redirectTo)
}

onMounted(() => {
  void fetchCASStatus()
})
</script>

<style scoped>
.auth-login-form__submit,
.auth-login-form__cas-button {
  width: 100%;
}

.auth-login-form__submit {
  margin-top: 0.5rem;
}

.auth-login-form__cas {
  margin-top: 1rem;
  padding: 1rem;
  border: 1px solid color-mix(in srgb, var(--color-border-default) 80%, transparent);
  border-radius: 1rem;
  background: color-mix(in srgb, var(--color-bg-elevated) 78%, transparent);
}

.auth-login-form__cas-eyebrow {
  font-size: 0.7rem;
  font-weight: 700;
  letter-spacing: 0.16em;
  text-transform: uppercase;
  color: var(--color-text-muted);
}

.auth-login-form__cas-title {
  margin-top: 0.45rem;
  font-size: 0.98rem;
  font-weight: 600;
  color: var(--color-text-primary);
}

.auth-login-form__cas-copy,
.auth-login-form__footer {
  color: var(--color-text-secondary);
  line-height: 1.7;
}

.auth-login-form__cas-copy {
  margin-top: 0.45rem;
  font-size: 0.84rem;
}

.auth-login-form__cas-button {
  margin-top: 0.85rem;
}

.auth-login-form__footer {
  font-size: 0.88rem;
}

.auth-login-form__link {
  color: var(--color-primary-hover);
  font-weight: 600;
}
</style>
