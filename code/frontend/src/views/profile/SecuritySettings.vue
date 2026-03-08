<script setup lang="ts">
import { reactive, ref } from 'vue'

import { changePassword } from '@/api/auth'
import AppCard from '@/components/common/AppCard.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import SectionCard from '@/components/common/SectionCard.vue'
import { useToast } from '@/composables/useToast'

const toast = useToast()

const passwordSaving = ref(false)
const passwordError = ref<string | null>(null)

const passwordForm = reactive({
  oldPassword: '',
  newPassword: '',
  confirmPassword: '',
})

const passwordFieldErrors = reactive({
  oldPassword: '',
  newPassword: '',
  confirmPassword: '',
})

function resetPasswordErrors(): void {
  passwordError.value = null
  passwordFieldErrors.oldPassword = ''
  passwordFieldErrors.newPassword = ''
  passwordFieldErrors.confirmPassword = ''
}

function validatePasswordForm(): boolean {
  resetPasswordErrors()

  if (!passwordForm.oldPassword.trim()) {
    passwordFieldErrors.oldPassword = '请填写当前密码'
  }

  if (!passwordForm.newPassword.trim()) {
    passwordFieldErrors.newPassword = '请填写新密码'
  } else if (passwordForm.newPassword.trim().length < 8) {
    passwordFieldErrors.newPassword = '新密码至少需要 8 位'
  }

  if (!passwordForm.confirmPassword.trim()) {
    passwordFieldErrors.confirmPassword = '请再次输入新密码'
  } else if (passwordForm.confirmPassword !== passwordForm.newPassword) {
    passwordFieldErrors.confirmPassword = '两次输入的新密码不一致'
  }

  return (
    !passwordFieldErrors.oldPassword &&
    !passwordFieldErrors.newPassword &&
    !passwordFieldErrors.confirmPassword
  )
}

async function submitPasswordChange(): Promise<void> {
  if (!validatePasswordForm()) {
    return
  }

  passwordSaving.value = true
  passwordError.value = null
  try {
    await changePassword({
      old_password: passwordForm.oldPassword,
      new_password: passwordForm.newPassword,
    })
    passwordForm.oldPassword = ''
    passwordForm.newPassword = ''
    passwordForm.confirmPassword = ''
    toast.success('密码修改成功')
  } catch (err) {
    console.error('修改密码失败:', err)
    passwordError.value = err instanceof Error ? err.message : '修改密码失败，请稍后重试'
  } finally {
    passwordSaving.value = false
  }
}
</script>

<template>
  <div class="mx-auto max-w-3xl space-y-6">
    <PageHeader
      eyebrow="Security Settings"
      title="安全设置"
      description="修改登录密码。后续的安全能力也会统一收敛到这里。"
    />

    <SectionCard title="密码修改" subtitle="提交前需要先输入当前密码，新密码长度不少于 8 位。">
      <AppCard variant="action" accent="neutral">
        <div class="space-y-4">
          <div class="grid gap-4 sm:grid-cols-2">
            <label class="space-y-2">
              <span class="text-sm text-[var(--color-text-secondary)]">当前密码</span>
              <input
                v-model="passwordForm.oldPassword"
                type="password"
                autocomplete="current-password"
                class="w-full rounded-xl border border-[var(--color-border-default)] bg-[var(--color-bg-base)] px-4 py-3 text-sm text-[var(--color-text-primary)] outline-none transition focus:border-[var(--color-primary)]"
                placeholder="输入当前密码"
              />
              <p v-if="passwordFieldErrors.oldPassword" class="text-xs text-rose-400">
                {{ passwordFieldErrors.oldPassword }}
              </p>
            </label>

            <label class="space-y-2">
              <span class="text-sm text-[var(--color-text-secondary)]">新密码</span>
              <input
                v-model="passwordForm.newPassword"
                type="password"
                autocomplete="new-password"
                class="w-full rounded-xl border border-[var(--color-border-default)] bg-[var(--color-bg-base)] px-4 py-3 text-sm text-[var(--color-text-primary)] outline-none transition focus:border-[var(--color-primary)]"
                placeholder="输入新密码"
              />
              <p v-if="passwordFieldErrors.newPassword" class="text-xs text-rose-400">
                {{ passwordFieldErrors.newPassword }}
              </p>
            </label>
          </div>

          <label class="space-y-2">
            <span class="text-sm text-[var(--color-text-secondary)]">确认新密码</span>
            <input
              v-model="passwordForm.confirmPassword"
              type="password"
              autocomplete="new-password"
              class="w-full rounded-xl border border-[var(--color-border-default)] bg-[var(--color-bg-base)] px-4 py-3 text-sm text-[var(--color-text-primary)] outline-none transition focus:border-[var(--color-primary)]"
              placeholder="再次输入新密码"
            />
            <p v-if="passwordFieldErrors.confirmPassword" class="text-xs text-rose-400">
              {{ passwordFieldErrors.confirmPassword }}
            </p>
          </label>

          <div
            v-if="passwordError"
            class="rounded-xl border border-red-200 bg-red-50 px-4 py-4 text-sm text-red-600"
          >
            {{ passwordError }}
          </div>

          <div class="flex justify-end">
            <button
              type="button"
              class="inline-flex items-center rounded-xl bg-[var(--color-primary)] px-5 py-3 text-sm font-medium text-white transition hover:bg-[var(--color-primary-hover)] disabled:cursor-not-allowed disabled:opacity-60"
              :disabled="passwordSaving"
              @click="submitPasswordChange"
            >
              {{ passwordSaving ? '正在提交...' : '更新密码' }}
            </button>
          </div>
        </div>
      </AppCard>
    </SectionCard>
  </div>
</template>
