<script setup lang="ts">
import { reactive, ref } from 'vue'
import { KeyRound, Loader2 } from 'lucide-vue-next'

import { changePassword } from '@/api/auth'
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

const securityStats = [
  {
    key: 'policy',
    label: '密码策略',
    value: '已启用',
    helper: '最少 8 位并建议使用字母与数字混合',
  },
  { key: 'rotation', label: '建议轮换', value: '90 天', helper: '定期更新，降低长期暴露风险' },
  { key: 'session', label: '安全通道', value: '在线', helper: '密码修改请求通过受保护会话提交' },
  { key: 'scope', label: '同步范围', value: '全账号', helper: '更新后其他设备需重新登录验证' },
]

const passwordTips = [
  '不要使用生日、姓名等易猜测信息',
  '避免在多个平台复用同一密码',
  '修改后其他设备需重新登录',
]

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
  <section
    class="journal-shell journal-hero flex min-h-full flex-1 flex-col rounded-[30px] border px-6 py-6 md:px-8"
  >
    <header class="security-header">
      <div class="security-header__intro">
        <div class="journal-eyebrow">Security</div>
        <h1 class="security-page-title">安全设置</h1>
        <p class="security-page-copy">更新账号密码并检查当前安全策略。</p>

        <div class="security-header__actions">
          <div class="security-pill">
            <span class="status-dot status-dot-active" />
            密码策略已启用
          </div>
        </div>
      </div>

      <div class="security-summary-grid">
        <article v-for="stat in securityStats" :key="stat.key" class="security-summary-item">
          <div class="security-summary-icon">
            <KeyRound class="h-4 w-4" />
          </div>
          <div>
            <div class="journal-note-label">{{ stat.label }}</div>
            <div class="security-summary-value" :class="{ 'tech-font': stat.key === 'rotation' }">
              {{ stat.value }}
            </div>
            <div class="journal-note-helper">{{ stat.helper }}</div>
          </div>
        </article>
      </div>
    </header>

    <div class="journal-divider security-divider" />

    <div class="security-layout">
      <form class="security-section" @submit.prevent="submitPasswordChange">
        <div class="security-section-head">
          <div>
            <div class="journal-eyebrow journal-eyebrow-soft">Password</div>
            <h2 class="security-section-title">密码修改</h2>
          </div>
        </div>

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
          <p v-if="passwordFieldErrors.oldPassword" class="journal-field-error">
            {{ passwordFieldErrors.oldPassword }}
          </p>
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
          <p v-if="passwordFieldErrors.newPassword" class="journal-field-error">
            {{ passwordFieldErrors.newPassword }}
          </p>
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
          <p v-if="passwordFieldErrors.confirmPassword" class="journal-field-error">
            {{ passwordFieldErrors.confirmPassword }}
          </p>
        </div>

        <div v-if="passwordError" class="security-error">
          {{ passwordError }}
        </div>

        <div class="security-actions">
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
      </form>

      <aside class="security-section security-section--aside">
        <div class="security-section-head">
          <div>
            <div class="journal-eyebrow journal-eyebrow-soft">Tips</div>
            <h2 class="security-section-title">安全提示</h2>
          </div>
        </div>

        <div class="security-side-lead">
          <div class="flex items-center gap-2 text-sm font-medium text-[var(--journal-ink)]">
            <span class="status-dot status-dot-active" />
            修改后会同步退出其他设备
          </div>
          <p class="mt-3 text-sm leading-6 text-[var(--journal-muted)]">
            提交后会立即更新当前账号密码，并提示其他设备重新完成认证。
          </p>
        </div>

        <div class="security-tip-list">
          <div v-for="tip in passwordTips" :key="tip" class="security-tip-item">
            <div class="journal-note-label">安全提示</div>
            <div class="mt-2 text-sm leading-6 text-[var(--journal-ink)]">{{ tip }}</div>
          </div>
        </div>
      </aside>
    </div>
  </section>
</template>

<style scoped>
.journal-shell {
  --journal-ink: var(--color-text-primary);
  --journal-muted: var(--color-text-secondary);
  --journal-accent: var(--color-primary);
  --journal-accent-strong: color-mix(in srgb, var(--color-primary-hover) 82%, var(--journal-ink));
  --journal-border: color-mix(in srgb, var(--color-border-default) 82%, transparent);
  --journal-surface: color-mix(in srgb, var(--color-bg-surface) 88%, var(--color-bg-base));
  --journal-surface-subtle: color-mix(in srgb, var(--color-bg-surface) 74%, var(--color-bg-base));
  font-family: 'Inter', 'Noto Sans SC', system-ui, sans-serif;
}

.journal-hero {
  border-color: var(--journal-border);
  background:
    radial-gradient(circle at top right, color-mix(in srgb, var(--journal-accent) 8%, transparent), transparent 18rem),
    linear-gradient(180deg, color-mix(in srgb, var(--journal-surface) 96%, var(--color-bg-base)), color-mix(in srgb, var(--journal-surface-subtle) 94%, var(--color-bg-base)));
  box-shadow: 0 18px 40px var(--color-shadow-soft);
}

.journal-eyebrow {
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  color: var(--journal-accent);
}

.journal-eyebrow-soft {
  color: var(--journal-muted);
}

