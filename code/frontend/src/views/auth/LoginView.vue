<template>
  <AuthEntryShell
    panel-eyebrow="账户访问"
    panel-title="登录工作台"
    panel-description="验证凭据以进入安全实战系统。"
  >
    <form
      class="auth-login-form"
      @submit.prevent="onSubmit"
    >
      <div class="auth-field">
        <label class="auth-label">用户名 / 学号</label>
        <div class="ui-control-wrap">
          <input
            ref="usernameInput"
            v-model="form.username"
            autocomplete="username"
            class="ui-control"
            placeholder="输入您的登录名"
            @input="submitError = ''"
            @keyup.enter="onSubmit"
          >
        </div>
      </div>

      <div class="auth-field">
        <div class="auth-label-row">
          <label class="auth-label">安全密码</label>
          <button type="button" class="auth-aux-link">忘记密码?</button>
        </div>
        <div class="ui-control-wrap">
          <input
            ref="passwordInput"
            v-model="form.password"
            type="password"
            autocomplete="current-password"
            class="ui-control"
            placeholder="输入您的访问密码"
            @input="submitError = ''"
          >
        </div>
      </div>

      <div
        v-if="submitError"
        class="auth-error-block"
      >
        {{ submitError }}
      </div>

      <button
        class="ui-btn ui-btn--primary ui-btn--block auth-submit-btn"
        type="submit"
        :disabled="loading"
      >
        <span v-if="loading">正在验证身份...</span>
        <span v-else>立即登录系统</span>
      </button>
    </form>

    <template #footer>
      <div class="auth-footer-nav">
        <span>还没有平台账号？</span>
        <p class="auth-contact-hint">请联系您的系统管理员进行账号分配与导入</p>
      </div>
    </template>
  </AuthEntryShell>
</template>

<script setup lang="ts">
import { reactive, ref, useTemplateRef } from 'vue'
import { useRoute } from 'vue-router'

import AuthEntryShell from '@/components/auth/AuthEntryShell.vue'
import { sanitizeRedirectPath } from '@/router/guards'
import { useAuth } from '@/composables/useAuth'

const { login } = useAuth()
const route = useRoute()

const loading = ref(false)
const submitError = ref('')
const form = reactive({ username: '', password: '' })
const usernameInput = useTemplateRef<HTMLInputElement>('usernameInput')
const passwordInput = useTemplateRef<HTMLInputElement>('passwordInput')

async function onSubmit() {
  if (loading.value || !form.username || !form.password) return
  loading.value = true
  submitError.value = ''
  try {
    const redirectTo = sanitizeRedirectPath(route.query.redirect)
    await login(
      { username: form.username, password: form.password },
      redirectTo === '/' ? undefined : redirectTo
    )
  } catch (err) {
    submitError.value = err instanceof Error && err.message.trim() ? err.message : '身份验证失败，请核对信息'
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.auth-login-form {
  display: grid;
  gap: var(--space-5);
}

.auth-field {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}

.auth-label {
  font-size: var(--font-size-10, 10px);
  font-weight: 800;
  text-transform: uppercase;
  letter-spacing: 0.15em;
  color: var(--color-text-muted);
}

.auth-label-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.auth-aux-link {
  font-size: var(--font-size-10, 10px);
  font-weight: 700;
  background: transparent;
  border: none;
  color: var(--color-primary);
  cursor: pointer;
  padding: 0;
}

.auth-error-block {
  padding: var(--space-3-5) var(--space-4);
  background: var(--color-danger-soft);
  color: var(--color-danger);
  border-radius: var(--space-3);
  font-size: var(--font-size-13);
  font-weight: 700;
  border: 1px solid color-mix(in srgb, var(--color-danger) 15%, transparent);
}

.auth-submit-btn {
  margin-top: var(--space-2);
  min-height: var(--space-12);
  font-size: var(--font-size-13);
  font-weight: 900;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  border-radius: var(--space-3-5);
  box-shadow: 0 10px 24px -8px color-mix(in srgb, var(--color-primary) 25%, transparent);
}

.auth-footer-nav {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
  align-items: center;
  font-size: var(--font-size-10, 10px);
  font-weight: 800;
  letter-spacing: 0.1em;
  color: var(--color-text-muted);
}

.auth-contact-hint {
  margin: var(--space-1) 0 0;
  color: var(--color-primary);
  font-size: var(--font-size-11);
  font-weight: 700;
}
</style>
