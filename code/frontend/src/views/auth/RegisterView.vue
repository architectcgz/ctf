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
      <div class="ui-field">
        <label class="ui-field__label">用户名</label>
        <div class="ui-control-wrap">
          <input
            v-model="form.username"
            autocomplete="username"
            placeholder="设置你的登录账号"
            @input="submitError = ''"
          >
        </div>
      </div>

      <div class="ui-field">
        <label class="ui-field__label">设置密码</label>
        <div class="ui-control-wrap">
          <input
            v-model="form.password"
            type="password"
            autocomplete="new-password"
            placeholder="建议使用 8 位以上字母数字组合"
            @input="submitError = ''"
          >
        </div>
      </div>

      <div class="ui-field">
        <label class="ui-field__label">班级邀请码（可选）</label>
        <div class="ui-control-wrap">
          <input
            v-model="form.class_name"
            placeholder="输入班级全称以自动加入"
            @input="submitError = ''"
          >
        </div>
      </div>

      <p
        v-if="submitError"
        class="ui-field__error auth-register-form__error"
      >
        {{ submitError }}
      </p>

      <button
        class="ui-btn ui-btn--primary ui-btn--block auth-register-submit"
        type="submit"
        :disabled="loading"
      >
        {{ loading ? '正在提交注册...' : '立即创建账号' }}
      </button>
    </form>

    <template #footer>
      <div class="auth-register-footer">
        已经有账号了？
        <RouterLink
          class="ui-btn ui-btn--link"
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
const submitError = ref('')
const form = reactive({ username: '', password: '', class_name: '' })

async function onSubmit() {
  if (loading.value || !form.username || !form.password) return
  loading.value = true
  submitError.value = ''
  try {
    await register({
      username: form.username,
      password: form.password,
      class_name: form.class_name.trim() || undefined,
    })
  } catch (err) {
    submitError.value = err instanceof Error && err.message.trim() ? err.message : '注册失败，请稍后重试'
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

.auth-register-submit {
  margin-top: var(--space-2);
  min-height: var(--ui-control-height-lg);
  font-size: var(--font-size-15);
  font-weight: 800;
}

.auth-register-form__error {
  margin: 0;
  padding: 0.75rem 1rem;
  border-radius: 0.75rem;
  background: var(--color-danger-soft);
  color: var(--color-danger);
  font-size: var(--font-size-13);
  font-weight: 700;
  border: 1px solid color-mix(in srgb, var(--color-danger) 20%, transparent);
}

.auth-register-footer {
  text-align: center;
  font-size: var(--font-size-14);
  font-weight: 500;
  color: var(--color-text-secondary);
}

.auth-register-footer .ui-btn--link {
  font-weight: 800;
  text-decoration: underline;
  text-underline-offset: 0.3em;
}
</style>