.journal-note-label {
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.16em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.journal-note-helper {
  margin-top: 0.55rem;
  font-size: 0.78rem;
  line-height: 1.45;
  color: var(--journal-muted);
}

.security-header {
  display: grid;
  gap: 1rem;
}

.security-page-title {
  margin-top: 0.85rem;
  font-size: clamp(2rem, 2.7vw, 2.8rem);
  font-weight: 700;
  line-height: 1.08;
  color: var(--journal-ink);
}

.security-page-copy {
  margin-top: 0.8rem;
  max-width: 46rem;
  font-size: 0.94rem;
  line-height: 1.75;
  color: var(--journal-muted);
}

.security-header__actions {
  margin-top: 1rem;
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
}

.security-summary-grid {
  display: grid;
  gap: 0.9rem;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.security-summary-item {
  display: flex;
  gap: 0.75rem;
  border-top: 1px solid color-mix(in srgb, var(--journal-border) 86%, transparent);
  padding-top: 0.85rem;
}

.security-summary-icon {
  display: inline-flex;
  margin-top: 0.1rem;
  color: var(--journal-accent);
}

.security-summary-value {
  margin-top: 0.4rem;
  font-size: 1rem;
  font-weight: 600;
  color: var(--journal-ink);
}

.security-divider {
  margin: 1.2rem 0 0;
}

.security-layout {
  display: grid;
  gap: 1.5rem;
  padding-top: 1.25rem;
  grid-template-columns: minmax(0, 1.08fr) minmax(280px, 0.92fr);
}

.security-section + .security-section {
  border-left: 1px solid color-mix(in srgb, var(--journal-border) 86%, transparent);
  padding-left: 1.5rem;
}

.security-section-head {
  margin-bottom: 1rem;
}

.security-section-title {
  margin-top: 0.35rem;
  font-size: 1.15rem;
  font-weight: 700;
  color: var(--journal-ink);
}

.journal-label {
  display: block;
  font-size: 0.8125rem;
  font-weight: 500;
  color: var(--journal-ink);
}

.journal-input {
  width: 100%;
  border-radius: 1rem;
  border: 1px solid var(--journal-border);
  background: var(--journal-surface);
  padding: 0.7rem 0.95rem;
  font-size: 0.875rem;
  color: var(--journal-ink);
  outline: none;
  transition:
    border-color 0.2s,
    box-shadow 0.2s;
}

.journal-input:focus {
  border-color: color-mix(in srgb, var(--journal-accent) 50%, transparent);
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--journal-accent) 12%, transparent);
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
  justify-content: center;
  gap: 0.375rem;
  min-height: 2.7rem;
  border-radius: 999px;
  border: 1px solid color-mix(in srgb, var(--journal-border) 84%, transparent);
  background: color-mix(in srgb, var(--journal-surface) 95%, var(--color-bg-base));
  padding: 0.62rem 1rem;
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--journal-ink);
  transition:
    border-color 0.2s,
    color 0.2s,
    background 0.2s;
  cursor: pointer;
}

.journal-btn:hover {
  border-color: color-mix(in srgb, var(--journal-accent) 42%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 4%, var(--journal-surface));
}

.journal-btn:disabled {
  cursor: not-allowed;
  opacity: 0.5;
}

.journal-btn--primary {
  border-color: color-mix(in srgb, var(--journal-accent) 32%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 12%, var(--journal-surface));
  color: color-mix(in srgb, var(--journal-accent) 88%, var(--journal-ink));
}

.journal-btn--primary:hover:not(:disabled) {
  background: color-mix(in srgb, var(--journal-accent) 16%, transparent);
}

.security-pill {
  display: inline-flex;
  align-items: center;
  gap: 0.45rem;
  border-radius: 999px;
  border: 1px solid color-mix(in srgb, var(--journal-accent) 18%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 8%, transparent);
  padding: 0.48rem 0.9rem;
  font-size: 0.8rem;
  font-weight: 600;
  color: color-mix(in srgb, var(--journal-accent) 84%, var(--journal-ink));
}

.security-side-lead {
  border-top: 1px solid color-mix(in srgb, var(--journal-border) 86%, transparent);
  padding-top: 1rem;
}

.security-tip-list {
  margin-top: 1rem;
  border-top: 1px solid color-mix(in srgb, var(--journal-border) 86%, transparent);
}

.security-tip-item {
  padding: 1rem 0;
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 86%, transparent);
}

.security-tip-item:last-child {
  border-bottom: 0;
  padding-bottom: 0;
}

.security-actions {
  display: flex;
  justify-content: flex-end;
  padding-top: 0.5rem;
}

.security-error {
  border-inline-start: 2px solid color-mix(in srgb, var(--color-danger) 60%, transparent);
  background: color-mix(in srgb, var(--color-danger) 8%, transparent);
  padding: 0.75rem 0.9rem;
  font-size: 0.84rem;
  color: color-mix(in srgb, var(--color-danger) 88%, var(--journal-ink));
}

.tech-font {
  font-family: 'JetBrains Mono', 'Fira Code', 'SFMono-Regular', monospace;
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
  0%,
  100% {
    opacity: 1;
  }
  50% {
    opacity: 0.5;
  }
}

:global([data-theme='dark']) .journal-hero {
  background:
    radial-gradient(circle at top right, color-mix(in srgb, var(--journal-accent) 18%, transparent), transparent 20rem),
    linear-gradient(180deg, rgba(15, 23, 42, 0.95), rgba(2, 6, 23, 0.98));
}

@media (max-width: 1024px) {
  .security-layout {
    grid-template-columns: minmax(0, 1fr);
  }

  .security-section + .security-section {
    border-left: 0;
    border-top: 1px solid color-mix(in srgb, var(--journal-border) 86%, transparent);
    padding-left: 0;
    padding-top: 1.25rem;
  }
}

@media (max-width: 720px) {
  .journal-shell {
    padding-inline: 1rem;
  }

  .security-summary-grid {
    grid-template-columns: minmax(0, 1fr);
  }
}
</style>
