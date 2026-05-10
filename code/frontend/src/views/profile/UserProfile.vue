<script setup lang="ts">
import { Activity, FileDown, Loader2, RefreshCw, ShieldCheck, UserCircle2 } from 'lucide-vue-next'
import AppEmpty from '@/components/common/AppEmpty.vue'
import { useUserProfilePage } from '@/features/profile'
import { formatDate } from '@/utils/format'

const {
  loading,
  error,
  profile,
  exportLoading,
  exportError,
  reportFormat,
  latestReport,
  latestReportFormat,
  latestReportCreatedAt,
  canManagePersonalReport,
  currentProfile,
  pageCopy,
  profileFields,
  reportTaskMeta,
  profileSummaryItems,
  loadProfile,
  createReport,
  handleDownload,
} = useUserProfilePage()
</script>

<template>
  <section
    class="workspace-shell journal-shell journal-shell-user journal-hero flex min-h-full flex-1 flex-col"
  >
    <main class="content-pane">
      <div
        v-if="error"
        class="profile-inline-notice"
      >
        {{ error }}
      </div>

      <div
        v-if="loading"
        class="profile-loading"
      >
        <div class="h-12 animate-pulse rounded-2xl bg-[var(--journal-surface)]/90" />
        <div class="grid gap-4 md:grid-cols-2">
          <div class="h-24 animate-pulse rounded-2xl bg-[var(--journal-surface)]" />
          <div class="h-24 animate-pulse rounded-2xl bg-[var(--journal-surface)]" />
        </div>
        <div class="grid gap-6 xl:grid-cols-[minmax(0,1fr)_minmax(320px,0.98fr)]">
          <div class="profile-loading-card h-72 animate-pulse bg-[var(--journal-surface)]" />
          <div class="profile-loading-card h-72 animate-pulse bg-[var(--journal-surface)]" />
        </div>
      </div>

      <div
        v-else
        class="profile-page flex flex-1 flex-col"
      >
        <header class="workspace-page-header profile-topbar">
          <div class="profile-heading">
            <div class="workspace-overline">Profile</div>
            <h1 class="workspace-page-title profile-title">
              个人资料
            </h1>
            <p class="workspace-page-copy profile-subtitle">
              {{ pageCopy }}
            </p>
          </div>

          <div class="profile-topbar-meta">
            <div
              class="profile-head-stats"
              aria-label="账号状态"
            >
              <div class="profile-head-stat">
                <span class="profile-head-stat__label">账号状态</span>
                <strong class="profile-head-stat__value">
                  <span class="status-dot status-dot-ready" />
                  正常
                </strong>
              </div>
              <div class="profile-head-stat">
                <span class="profile-head-stat__label">
                  {{ canManagePersonalReport ? '报告状态' : '账号类型' }}
                </span>
                <strong class="profile-head-stat__value">
                  {{ canManagePersonalReport ? reportTaskMeta.label : '管理账号' }}
                </strong>
              </div>
            </div>

            <div class="profile-actions">
              <button
                type="button"
                class="journal-btn"
                @click="loadProfile"
              >
                <RefreshCw class="h-4 w-4" />
                刷新
              </button>
            </div>
          </div>
        </header>

        <section
          class="profile-summary metric-panel-default-surface"
          aria-label="账号概况"
        >
          <div class="profile-summary-title">
            <ShieldCheck class="h-4 w-4" />
            <span>账号概况</span>
          </div>
          <div class="profile-summary-grid metric-panel-grid">
            <article
              v-for="item in profileSummaryItems"
              :key="item.key"
              class="profile-summary-item progress-card metric-panel-card"
            >
              <div class="journal-note-label progress-card-label metric-panel-label">
                <span>{{ item.label }}</span>
                <component
                  :is="item.icon"
                  class="h-4 w-4"
                />
              </div>
              <div
                class="profile-summary-value progress-card-value metric-panel-value"
                :class="{ 'tech-font': item.techFont }"
              >
                {{ item.value }}
              </div>
              <div class="journal-note-helper progress-card-hint metric-panel-helper">
                {{ item.helper }}
              </div>
            </article>
          </div>
        </section>

        <div class="journal-divider profile-divider" />

        <div
          class="profile-layout"
          :class="{ 'profile-layout--single': !canManagePersonalReport }"
        >
          <section class="profile-section">
            <div class="profile-section-head">
              <div>
                <div class="profile-section-kicker">Account</div>
                <h2 class="profile-section-title">
                  <UserCircle2 class="profile-accent-icon h-5 w-5" />
                  账号信息
                </h2>
              </div>
            </div>

            <div
              v-if="profile"
              class="profile-field-list"
            >
              <article
                v-for="item in profileFields"
                :key="item.label"
                class="profile-field-item"
              >
                <div class="journal-note-label">
                  {{ item.label }}
                </div>
                <div class="profile-field-value tech-font">
                  {{ item.value }}
                </div>
              </article>
            </div>

            <AppEmpty
              v-else
              title="暂无用户信息"
              description="当前没有可展示的用户信息。"
              icon="UsersRound"
            />
          </section>

          <section
            v-if="canManagePersonalReport"
            class="profile-section profile-section--report"
          >
            <div class="profile-section-head">
              <div>
                <div class="profile-section-kicker">Report</div>
                <h2 class="profile-section-title">
                  个人报告
                </h2>
              </div>
              <span
                class="journal-chip"
                :class="reportTaskMeta.chipClass"
              >
                {{ reportTaskMeta.label }}
              </span>
            </div>

            <div class="profile-status">
              <div class="profile-status__row">
                <div class="profile-status__label">
                  <Activity class="profile-accent-icon h-4 w-4" />
                  当前状态
                </div>
                <div class="profile-status__value">
                  <span
                    class="status-dot"
                    :class="{
                      'status-dot-ready': reportTaskMeta.status === 'ready',
                      'status-dot-warning': reportTaskMeta.status === 'processing',
                      'status-dot-idle': reportTaskMeta.status === 'idle',
                      'status-dot-danger': reportTaskMeta.status === 'failed',
                    }"
                  />
                  {{ reportTaskMeta.label }}
                </div>
              </div>
              <div
                v-if="latestReport"
                class="profile-status__meta"
              >
                报告编号：{{ latestReport.report_id }}
              </div>
            </div>

            <fieldset class="profile-fieldset">
              <legend class="profile-fieldset__legend">
                导出格式
              </legend>
              <div class="profile-format-grid">
                <label
                  class="journal-format-option"
                  :class="{ 'journal-format-option--active': reportFormat === 'pdf' }"
                >
                  <input
                    v-model="reportFormat"
                    type="radio"
                    value="pdf"
                    class="sr-only"
                  >
                  <div class="profile-format-title">PDF 报告</div>
                  <div class="profile-format-copy mt-1">适合阅读和保存</div>
                </label>
                <label
                  class="journal-format-option"
                  :class="{ 'journal-format-option--active': reportFormat === 'excel' }"
                >
                  <input
                    v-model="reportFormat"
                    type="radio"
                    value="excel"
                    class="sr-only"
                  >
                  <div class="profile-format-title">Excel 报告</div>
                  <div class="profile-format-copy mt-1">适合筛选和整理数据</div>
                </label>
              </div>
            </fieldset>

            <button
              type="button"
              class="journal-btn journal-btn--primary mt-4 w-full justify-center"
              :disabled="exportLoading"
              @click="createReport"
            >
              <Loader2
                v-if="exportLoading"
                class="h-4 w-4 animate-spin"
              />
              {{ exportLoading ? '创建中…' : '生成个人报告' }}
            </button>

            <p
              v-if="exportError"
              class="profile-error-copy"
            >
              {{ exportError }}
            </p>

            <template v-if="latestReport">
              <div class="profile-report-meta">
                <div class="profile-report-meta__item">
                  <div class="journal-note-label">
                    格式
                  </div>
                  <div class="profile-report-meta__value tech-font">
                    {{ latestReportFormat.toUpperCase() }}
                  </div>
                </div>
                <div class="profile-report-meta__item">
                  <div class="journal-note-label">
                    创建时间
                  </div>
                  <div class="profile-report-meta__value">
                    {{ latestReportCreatedAt ? formatDate(latestReportCreatedAt) : '—' }}
                  </div>
                </div>
                <div class="profile-report-meta__item">
                  <div class="journal-note-label">
                    有效期
                  </div>
                  <div class="profile-report-meta__value">
                    {{
                      latestReport.expires_at ? formatDate(latestReport.expires_at) : '待完成后返回'
                    }}
                  </div>
                </div>
              </div>

              <p
                v-if="latestReport.error_message"
                class="profile-error-copy"
              >
                {{ latestReport.error_message }}
              </p>

              <button
                type="button"
                class="journal-btn journal-btn--download mt-4 w-full justify-center"
                :disabled="latestReport.status !== 'ready'"
                @click="handleDownload"
              >
                <FileDown class="h-4 w-4" />
                下载最近报告
              </button>
            </template>
          </section>
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
  --journal-user-button-height: 2.7rem;
  --journal-user-button-radius: 999px;
  --journal-user-button-padding: 0.62rem 1rem;
  --journal-user-button-size: 0.875rem;
  --journal-user-button-weight: 600;
  --journal-user-button-hover-color: var(--journal-accent);
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

