<script setup lang="ts">
import {
  EXTEND_DURATION_SECONDS,
  WARNING_THRESHOLD_SECONDS,
  formatRemainingTime,
  getInstanceStatusClass,
  getInstanceStatusLabel,
  getInstanceWaitingHint,
  useInstanceListPage,
} from '@/composables/useInstanceListPage'

const {
  loading,
  maxInstances,
  instances,
  runningCount,
  waitingCount,
  showWarning,
  warningInstance,
  copyAddress,
  extendTime,
  openTarget,
  destroyInstance,
  extendFromWarning,
  closeWarning,
} = useInstanceListPage()
</script>

<template>
  <section
    class="journal-shell journal-hero flex min-h-full flex-col space-y-6 rounded-[30px] border px-6 py-6 md:px-8"
  >
    <div class="grid gap-6 xl:grid-cols-[1.05fr_0.95fr]">
      <div>
        <div class="journal-eyebrow">Instance Console</div>
        <h1
          class="mt-3 text-3xl font-semibold tracking-tight text-[var(--journal-ink)] md:text-[2.45rem]"
        >
          我的实例
        </h1>
        <p class="mt-3 max-w-2xl text-sm leading-7 text-[var(--journal-muted)]">
          管理运行中与等待创建中的靶机实例，查看状态并执行延时或销毁。
        </p>
      </div>

      <article class="journal-brief rounded-[24px] border px-5 py-5">
        <div class="text-sm font-medium text-[var(--journal-ink)]">当前运行概况</div>
        <div class="mt-5 grid gap-3 sm:grid-cols-3">
          <div class="journal-note">
            <div class="journal-note-label">运行中</div>
            <div class="journal-note-value">{{ runningCount }}</div>
          </div>
          <div class="journal-note">
            <div class="journal-note-label">等待创建</div>
            <div class="journal-note-value">{{ waitingCount }}</div>
          </div>
          <div class="journal-note">
            <div class="journal-note-label">实例上限</div>
            <div class="journal-note-value">{{ maxInstances }}</div>
          </div>
        </div>
      </article>
    </div>

    <div class="instance-board mt-6 flex-1 px-1 pt-5 md:px-2 md:pt-6">
      <div v-if="loading" class="flex justify-center py-12">
        <div
          class="h-8 w-8 animate-spin rounded-full border-4 border-[var(--journal-control-border)] border-t-[var(--journal-accent)]"
        />
      </div>

      <div
        v-else-if="instances.length === 0"
        class="rounded-[22px] border border-dashed border-[var(--journal-soft-border)] bg-[var(--journal-surface-subtle)]/52 px-4 py-12 text-center"
      >
        <div class="text-sm text-[var(--journal-muted)]">暂无运行中或等待创建的实例</div>
        <router-link
          to="/challenges"
          class="mt-3 inline-block text-sm text-[var(--journal-accent)] hover:underline"
        >
          前往靶场列表创建实例
        </router-link>
      </div>

      <div v-else class="instance-list">
        <article v-for="instance in instances" :key="instance.id" class="instance-item">
          <div class="flex flex-wrap items-start justify-between gap-4">
            <div class="min-w-0">
              <h3 class="text-lg font-semibold text-[var(--journal-ink)]">
                {{ instance.challenge_title }}
              </h3>
              <div class="mt-3 flex flex-wrap gap-2">
                <span
                  class="rounded-full bg-[var(--journal-accent)]/10 px-2.5 py-0.5 text-xs font-medium text-[var(--journal-accent)]"
                >
                  {{ instance.category }}
                </span>
                <span
                  class="rounded-full bg-[var(--color-success)]/10 px-2.5 py-0.5 text-xs font-medium text-[var(--color-success)]"
                >
                  {{ instance.difficulty }}
                </span>
              </div>
            </div>

            <div class="instance-status text-sm">
              <span :class="getInstanceStatusClass(instance.status)">●</span>
              <span class="text-[var(--journal-muted)]">{{
                getInstanceStatusLabel(instance.status)
              }}</span>
            </div>
          </div>

          <div v-if="instance.status === 'running'" class="instance-meta mt-5">
            <div class="instance-meta__row">
              <span class="instance-meta__label">地址</span>
              <div class="flex flex-wrap items-center justify-end gap-2">
                <span class="font-mono text-sm text-[var(--journal-ink)]">
                  {{
                    instance.access_url ||
                    (instance.ssh_info ? `${instance.ssh_info.host}:${instance.ssh_info.port}` : '')
                  }}
                </span>
                <button
                  class="instance-action-link"
                  @click="
                    copyAddress(
                      instance.access_url ||
                        (instance.ssh_info ? `${instance.ssh_info.host}:${instance.ssh_info.port}` : '')
                    )
                  "
                >
                  复制
                </button>
                <button
                  v-if="instance.access_url"
                  class="instance-action-link"
                  @click="openTarget(instance.id)"
                >
                  打开目标
                </button>
              </div>
            </div>

            <div class="instance-meta__row">
              <span class="instance-meta__label">剩余</span>
              <span
                class="font-mono text-sm"
                :class="
                  instance.remaining < WARNING_THRESHOLD_SECONDS
                    ? 'font-semibold text-[var(--color-warning)]'
                    : 'text-[var(--journal-ink)]'
                "
              >
                {{ formatRemainingTime(instance.remaining) }}
              </span>
            </div>
          </div>
          <div
            v-else-if="instance.status === 'pending' || instance.status === 'creating'"
            class="mt-5 rounded-[14px] border border-[var(--journal-soft-border)] bg-[var(--journal-surface-subtle)]/84 px-4 py-3 text-xs leading-6 text-[var(--journal-muted)]"
          >
            {{ getInstanceWaitingHint(instance) }}
          </div>

          <div class="mt-5 flex flex-wrap gap-3">
            <button
              v-if="instance.status === 'running'"
              :disabled="instance.remaining_extends <= 0"
              class="journal-btn journal-btn--primary"
              @click="extendTime(instance.id)"
            >
              延时 +{{ EXTEND_DURATION_SECONDS / 60 }}min ({{ instance.remaining_extends }})
            </button>
            <button class="journal-btn journal-btn--danger" @click="destroyInstance(instance.id)">
              销毁
            </button>
          </div>
        </article>
      </div>
    </div>

    <div
      v-if="showWarning"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 px-4"
      @click.self="closeWarning"
    >
      <div class="warning-dialog w-full max-w-md rounded-[24px] border px-6 py-6 shadow-xl">
        <h3 class="text-lg font-semibold text-[var(--journal-ink)]">实例即将过期</h3>
        <p class="mt-2 text-sm leading-6 text-[var(--journal-muted)]">
          实例 "{{ warningInstance?.challenge_title }}" 剩余时间不足 5 分钟，是否延长？
        </p>
        <div class="mt-6 flex justify-end gap-3">
          <button class="journal-btn" @click="closeWarning">取消</button>
          <button class="journal-btn journal-btn--primary" @click="extendFromWarning">
            延长 {{ EXTEND_DURATION_SECONDS / 60 }} 分钟
          </button>
        </div>
      </div>
    </div>
  </section>
