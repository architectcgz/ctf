<template>
  <AuthEntryShell
    panel-eyebrow="New Account"
    panel-title="创建账号"
    panel-description="填写基础信息后进入平台。"
  >
    <form
      class="auth-register-form"
      @submit.prevent="onSubmit"
    >
      <label class="auth-register-form__field">
        <span class="ui-field__label">用户名</span>
        <div class="ui-control-wrap">
          <input
            v-model="form.username"
            autocomplete="username"
            class="ui-control"
          >
        </div>
      </label>
      <label class="auth-register-form__field">
        <span class="ui-field__label">密码</span>
        <div class="ui-control-wrap">
          <input
            v-model="form.password"
            type="password"
            autocomplete="new-password"
            class="ui-control"
          >
        </div>
      </label>
      <label class="auth-register-form__field">
        <span class="ui-field__label">班级（可选）</span>
        <div class="ui-control-wrap">
          <input
            v-model="form.class_name"
            class="ui-control"
          >
        </div>
      </label>

      <button
        class="ui-btn ui-btn--primary ui-btn--block auth-register-form__submit"
        type="submit"
        :disabled="loading"
      >
        {{ loading ? '注册中…' : '注册' }}
      </button>
    </form>

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
.auth-register-form {
  display: grid;
  gap: 1rem;
}

.auth-register-form__field {
  display: grid;
  gap: 0.45rem;
}

.auth-register-form__submit {
  margin-top: 0.5rem;
}

.auth-register-form__footer {
  color: var(--color-text-secondary);
  font-size: var(--font-size-0-88);
  line-height: 1.7;
}

.auth-register-form__link {
  color: var(--color-primary-hover);
  font-weight: 600;
}
</style>