.profile-loading {
  display: grid;
  gap: 1rem;
}

.profile-loading-card {
  border-radius: 1.5rem;
}

.profile-page {
  min-height: 100%;
}

.profile-subtitle {
  max-width: 720px;
}

.profile-topbar-meta {
  display: grid;
  justify-items: end;
  gap: 0.75rem;
}

.profile-head-stats {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 0.75rem;
}

.profile-head-stat {
  display: inline-flex;
  align-items: center;
  gap: 0.75rem;
  min-height: 2.75rem;
  padding: 0 0.875rem;
  border: 1px solid color-mix(in srgb, var(--journal-border) 82%, transparent);
  border-radius: 0.875rem;
  background: color-mix(in srgb, var(--journal-surface) 92%, var(--color-bg-base));
}

.profile-head-stat__label {
  font-size: var(--font-size-13);
  font-weight: 600;
  color: var(--journal-muted);
}

.profile-head-stat__value {
  display: inline-flex;
  align-items: center;
  gap: 0.45rem;
  font-size: var(--font-size-16);
  font-weight: 700;
  color: var(--journal-ink);
}

.profile-actions {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 0.5rem;
}

.profile-layout {
  display: grid;
  gap: 1.5rem;
  padding-top: 1.25rem;
  grid-template-columns: minmax(0, 1fr) minmax(320px, 0.98fr);
}

