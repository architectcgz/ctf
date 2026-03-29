<script setup lang="ts">
import { reactive, ref } from 'vue'
import { KeyRound, Loader2 } from 'lucide-vue-next'

import { changePassword } from '@/api/auth'
import PageHeader from '@/components/common/PageHeader.vue'
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
  <div class="journal-shell space-y-6">
    <PageHeader
      eyebrow="Security Settings"
      title="安全设置"
      description="修改登录密码。后续的安全能力也会统一收敛到这里。"
    />

    <!-- 修改密码 -->
    <section class="journal-hero rounded-[30px] border px-6 py-6 md:px-8">
      <div class="grid gap-6 xl:grid-cols-[1.1fr_0.9fr]">
        <!-- 左：表单 -->
        <div>
          <div class="journal-eyebrow">Password</div>
          <h2 class="mt-3 text-3xl font-semibold tracking-tight text-[var(--journal-ink)] md:text-[2.45rem]">
            修改登录密码
          </h2>
          <div class="mt-4 flex items-center gap-2 text-xs text-[var(--journal-muted)]">
            <span class="status-dot status-dot-active" />
            <span class="tech-font">session://secure-channel-active</span>
          </div>
          <p class="mt-3 max-w-xl text-sm leading-7 text-[var(--journal-muted)]">
            建议使用至少 8 位、包含字母和数字的强密码，并定期更换以保护账号安全。
          </p>

          <div class="mt-6 space-y-4">
            <div class="space-y-1.5">
              <label class="journal-label">当前密码</label>
              <input
                v-model="passwordForm.oldPassword"
                type="password"
                autocomplete="current-password"
                class="journal-input"
                :class="{ 'journal-input--error': passwordFieldErrors.oldPassword }"
                placeholder="输入当前密码"
              />
              <p v-if="passwordFieldErrors.oldPassword" class="journal-field-error">{{ passwordFieldErrors.oldPassword }}</p>
            </div>

            <div class="space-y-1.5">
              <label class="journal-label">新密码</label>
              <input
                v-model="passwordForm.newPassword"
                type="password"
                autocomplete="new-password"
                class="journal-input"
                :class="{ 'journal-input--error': passwordFieldErrors.newPassword }"
                placeholder="至少 8 位"
              />
              <p v-if="passwordFieldErrors.newPassword" class="journal-field-error">{{ passwordFieldErrors.newPassword }}</p>
            </div>

            <div class="space-y-1.5">
              <label class="journal-label">确认新密码</label>
              <input
                v-model="passwordForm.confirmPassword"
                type="password"
                autocomplete="new-password"
                class="journal-input"
                :class="{ 'journal-input--error': passwordFieldErrors.confirmPassword }"
                placeholder="再次输入新密码"
              />
              <p v-if="passwordFieldErrors.confirmPassword" class="journal-field-error">{{ passwordFieldErrors.confirmPassword }}</p>
            </div>

            <div
              v-if="passwordError"
              class="rounded-[16px] border border-[var(--color-danger)]/20 bg-[var(--color-danger)]/8 px-4 py-3 text-sm text-[var(--color-danger)]"
            >
              {{ passwordError }}
            </div>

            <div class="flex justify-end pt-2">
              <button
                type="button"
                class="journal-btn journal-btn--primary"
                :disabled="passwordSaving"
                @click="submitPasswordChange"
              >
                <Loader2 v-if="passwordSaving" class="h-4 w-4 animate-spin" />
                {{ passwordSaving ? '提交中…' : '更新密码' }}
              </button>
            </div>
          </div>
        </div>

        <!-- 右：安全提示侧边栏 -->
        <article class="journal-brief rounded-[24px] border px-5 py-5">
          <div class="flex items-center gap-3 text-sm font-medium text-[var(--journal-ink)]">
            <KeyRound class="h-5 w-5 text-[var(--journal-accent)]" />
            密码安全要求
          </div>

          <!-- 状态行 -->
          <div class="mt-4 rounded-[16px] border border-[var(--journal-border)] bg-[var(--journal-surface)] px-4 py-3">
            <div class="flex items-center gap-2 text-sm text-[var(--journal-muted)]">
              <span class="status-dot status-dot-active" />
              密码策略已激活
            </div>
            <div class="mt-1 tech-font text-xs text-[var(--journal-muted)]">policy://min-8-alphanumeric</div>
          </div>

          <div class="mt-4 space-y-3">
            <div class="journal-note">
              <div class="journal-note-label">最小长度</div>
              <div class="journal-note-value tech-font">8 位</div>
            </div>
            <div class="journal-note">
              <div class="journal-note-label">建议强度</div>
              <div class="journal-note-value">字母 + 数字混合</div>
            </div>
            <div class="journal-note">
              <div class="journal-note-label">更新频率</div>
              <div class="journal-note-value">建议每 90 天</div>
            </div>
          </div>
          <div class="mt-5 space-y-2">
            <div
              v-for="tip in ['不要使用生日、姓名等易猜测信息', '避免在多个平台复用同一密码', '修改后其他设备需重新登录']" :key="tip"
              class="flex items-start gap-2 text-xs leading-5 text-[var(--journal-muted)]"
            >
              <span class="mt-1.5 h-1.5 w-1.5 shrink-0 rounded-full" style="background: var(--journal-accent)" />
              {{ tip }}
            </div>
          </div>
        </article>
      </div>
    </section>
  </div>