</template>

<style scoped>
.journal-shell {
  --journal-ink: var(--color-text-primary);
  --journal-muted: var(--color-text-secondary);
  --journal-accent: #4f46e5;
  --journal-border: color-mix(in srgb, var(--color-border-default) 82%, transparent);
  --journal-shell-border: color-mix(in srgb, var(--journal-border) 76%, transparent);
  --journal-soft-border: color-mix(in srgb, var(--journal-border) 68%, transparent);
  --journal-divider: color-mix(in srgb, var(--journal-border) 56%, transparent);
  --journal-control-border: color-mix(in srgb, var(--journal-border) 72%, transparent);
  --journal-surface: color-mix(in srgb, var(--color-bg-surface) 88%, var(--color-bg-base));
  --journal-surface-subtle: color-mix(in srgb, var(--color-bg-surface) 74%, var(--color-bg-base));
  font-family: 'Inter', 'Noto Sans SC', system-ui, sans-serif;
}

.journal-hero {
  border-color: var(--journal-border);
  background:
    radial-gradient(circle at top right, rgba(79, 70, 229, 0.06), transparent 20rem),
    linear-gradient(180deg, color-mix(in srgb, var(--journal-surface, var(--color-bg-surface)) 96%, var(--color-bg-base)), color-mix(in srgb, var(--journal-surface-subtle, var(--color-bg-elevated)) 94%, var(--color-bg-base)));
  border-radius: 16px !important;
  overflow: hidden;
  box-shadow: 0 18px 40px rgba(15, 23, 42, 0.05);
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

.journal-brief {
  border-color: var(--journal-border);
  background: var(--journal-surface-subtle);
}

.journal-note {
  border-radius: 16px;
  border: 1px solid var(--journal-shell-border);
  background: linear-gradient(180deg, color-mix(in srgb, var(--journal-surface) 94%, var(--color-bg-base)), color-mix(in srgb, var(--journal-surface-subtle) 96%, var(--color-bg-base)));
  padding: 0.875rem 1rem;
}

.journal-note-label {
  font-size: 0.68rem;
  font-weight: 600;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.journal-note-value {
  margin-top: 0.65rem;
  font-size: 1.05rem;
  font-weight: 600;
  color: var(--journal-ink);
}

.instance-board {
  border-top: 1px dashed var(--journal-divider);
}

.instance-list {
  border-radius: 22px;
  border: 1px solid var(--journal-shell-border);
  background: color-mix(in srgb, var(--journal-surface, var(--color-bg-surface)) 92%, var(--color-bg-base));
}

.instance-item {
  padding: 1rem 1.1rem;
}

.instance-item + .instance-item {
  border-top: 1px dashed var(--journal-divider);
}

.instance-status {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
}

.instance-meta {
  border-radius: 18px;
  border: 1px solid var(--journal-shell-border);
  background: color-mix(in srgb, var(--journal-surface, var(--color-bg-surface)) 92%, var(--color-bg-base));
}

.instance-meta__row {
  display: flex;
  justify-content: space-between;
  gap: 1rem;
  padding: 0.85rem 1rem;
}

.instance-meta__row + .instance-meta__row {
  border-top: 1px dashed var(--journal-divider);
}

.instance-meta__label {
  font-size: 0.78rem;
  font-weight: 600;
  color: var(--journal-muted);
}

.instance-action-link {
  border-radius: 999px;
  padding: 0.25rem 0.65rem;
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--journal-accent);
  transition: background 0.15s;
}

.instance-action-link:hover {
  background: color-mix(in srgb, var(--journal-accent) 8%, transparent);
}

.journal-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  border-radius: 10px;
  border: 1px solid var(--journal-control-border);
  background: var(--journal-surface);
  padding: 0.5rem 1rem;
  font-size: 0.84rem;
  font-weight: 600;
  color: var(--journal-muted);
  transition: all 0.15s;
}

.journal-btn--primary {
  border-color: color-mix(in srgb, var(--journal-accent) 50%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 8%, transparent);
  color: var(--journal-accent);
}

.journal-btn--danger {
  border-color: rgba(239, 68, 68, 0.2);
  background: rgba(239, 68, 68, 0.08);
  color: var(--color-danger);
}

.journal-btn:disabled {
  cursor: not-allowed;
  opacity: 0.55;
}

.warning-dialog {
  border-color: var(--journal-shell-border);
  background: linear-gradient(180deg, color-mix(in srgb, var(--journal-surface) 96%, var(--color-bg-base)), color-mix(in srgb, var(--journal-surface-subtle) 94%, var(--color-bg-base)));
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

:global([data-theme='dark']) .journal-note,
:global([data-theme='dark']) .instance-list,
:global([data-theme='dark']) .instance-meta,
:global([data-theme='dark']) .warning-dialog,
:global([data-theme='dark']) .journal-btn {
  background: rgba(15, 23, 42, 0.42);
}
</style>
