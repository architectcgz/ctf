<template>
  <AuthEntryShell
    panel-eyebrow="New Account"
    panel-title="注册账号"
    panel-description="填写必要的基础信息以完成平台注册。"
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
            placeholder="设置你的登录账号"
          >
        </div>
      </label>
      <label class="auth-register-form__field">
        <span class="ui-field__label">设置密码</span>
        <div class="ui-control-wrap">
          <input
            v-model="form.password"
            type="password"
            autocomplete="new-password"
            class="ui-control"
            placeholder="建议使用 8 位以上字母数字组合"
          >
        </div>
      </label>
      <label class="auth-register-form__field">
        <span class="ui-field__label">班级邀请码（可选）</span>
        <div class="ui-control-wrap">
          <input
            v-model="form.class_name"
            class="ui-control"
            placeholder="输入班级全称以自动加入"
          >
        </div>
      </label>

      <button
        class="ui-btn ui-btn--primary ui-btn--block auth-register-form__submit"
        type="submit"
        :disabled="loading"
      >
        {{ loading ? '正在提交注册...' : '立即创建账号' }}
      </button>
    </form>

    <template #footer>
      <div class="auth-register-form__footer">
        已经有账号了？
        <RouterLink
          class="auth-register-form__link"
          to="/login"
        >
          返回登录
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
  gap: var(--space-5);
}

.auth-register-form__field {
  display: grid;
  gap: var(--space-2);
}

.auth-register-form__submit {
  margin-top: var(--space-2);
  min-height: var(--ui-control-height-lg);
  font-size: var(--font-size-15);
  font-weight: 800;
  border-radius: var(--ui-control-radius-lg);
}

.auth-register-form__footer {
  color: var(--color-text-secondary);
  font-size: var(--font-size-14);
  line-height: 1.7;
  text-align: center;
}

.auth-register-form__link {
  color: var(--color-primary);
  font-weight: 800;
  margin-left: 0.25rem;
  text-decoration: underline;
  text-underline-offset: 0.25em;
  transition: color 0.2s ease;
}

.auth-register-form__link:hover {
  color: var(--color-primary-hover);
}
</style>