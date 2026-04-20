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
      <label class="auth-login-form__field">
        <span class="ui-field__label">用户名</span>
        <div class="ui-control-wrap">
          <input
            ref="usernameInput"
            v-model="form.username"
            autocomplete="username"
            class="ui-control"
            @keyup.enter="onSubmit"
          >
        </div>
      </label>
      <label class="auth-login-form__field">
        <span class="ui-field__label">密码</span>
        <div class="ui-control-wrap">
          <input
            ref="passwordInput"
            v-model="form.password"
            type="password"
            autocomplete="current-password"
            class="ui-control"
          >
        </div>
      </label>

      <button
        class="ui-btn ui-btn--primary ui-btn--block auth-login-form__submit"
        type="submit"
        :disabled="loading"
      >
        {{ loading ? '登录中…' : '登录' }}
      </button>
    </form>

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
const usernameInput = useTemplateRef<HTMLInputElement>('usernameInput')
const passwordInput = useTemplateRef<HTMLInputElement>('passwordInput')

async function onSubmit() {
  form.username = usernameInput.value?.value || form.username
  form.password = passwordInput.value?.value || form.password

  if (!form.username || !form.password) return
  loading.value = true
  try {
    const redirectTo = sanitizeRedirectPath(route.query.redirect)
    await login(
      { username: form.username, password: form.password },
      redirectTo === '/' ? undefined : redirectTo
    )
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.auth-login-form {
  display: grid;
  gap: 1rem;
}

.auth-login-form__field {
  display: grid;
  gap: 0.45rem;
}

.auth-login-form__submit {
  margin-top: 0.5rem;
}

.auth-login-form__footer {
  color: var(--color-text-secondary);
  line-height: 1.7;
  font-size: var(--font-size-0-88);
}

.auth-login-form__link {
  color: var(--color-primary-hover);
  font-weight: 600;
}
</style>
