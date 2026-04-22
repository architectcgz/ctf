<template>
  <AuthEntryShell
    panel-eyebrow="Account Access"
    panel-title="登录平台"
    panel-description="使用你的训练账号进入工作台。"
  >
    <form
      class="auth-login-form"
      @submit.prevent="onSubmit"
    >
      <div class="ui-field">
        <label class="ui-field__label">用户名</label>
        <div class="ui-control-wrap">
          <input
            ref="usernameInput"
            v-model="form.username"
            autocomplete="username"
            placeholder="输入用户名/学号"
            @input="submitError = ''"
            @keyup.enter="onSubmit"
          >
        </div>
      </div>

      <div class="ui-field">
        <div class="ui-field__head">
          <label class="ui-field__label">密码</label>
        </div>
        <div class="ui-control-wrap">
          <input
            ref="passwordInput"
            v-model="form.password"
            type="password"
            autocomplete="current-password"
            placeholder="输入你的账号密码"
            @input="submitError = ''"
          >
        </div>
      </div>

      <div
        v-if="submitError"
        class="ui-field__error auth-login-error"
      >
        {{ submitError }}
      </div>

      <button
        class="ui-btn ui-btn--primary ui-btn--block auth-login-submit"
        type="submit"
        :disabled="loading"
      >
        {{ loading ? '正在验证身份...' : '立即登录' }}
      </button>
    </form>

    <template #footer>
      <div class="auth-login-footer">
        没有账号？
        <RouterLink
          class="ui-btn ui-btn--link"
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
const submitError = ref('')
const form = reactive({ username: '', password: '' })
const usernameInput = useTemplateRef<HTMLInputElement>('usernameInput')
const passwordInput = useTemplateRef<HTMLInputElement>('passwordInput')

async function onSubmit() {
  form.username = usernameInput.value?.value || form.username
  form.password = passwordInput.value?.value || form.password

  if (!form.username || !form.password) return
  loading.value = true
  submitError.value = ''
  try {
    const redirectTo = sanitizeRedirectPath(route.query.redirect)
    await login(
      { username: form.username, password: form.password },
      redirectTo === '/' ? undefined : redirectTo
    )
  } catch (err) {
    submitError.value = err instanceof Error && err.message.trim() ? err.message : '登录失败，请稍后重试'
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

.auth-login-error {
  padding: var(--space-3) var(--space-4);
  background: var(--color-danger-soft);
  border: 1px solid color-mix(in srgb, var(--color-danger) 20%, transparent);
  border-radius: 0.75rem;
  font-size: var(--font-size-13);
}

.auth-login-submit {
  margin-top: var(--space-2);
  min-height: var(--ui-control-height-lg);
  font-size: var(--font-size-15);
  font-weight: 800;
}

.auth-login-footer {
  font-size: var(--font-size-14);
  font-weight: 500;
  color: var(--color-text-secondary);
}

.auth-login-footer .ui-btn--link {
  font-weight: 800;
  text-decoration: underline;
  text-underline-offset: 0.3em;
}
</style>