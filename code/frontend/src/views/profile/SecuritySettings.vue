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
  <section class="journal-shell space-y-6 journal-hero flex min-h-full flex-1 flex-col rounded-[30px] border px-6 py-6 md:px-8">
      <div class="grid gap-6 xl:grid-cols-[1.06fr_0.94fr]">
        <div>
          <div class="journal-eyebrow">Security Console</div>
          <h2
            class="mt-3 text-3xl font-semibold tracking-tight text-[var(--journal-ink)] md:text-[2.45rem]"
          >
            安全设置
          </h2>
          <p class="mt-3 max-w-2xl text-sm leading-7 text-[var(--journal-muted)]">
            在这里修改密码和查看安全提示。
          </p>

          <div class="mt-6 flex flex-wrap gap-3">
            <div class="security-pill">
              <span class="status-dot status-dot-active" />
              密码策略已启用
            </div>
          </div>
        </div>

        <article class="journal-brief rounded-[24px] border px-5 py-5">
          <div class="flex items-center gap-3 text-sm font-medium text-[var(--journal-ink)]">
            <KeyRound class="h-5 w-5 text-[var(--journal-accent)]" />
            账户安全概况
          </div>
          <div class="mt-5 grid gap-3 sm:grid-cols-2">
            <div v-for="stat in securityStats" :key="stat.key" class="journal-note">
              <div class="journal-note-label">{{ stat.label }}</div>
              <div class="journal-note-value" :class="{ 'tech-font': stat.key === 'rotation' }">
                {{ stat.value }}
              </div>
              <div class="journal-note-helper">{{ stat.helper }}</div>
            </div>
          </div>
        </article>
      </div>
      <div class="security-panel mt-6 px-1 pt-5 md:px-2 md:pt-6">
        <div class="grid gap-6 xl:grid-cols-[minmax(0,1.08fr)_minmax(280px,0.92fr)]">
          <form class="space-y-4" @submit.prevent="submitPasswordChange">
            <div class="security-section-head">
              <div class="journal-eyebrow">Password</div>
              <h3 class="mt-3 text-xl font-semibold text-[var(--journal-ink)]">密码修改</h3>
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
          </form>

          <aside class="security-side">
            <div class="security-section-head">
              <div class="journal-eyebrow">Security Tips</div>
              <h3 class="mt-3 text-xl font-semibold text-[var(--journal-ink)]">安全提示</h3>
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
      </div>
    </section>
</template>

<style scoped>
.journal-shell {
  --journal-ink: var(--color-text-primary);
  --journal-muted: var(--color-text-secondary);
  --journal-accent: #4f46e5;
  --journal-accent-strong: #4338ca;
  --journal-border: color-mix(in srgb, var(--color-border-default) 82%, transparent);
  --journal-surface: color-mix(in srgb, var(--color-bg-surface) 88%, var(--color-bg-base));
  --journal-surface-subtle: color-mix(in srgb, var(--color-bg-surface) 74%, var(--color-bg-base));
  font-family: 'Inter', 'Noto Sans SC', system-ui, sans-serif;
}

.journal-hero {
  border-color: var(--journal-border);
  background:
    radial-gradient(circle at top right, rgba(37, 99, 235, 0.08), transparent 18rem),
    linear-gradient(180deg, color-mix(in srgb, var(--journal-surface, var(--color-bg-surface)) 96%, var(--color-bg-base)), color-mix(in srgb, var(--journal-surface-subtle, var(--color-bg-elevated)) 94%, var(--color-bg-base)));
  border-radius: 16px !important;
  overflow: hidden;
}

.journal-eyebrow {
  display: inline-flex;
  align-items: center;
  border-radius: 999px;
  border: 1px solid rgba(99, 102, 241, 0.22);
  background: rgba(99, 102, 241, 0.07);
  padding: 0.2rem 0.75rem;
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.08em;
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
  box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.12);
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
  border-radius: 0.9rem;
  border: 1px solid var(--journal-border);
  background: var(--journal-surface);
  padding: 0.55rem 1.15rem;
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
  border-color: var(--journal-accent);
  background: color-mix(in srgb, var(--journal-accent) 4%, var(--journal-surface));
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
  border-radius: 16px !important;
  overflow: hidden;
  box-shadow: 0 10px 24px rgba(15, 23, 42, 0.05);
}

.journal-note {
  border-radius: 18px;
  border: 1px solid var(--journal-border);
  background: var(--journal-surface);
  padding: 0.9rem 0.95rem;
}

.journal-note-label {
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.16em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.journal-note-value {
  margin-top: 0.55rem;
  font-size: 1.05rem;
  font-weight: 600;
  color: var(--journal-ink);
}

.journal-note-helper {
  margin-top: 0.55rem;
  font-size: 0.78rem;
  line-height: 1.45;
  color: var(--journal-muted);
}

.tech-font {
  font-family: 'JetBrains Mono', 'Fira Code', 'SFMono-Regular', monospace;
}

.security-pill {
  display: inline-flex;
  align-items: center;
  gap: 0.45rem;
  border-radius: 999px;
  border: 1px solid rgba(99, 102, 241, 0.16);
  background: rgba(99, 102, 241, 0.06);
  padding: 0.48rem 0.9rem;
  font-size: 0.8rem;
  font-weight: 600;
  color: color-mix(in srgb, var(--journal-accent) 84%, #312e81);
}

.security-pill--muted {
  background: color-mix(in srgb, var(--journal-surface, var(--color-bg-surface)) 92%, var(--color-bg-base));
  color: var(--journal-muted);
}

.security-panel {
  border-top: 1px dashed color-mix(in srgb, var(--journal-border, var(--color-border-default)) 88%, transparent);
}

.security-section-head {
  margin-bottom: 1rem;
}

.security-side {
  border-top: 1px dashed color-mix(in srgb, var(--journal-border, var(--color-border-default)) 88%, transparent);
  padding-top: 1rem;
}

.security-side-lead {
  padding-bottom: 1rem;
}

.security-tip-list {
  border-top: 1px dashed color-mix(in srgb, var(--journal-border, var(--color-border-default)) 88%, transparent);
}

.security-tip-item {
  padding: 1rem 0;
  border-bottom: 1px dashed color-mix(in srgb, var(--journal-border, var(--color-border-default)) 88%, transparent);
}

.security-tip-item:last-child {
  border-bottom: 0;
  padding-bottom: 0;
}

@media (min-width: 1280px) {
  .security-side {
    border-top: 0;
    border-left: 1px dashed color-mix(in srgb, var(--journal-border, var(--color-border-default)) 88%, transparent);
    padding-top: 0;
    padding-left: 1.5rem;
  }
}

.journal-hero {
  box-shadow: 0 18px 40px var(--color-shadow-soft);
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

:global([data-theme='dark']) .journal-shell {
  --journal-ink: var(--color-text-primary);
  --journal-muted: var(--color-text-secondary);
  --journal-border: color-mix(in srgb, var(--color-border-default) 82%, transparent);
  --journal-surface: color-mix(in srgb, var(--color-bg-surface) 88%, var(--color-bg-base));
  --journal-surface-subtle: color-mix(in srgb, var(--color-bg-surface) 74%, var(--color-bg-base));
}

:global([data-theme='dark']) .journal-hero {
  background:
    radial-gradient(circle at top right, rgba(79, 70, 229, 0.18), transparent 20rem),
    linear-gradient(180deg, rgba(15, 23, 42, 0.95), rgba(2, 6, 23, 0.98));
}
</style>
