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
    class="journal-shell journal-shell-user journal-hero flex min-h-full flex-1 flex-col rounded-[30px] border px-6 py-6 md:px-8"
  >
    <div class="instance-page">
      <header class="instance-topbar">
        <div class="instance-heading">
          <div class="workspace-overline">Instances</div>
          <h1 class="instance-title workspace-page-title">我的实例</h1>
          <p class="instance-subtitle">
            管理运行中与等待创建中的靶机实例，查看状态并执行延时或销毁。
          </p>
        </div>
      </header>

      <section class="instance-summary">
        <div class="instance-summary-title">当前运行概况</div>
        <div class="instance-summary-grid metric-panel-grid">
          <div class="instance-summary-item metric-panel-card">
            <div class="instance-summary-label metric-panel-label">运行中</div>
            <div class="instance-summary-value metric-panel-value">{{ runningCount }}</div>
            <div class="instance-summary-helper metric-panel-helper">
              当前仍在运行、可直接访问的实例数量
            </div>
          </div>
          <div class="instance-summary-item metric-panel-card">
            <div class="instance-summary-label metric-panel-label">等待创建</div>
            <div class="instance-summary-value metric-panel-value">{{ waitingCount }}</div>
            <div class="instance-summary-helper metric-panel-helper">
              已经提交创建请求、正在排队或启动中的实例数量
            </div>
          </div>
          <div class="instance-summary-item metric-panel-card">
            <div class="instance-summary-label metric-panel-label">实例上限</div>
            <div class="instance-summary-value metric-panel-value">{{ maxInstances }}</div>
            <div class="instance-summary-helper metric-panel-helper">
              当前账号最多可同时保留的实例数量
            </div>
          </div>
        </div>
      </section>

      <div v-if="loading" class="instance-loading">
        <div class="instance-loading-spinner" />
      </div>

      <div v-else-if="instances.length === 0" class="instance-empty">
        <div class="instance-empty-title">暂无运行中或等待创建的实例</div>
        <router-link to="/challenges" class="instance-empty-link">前往靶场列表创建实例</router-link>
      </div>

      <section v-else class="instance-directory" aria-label="实例目录">
        <div class="instance-directory-top">
          <h2 class="instance-directory-title">实例列表</h2>
          <div class="instance-directory-meta">共 {{ instances.length }} 个实例</div>
        </div>

        <div class="instance-directory-head">
          <span>题目</span>
          <span>访问</span>
          <span>状态</span>
          <span>剩余时间</span>
          <span>操作</span>
        </div>

        <article v-for="instance in instances" :key="instance.id" class="instance-row">
          <div class="instance-row-main">
            <h2 class="instance-row-title" :title="instance.challenge_title">
              {{ instance.challenge_title }}
            </h2>
            <div class="instance-row-tags">
              <span class="instance-chip instance-chip-category">{{ instance.category }}</span>
              <span class="instance-chip instance-chip-difficulty">{{ instance.difficulty }}</span>
            </div>
          </div>

          <div class="instance-row-access">
            <template v-if="instance.status === 'running'">
              <div
                class="instance-row-mono instance-row-access-value"
                :title="
                  instance.access_url ||
                  (instance.ssh_info ? `${instance.ssh_info.host}:${instance.ssh_info.port}` : '')
                "
              >
                {{
                  instance.access_url ||
                  (instance.ssh_info ? `${instance.ssh_info.host}:${instance.ssh_info.port}` : '')
                }}
              </div>
              <div class="instance-row-inline-actions">
                <button
                  type="button"
                  class="instance-link-btn"
                  @click="
                    copyAddress(
                      instance.access_url ||
                        (instance.ssh_info
                          ? `${instance.ssh_info.host}:${instance.ssh_info.port}`
                          : '')
                    )
                  "
                >
                  复制
                </button>
                <button
                  v-if="instance.access_url"
                  type="button"
                  class="instance-link-btn"
                  @click="openTarget(instance.id)"
                >
                  打开目标
                </button>
              </div>
            </template>
            <div v-else class="instance-row-note">
              {{ getInstanceWaitingHint(instance) }}
            </div>
          </div>

          <div class="instance-row-status">
            <span class="instance-state-chip">
              <span :class="getInstanceStatusClass(instance.status)">●</span>
              <span>{{ getInstanceStatusLabel(instance.status) }}</span>
            </span>
          </div>

          <div class="instance-row-remaining">
            <span
              v-if="instance.status === 'running'"
              class="instance-row-mono"
              :class="
                instance.remaining < WARNING_THRESHOLD_SECONDS ? 'instance-row-mono-warning' : ''
              "
            >
              {{ formatRemainingTime(instance.remaining) }}
            </span>
            <span v-else class="instance-row-note">
              {{
                instance.status === 'failed'
                  ? '启动失败'
                  : instance.status === 'crashed'
                    ? '运行异常'
                    : '等待创建'
              }}
            </span>
          </div>

          <div class="instance-row-actions">
            <button
              v-if="instance.status === 'running' && instance.share_scope !== 'shared'"
              :disabled="instance.remaining_extends <= 0"
              class="instance-btn instance-btn-primary"
              @click="extendTime(instance.id)"
            >
              延时 +{{ EXTEND_DURATION_SECONDS / 60 }}min
            </button>
            <button
              v-if="instance.share_scope !== 'shared'"
              class="instance-btn instance-btn-danger"
              @click="destroyInstance(instance.id)"
            >
              销毁
            </button>
            <span v-if="instance.share_scope === 'shared'" class="instance-row-note">系统托管</span>
          </div>
        </article>
      </section>
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
          <button class="instance-btn" @click="closeWarning">取消</button>
          <button class="instance-btn instance-btn-primary" @click="extendFromWarning">
            延长 {{ EXTEND_DURATION_SECONDS / 60 }} 分钟
          </button>
        </div>
      </div>
    </div>
  </section>
