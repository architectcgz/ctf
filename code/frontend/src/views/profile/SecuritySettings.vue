<script setup lang="ts">
import { KeyRound, Loader2 } from 'lucide-vue-next'

import PageHeader from '@/components/common/PageHeader.vue'
import { useSecuritySettingsPage } from '@/features/profile'

const {
  passwordSaving,
  passwordError,
  passwordForm,
  passwordFieldErrors,
  securityStats,
  passwordTips,
  submitPasswordChange,
} = useSecuritySettingsPage()
</script>

<template>
  <section
    class="workspace-shell journal-shell journal-shell-user journal-hero flex min-h-full flex-1 flex-col"
  >
    <main class="content-pane">
      <div class="security-page flex flex-1 flex-col">
        <PageHeader
          class="security-topbar"
          title="安全设置"
          description="更新账号密码并检查当前安全策略。"
          eyebrow="Security"
        >
          <div class="security-topbar-actions">
            <div class="security-pill">
              <span class="status-dot status-dot-active" />
              密码策略已启用
            </div>
          </div>
        </PageHeader>

        <section
          class="security-summary"
          aria-label="安全概况"
        >
          <div class="security-summary-title">
            <KeyRound class="h-4 w-4" />
            <span>安全概况</span>
          </div>
          <div class="security-summary-grid metric-panel-grid">
            <article
              v-for="stat in securityStats"
              :key="stat.key"
              class="security-summary-item metric-panel-card"
            >
              <div class="security-summary-icon">
                <KeyRound class="h-4 w-4" />
              </div>
              <div>
                <div class="journal-note-label metric-panel-label">
                  {{ stat.label }}
                </div>
                <div
                  class="security-summary-value metric-panel-value"
                  :class="{ 'tech-font': stat.key === 'rotation' }"
                >
                  {{ stat.value }}
                </div>
                <div class="journal-note-helper metric-panel-helper">
                  {{ stat.helper }}
                </div>
              </div>
            </article>
          </div>
        </section>

        <div class="journal-divider security-divider" />

        <div class="security-layout">
          <form
            class="security-section"
            @submit.prevent="submitPasswordChange"
          >
            <div class="security-section-head">
              <div>
                <div class="security-section-kicker">
                  Password
                </div>
                <h2 class="security-section-title">
                  密码修改
                </h2>
              </div>
            </div>

            <div class="space-y-1.5">
              <label class="ui-field__label">当前密码</label>
              <div
                class="ui-control-wrap"
                :class="{ 'is-error': passwordFieldErrors.oldPassword }"
              >
                <input
                  v-model="passwordForm.oldPassword"
                  type="password"
                  autocomplete="current-password"
                  class="ui-control"
                  placeholder="输入当前密码"
                >
              </div>
              <p
                v-if="passwordFieldErrors.oldPassword"
                class="journal-field-error"
              >
                {{ passwordFieldErrors.oldPassword }}
              </p>
            </div>

            <div class="space-y-1.5">
              <label class="ui-field__label">新密码</label>
              <div
                class="ui-control-wrap"
                :class="{ 'is-error': passwordFieldErrors.newPassword }"
              >
                <input
                  v-model="passwordForm.newPassword"
                  type="password"
                  autocomplete="new-password"
                  class="ui-control"
                  placeholder="至少 8 位"
                >
              </div>
              <p
                v-if="passwordFieldErrors.newPassword"
                class="journal-field-error"
              >
                {{ passwordFieldErrors.newPassword }}
              </p>
            </div>

            <div class="space-y-1.5">
              <label class="ui-field__label">确认新密码</label>
              <div
                class="ui-control-wrap"
                :class="{ 'is-error': passwordFieldErrors.confirmPassword }"
              >
                <input
                  v-model="passwordForm.confirmPassword"
                  type="password"
                  autocomplete="new-password"
                  class="ui-control"
                  placeholder="再次输入新密码"
                >
              </div>
              <p
                v-if="passwordFieldErrors.confirmPassword"
                class="journal-field-error"
              >
                {{ passwordFieldErrors.confirmPassword }}
              </p>
            </div>

            <div
              v-if="passwordError"
              class="security-error"
            >
              {{ passwordError }}
            </div>

            <div class="security-actions">
              <button
                type="button"
                class="journal-btn journal-btn--primary"
                :disabled="passwordSaving"
                @click="submitPasswordChange"
              >
                <Loader2
                  v-if="passwordSaving"
                  class="h-4 w-4 animate-spin"
                />
                {{ passwordSaving ? '提交中…' : '更新密码' }}
              </button>
            </div>
          </form>

          <aside class="security-section security-section--aside">
            <div class="security-section-head">
              <div>
                <div class="security-section-kicker">
                  Tips
                </div>
                <h2 class="security-section-title">
                  安全提示
                </h2>
              </div>
            </div>

            <div class="security-side-lead">
              <div class="security-side-status flex items-center gap-2">
                <span class="status-dot status-dot-active" />
                修改后会同步退出其他设备
              </div>
              <p class="security-side-copy mt-3">
                提交后会立即更新当前账号密码，并提示其他设备重新完成认证。
              </p>
            </div>

            <div class="security-tip-list">
              <div
                v-for="tip in passwordTips"
                :key="tip"
                class="security-tip-item"
              >
                <div class="journal-note-label">
                  安全提示
                </div>
                <div class="security-tip-copy mt-2">
                  {{ tip }}
                </div>
              </div>
            </div>
          </aside>
        </div>
      </div>
    </main>
  </section>
