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
        <label
          class="ui-field__label"
          for="register-username"
        >用户名</label>
        <div class="ui-control-wrap">
          <input
            id="register-username"
            v-model="form.username"
            autocomplete="username"
            class="ui-control"
            placeholder="设置你的登录账号"
            @input="clearSubmitError"
          >
        </div>
      </div>

      <div class="ui-field">
        <label
          class="ui-field__label"
          for="register-password"
        >设置密码</label>
        <div class="ui-control-wrap">
          <input
            id="register-password"
            v-model="form.password"
            type="password"
            autocomplete="new-password"
            class="ui-control"
            placeholder="建议使用 8 位以上字母数字组合"
            @input="clearSubmitError"
          >
        </div>
      </div>

      <div class="ui-field">
        <label
          class="ui-field__label"
          for="register-class-name"
        >班级邀请码（可选）</label>
        <div class="ui-control-wrap">
          <input
            id="register-class-name"
            v-model="form.class_name"
            class="ui-control"
            placeholder="输入班级全称以自动加入"
            @input="clearSubmitError"
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
import { RouterLink } from 'vue-router'

import AuthEntryShell from '@/components/auth/AuthEntryShell.vue'
import { useRegisterPage } from '@/features/auth'

const { form, loading, submitError, clearSubmitError, onSubmit } = useRegisterPage()
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