</template>

<style scoped>
.journal-shell {
  --journal-ink: #0f172a;
  --journal-muted: #64748b;
  --journal-accent: #4f46e5;
  --journal-accent-strong: #4338ca;
  --journal-border: rgba(226, 232, 240, 0.8);
  --journal-surface: rgba(248, 250, 252, 0.9);
  --journal-surface-subtle: rgba(241, 245, 249, 0.7);
  font-family: "Inter", "Noto Sans SC", system-ui, sans-serif;
}

.journal-hero {
  border-color: var(--journal-border);
  background:
    radial-gradient(circle at top right, rgba(79, 70, 229, 0.06), transparent 20rem),
    linear-gradient(180deg, rgba(248, 250, 252, 0.98), rgba(241, 245, 249, 0.95));
}

.journal-eyebrow {
  font-size: 0.7rem;
  font-weight: 700;
  letter-spacing: 0.2em;
  text-transform: uppercase;
  color: var(--journal-accent);
}

.journal-label {
  display: block;
  font-size: 0.8125rem;
  font-weight: 500;
  color: var(--journal-ink);
}

.journal-input {
  width: 100%;
  border-radius: 14px;
  border: 1px solid var(--journal-border);
  background: var(--journal-surface);
  padding: 0.625rem 0.875rem;
  font-size: 0.875rem;
  color: var(--journal-ink);
  outline: none;
  transition: border-color 0.2s;
}

.journal-input:focus {
  border-color: color-mix(in srgb, var(--journal-accent) 50%, transparent);
}

.journal-input--error {
  border-color: color-mix(in srgb, var(--color-danger) 50%, transparent);
}

.journal-field-error {
  font-size: 0.75rem;
  color: var(--color-danger);
}

.journal-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  border-radius: 12px;
  border: 1px solid var(--color-border-default);
  padding: 0.5rem 1.25rem;
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--color-text-primary);
  background: transparent;
  transition: border-color 0.2s, color 0.2s;
  cursor: pointer;
}

.journal-btn:hover {
  border-color: var(--journal-accent);
  color: var(--journal-accent);
}

.journal-btn:disabled {
  cursor: not-allowed;
  opacity: 0.5;
}

.journal-btn--primary {
  border-color: color-mix(in srgb, var(--journal-accent) 50%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 8%, transparent);
  color: var(--journal-accent);
}

.journal-btn--primary:hover:not(:disabled) {
  background: color-mix(in srgb, var(--journal-accent) 14%, transparent);
}

.journal-brief {
  border-color: var(--journal-border);
  background: var(--journal-surface-subtle);
}

.journal-note {
  border-radius: 14px;
  border: 1px solid var(--journal-border);
  background: var(--journal-surface);
  padding: 0.625rem 0.875rem;
}

.journal-note-label {
  font-size: 0.7rem;
  font-weight: 600;
  letter-spacing: 0.15em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.journal-note-value {
  margin-top: 0.375rem;
  font-size: 1rem;
  font-weight: 600;
  color: var(--journal-ink);
}

.tech-font {
  font-family: "JetBrains Mono", "Fira Code", "SFMono-Regular", monospace;
}

.journal-hero {
  box-shadow: 0 18px 40px rgba(15, 23, 42, 0.06);
}

.journal-brief {
  box-shadow: 0 4px 12px rgba(15, 23, 42, 0.04);
}

.status-dot {
  display: inline-block;
  width: 7px;
  height: 7px;
  border-radius: 50%;
  flex-shrink: 0;
}

.status-dot-active {
  background: #10b981;
  box-shadow: 0 0 0 2px rgba(16, 185, 129, 0.2);
  animation: pulse-dot 2s infinite;
}

@keyframes pulse-dot {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

:global([data-theme='dark']) .journal-shell {
  --journal-ink: #f1f5f9;
  --journal-muted: #94a3b8;
  --journal-border: rgba(51, 65, 85, 0.72);
  --journal-surface: rgba(15, 23, 42, 0.85);
  --journal-surface-subtle: rgba(30, 41, 59, 0.6);
}

:global([data-theme='dark']) .journal-hero {
  background:
    radial-gradient(circle at top right, rgba(79, 70, 229, 0.18), transparent 20rem),
    linear-gradient(180deg, rgba(15, 23, 42, 0.95), rgba(2, 6, 23, 0.98));
}
</style>