.profile-layout--single {
  grid-template-columns: minmax(0, 1fr);
}

.profile-section {
  min-width: 0;
}

.profile-section + .profile-section {
  border-left: 1px solid color-mix(in srgb, var(--journal-border) 86%, transparent);
  padding-left: 1.5rem;
}

.profile-section-head {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-start;
  justify-content: space-between;
  gap: 0.8rem;
}

.profile-section-kicker {
  font-size: var(--font-size-0-72);
  font-weight: 700;
  letter-spacing: 0.16em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.profile-section-title {
  margin-top: 0.35rem;
  display: flex;
  align-items: center;
  gap: 0.6rem;
  font-size: var(--font-size-1-15);
  font-weight: 700;
  color: var(--journal-ink);
}

.profile-accent-icon {
  color: var(--journal-accent);
}

.profile-field-list {
  margin-top: 1rem;
  display: grid;
  gap: 0.85rem;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.profile-field-item,
.profile-report-meta__item {
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 84%, transparent);
  padding-bottom: 0.85rem;
}

.profile-field-value {
  margin-top: 0.45rem;
  font-size: var(--font-size-1-00);
  font-weight: 600;
  color: var(--journal-ink);
}

.profile-status {
  margin-top: 1rem;
  border-top: 1px solid color-mix(in srgb, var(--journal-border) 84%, transparent);
  padding-top: 1rem;
}

.profile-status__row {
  display: grid;
  gap: 0.75rem;
  grid-template-columns: minmax(0, 1fr) auto;
}

.profile-status__label,
.profile-status__value {
  display: inline-flex;
  align-items: center;
  gap: 0.45rem;
  font-size: var(--font-size-0-88);
  color: var(--journal-ink);
}

.profile-status__meta {
  margin-top: 0.5rem;
  font-size: var(--font-size-0-78);
  color: var(--journal-muted);
}

.profile-fieldset {
  margin-top: 1rem;
}

.profile-fieldset__legend {
  margin-bottom: 0.75rem;
  font-size: var(--font-size-0-88);
  font-weight: 600;
  color: var(--journal-ink);
}

.profile-format-grid {
  display: grid;
  gap: 0.75rem;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.journal-format-option {
  display: block;
  cursor: pointer;
  border-radius: 16px;
  border: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
  background: color-mix(in srgb, var(--journal-surface) 95%, var(--color-bg-base));
  padding: 0.8rem 0.95rem;
  transition:
    border-color 0.2s,
    background 0.2s;
}

.journal-format-option:hover {
  border-color: color-mix(in srgb, var(--journal-accent) 40%, transparent);
}

.journal-format-option--active {
  border-color: color-mix(in srgb, var(--journal-accent) 60%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 6%, transparent);
}

.profile-format-title {
  font-size: var(--font-size-0-875);
  font-weight: 600;
  color: var(--journal-ink);
}

.profile-format-copy {
  font-size: var(--font-size-0-75);
  color: var(--journal-muted);
}

.journal-btn--download {
  border-color: color-mix(in srgb, var(--journal-border) 88%, transparent);
}

.journal-chip {
  display: inline-flex;
  align-items: center;
  border-radius: 999px;
  padding: 0.4rem 0.8rem;
  font-size: var(--font-size-0-75);
  font-weight: 600;
}

.chip--primary {
  color: var(--journal-accent);
  background: color-mix(in srgb, var(--journal-accent) 10%, transparent);
}

.chip--success {
  color: var(--color-success);
  background: color-mix(in srgb, var(--color-success) 12%, transparent);
}

.chip--warning {
  color: var(--color-warning);
  background: color-mix(in srgb, var(--color-warning) 12%, transparent);
}

.chip--danger {
  color: var(--color-danger);
  background: color-mix(in srgb, var(--color-danger) 12%, transparent);
}

.profile-report-meta {
  margin-top: 1rem;
  display: grid;
  gap: 0.85rem;
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.profile-report-meta__value {
  margin-top: 0.45rem;
  font-size: var(--font-size-0-90);
  font-weight: 600;
  color: var(--journal-ink);
}

.profile-error-copy {
  margin-top: 0.75rem;
  font-size: var(--font-size-0-84);
  color: var(--color-danger);
}

.profile-inline-notice {
  margin-bottom: 1rem;
  border-inline-start: 2px solid color-mix(in srgb, var(--color-warning) 60%, transparent);
  background: color-mix(in srgb, var(--color-warning) 8%, transparent);
  padding: 0.8rem 0.95rem;
  font-size: var(--font-size-0-875);
  color: color-mix(in srgb, var(--color-warning) 88%, var(--journal-ink));
}

.profile-divider {
  margin: 0;
}

.status-dot {
  display: inline-block;
  width: 7px;
  height: 7px;
  border-radius: 50%;
  flex-shrink: 0;
}

.status-dot-ready {
  background: var(--color-success);
  box-shadow: 0 0 0 2px color-mix(in srgb, var(--color-success) 20%, transparent);
}

.status-dot-warning {
  background: var(--color-warning);
  box-shadow: 0 0 0 2px color-mix(in srgb, var(--color-warning) 18%, transparent);
}

.status-dot-idle {
  background: var(--color-text-muted);
  box-shadow: 0 0 0 2px color-mix(in srgb, var(--color-text-muted) 18%, transparent);
}

.status-dot-danger {
  background: var(--color-danger);
  box-shadow: 0 0 0 2px color-mix(in srgb, var(--color-danger) 16%, transparent);
}

@media (max-width: 1024px) {
  .profile-layout {
    grid-template-columns: minmax(0, 1fr);
  }

  .profile-section + .profile-section {
    border-left: 0;
    border-top: 1px solid color-mix(in srgb, var(--journal-border) 86%, transparent);
    padding-left: 0;
    padding-top: 1.25rem;
  }
}

@media (max-width: 1180px) {
  .profile-topbar-meta {
    width: 100%;
    justify-items: start;
  }

  .profile-head-stats,
  .profile-actions {
    justify-content: flex-start;
  }
}

@media (max-width: 720px) {
  .content-pane {
    padding-inline: 1rem;
  }

  .profile-field-list,
  .profile-format-grid,
  .profile-report-meta {
    grid-template-columns: minmax(0, 1fr);
  }
}
</style>