</template>

<style scoped>
.journal-shell {
  --journal-shell-font: var(--font-family-sans);
  --journal-shell-accent: var(--color-primary);
  --journal-shell-accent-strong: color-mix(
    in srgb,
    var(--color-primary-hover) 82%,
    var(--journal-ink)
  );
  --journal-shell-surface: color-mix(in srgb, var(--color-bg-surface) 88%, var(--color-bg-base));
  --journal-shell-surface-subtle: color-mix(
    in srgb,
    var(--color-bg-surface) 74%,
    var(--color-bg-base)
  );
  --journal-shell-hero-radial-strength: 8%;
  --journal-shell-hero-radial-size: 18rem;
  --journal-shell-hero-end: color-mix(
    in srgb,
    var(--journal-surface-subtle) 94%,
    var(--color-bg-base)
  );
  --journal-shell-hero-shadow: 0 18px 40px var(--color-shadow-soft);
  --journal-shell-dark-hero-radial-strength: 18%;
  --journal-shell-dark-hero-radial-size: 20rem;
  --journal-shell-dark-hero-top: color-mix(
    in srgb,
    var(--journal-surface) 97%,
    var(--color-bg-base)
  );
  --journal-shell-dark-hero-end: color-mix(
    in srgb,
    var(--journal-surface-subtle) 95%,
    var(--color-bg-base)
  );
  --journal-note-label-size: 0.72rem;
  --journal-note-label-weight: 700;
  --journal-note-label-spacing: 0.16em;
  --journal-note-helper-line-height: 1.45;
  --journal-user-button-height: 2.7rem;
  --journal-user-button-radius: 999px;
  --journal-user-button-padding: 0.62rem 1rem;
  --journal-user-button-size: 0.875rem;
  --journal-user-button-weight: 600;
  --journal-user-button-hover-background: color-mix(
    in srgb,
    var(--journal-accent) 4%,
    var(--journal-surface)
  );
  --journal-user-button-primary-border: color-mix(in srgb, var(--journal-accent) 32%, transparent);
  --journal-user-button-primary-background: color-mix(
    in srgb,
    var(--journal-accent) 12%,
    var(--journal-surface)
  );
  --journal-user-button-primary-color: color-mix(
    in srgb,
    var(--journal-accent) 88%,
    var(--journal-ink)
  );
  --journal-user-tech-font: var(--font-family-mono);
}

.security-page {
  min-height: 100%;
}

.security-topbar-actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: flex-end;
  gap: 0.75rem;
}

.security-divider {
  margin: 0;
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

.security-section-kicker {
  font-size: var(--font-size-0-72);
  font-weight: 700;
  letter-spacing: 0.16em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.security-section-title {
  margin-top: 0.35rem;
  font-size: var(--font-size-1-15);
  font-weight: 700;
  color: var(--journal-ink);
}

.journal-field-error {
  font-size: var(--font-size-0-75);
  color: var(--color-danger);
}

.security-pill {
  display: inline-flex;
  align-items: center;
  gap: 0.45rem;
  border-radius: 999px;
  border: 1px solid color-mix(in srgb, var(--journal-accent) 18%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 8%, transparent);
  padding: 0.48rem 0.9rem;
  font-size: var(--font-size-0-80);
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

.security-side-status {
  font-size: var(--font-size-0-875);
  font-weight: 500;
  color: var(--journal-ink);
}

.security-side-copy,
.security-tip-copy {
  font-size: var(--font-size-0-875);
  line-height: 1.5rem;
}

.security-side-copy {
  color: var(--journal-muted);
}

.security-tip-copy {
  color: var(--journal-ink);
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
  font-size: var(--font-size-0-84);
  color: color-mix(in srgb, var(--color-danger) 88%, var(--journal-ink));
}

.status-dot {
  display: inline-block;
  width: 7px;
  height: 7px;
  border-radius: 50%;
  flex-shrink: 0;
}

.status-dot-active {
  background: var(--color-success);
  box-shadow: 0 0 0 2px color-mix(in srgb, var(--color-success) 20%, transparent);
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
  .content-pane {
    padding-inline: 1rem;
  }

  .security-topbar-actions {
    justify-content: flex-start;
  }
}
</style>