</template>

<style scoped>
.journal-shell {
  --journal-shell-accent: color-mix(in srgb, var(--color-primary) 86%, var(--journal-ink));
}

.instance-page {
  display: flex;
  min-height: 100%;
  flex: 1 1 auto;
  flex-direction: column;
}

.instance-subtitle {
  max-width: 720px;
}

.instance-directory-head {
  font-size: var(--font-size-11);
  font-weight: 700;
  letter-spacing: 0.16em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.instance-loading {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 80px 0;
}

.instance-loading-spinner {
  width: 32px;
  height: 32px;
  border: 4px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
  border-top-color: var(--journal-accent);
  border-radius: 999px;
  animation: instanceSpin 900ms linear infinite;
}

.instance-empty {
  margin-top: 24px;
  padding: 32px 0;
  border-top: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
  text-align: center;
}

.instance-empty-title {
  font-size: var(--font-size-14);
  color: var(--journal-muted);
}

.instance-empty-link {
  display: inline-block;
  margin-top: 12px;
  font-size: var(--font-size-14);
  font-weight: 600;
  color: var(--journal-accent);
}

.instance-directory {
  margin-top: 24px;
}

.instance-directory-head {
  display: grid;
  grid-template-columns: minmax(0, 1.25fr) minmax(220px, 1.2fr) 140px 160px 220px;
  gap: 16px;
  padding: 0 0 12px;
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
}

.instance-row {
  display: grid;
  grid-template-columns: minmax(0, 1.25fr) minmax(220px, 1.2fr) 140px 160px 220px;
  gap: 16px;
  align-items: center;
  padding: 18px 0;
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
}

.instance-row-title {
  font-family: var(--font-family-mono);
  font-size: var(--font-size-18);
  font-weight: 700;
  line-height: 1.35;
  color: var(--journal-ink);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.instance-row-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 10px;
}

.instance-chip {
  display: inline-flex;
  align-items: center;
  min-height: 26px;
  padding: 0 9px;
  border-radius: 8px;
  font-size: var(--font-size-12);
  font-weight: 600;
}

.instance-chip-category {
  background: color-mix(in srgb, var(--journal-accent) 10%, transparent);
  color: var(--journal-accent);
}

.instance-chip-difficulty {
  background: color-mix(in srgb, var(--color-success) 10%, transparent);
  color: var(--color-success);
}

.instance-row-access,
.instance-row-remaining {
  min-width: 0;
  font-size: var(--font-size-13);
  line-height: 1.6;
  color: var(--journal-muted);
}

.instance-row-mono {
  font-family: var(--font-family-mono);
  color: var(--journal-ink);
}

.instance-row-access-value {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.instance-row-mono-warning {
  font-weight: 700;
  color: var(--color-warning);
}

.instance-row-inline-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 8px;
}

.instance-link-btn {
  padding: 0;
  border: 0;
  background: transparent;
  font-size: var(--font-size-13);
  font-weight: 600;
  color: var(--journal-accent);
  cursor: pointer;
}

.instance-row-note {
  font-size: var(--font-size-13);
  line-height: 1.6;
  color: var(--journal-muted);
}

.instance-state-chip {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  min-height: 28px;
  padding: 0 10px;
  border-radius: 8px;
  background: color-mix(in srgb, var(--journal-accent) 10%, transparent);
  font-size: var(--font-size-12);
  font-weight: 600;
  color: var(--journal-accent);
}

.instance-row-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.instance-btn-danger {
  color: var(--color-danger);
}

.warning-dialog {
  border-color: var(--journal-border);
  background: color-mix(in srgb, var(--journal-surface) 98%, var(--color-bg-base));
}

@keyframes instanceSpin {
  from {
    transform: rotate(0deg);
  }

  to {
    transform: rotate(360deg);
  }
}

@media (max-width: 1180px) {
  .instance-directory-head {
    display: none;
  }

  .instance-row {
    grid-template-columns: 1fr;
  }
}
</style>
