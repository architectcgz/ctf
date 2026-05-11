<script setup lang="ts">
import { ref } from 'vue'
import { Activity, Clock3, Server } from 'lucide-vue-next'

import type { ChallengeDifficulty } from '@/api/contracts'
import { ChallengeCategoryPill, ChallengeDifficultyText, getChallengeDifficultyColor } from '@/entities/challenge'
import {
  EXTEND_DURATION_SECONDS,
  WARNING_THRESHOLD_SECONDS,
  canOpenInstanceInBrowser,
  formatInstanceAccessDisplay,
  formatRemainingTime,
  getInstanceStatusClass,
  getInstanceStatusLabel,
  getInstanceWaitingHint,
  isInstanceManualActionAllowed,
  useInstanceListPage,
  useInstanceWarningFocus,
} from '@/features/instance-list'

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

const warningCloseButton = ref<HTMLButtonElement | null>(null)
useInstanceWarningFocus({ showWarning, warningCloseButton })

function difficultyPillStyle(difficulty: ChallengeDifficulty): Record<string, string> {
  const color = getChallengeDifficultyColor(difficulty)
  return {
    '--instance-difficulty-pill-color': color,
    '--instance-difficulty-pill-bg': `color-mix(in srgb, ${color} 10%, transparent)`,
    '--instance-difficulty-pill-border': `color-mix(in srgb, ${color} 22%, transparent)`,
  }
}
</script>

<template>
  <section
    class="workspace-shell journal-shell journal-shell-user journal-hero flex min-h-full flex-1 flex-col"
  >
    <main class="content-pane">
      <div class="instance-page">
        <header class="workspace-page-header instance-topbar">
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
            <div class="instance-summary-item progress-card metric-panel-card">
              <div class="instance-summary-label progress-card-label metric-panel-label">
                <span>运行中</span>
                <Activity class="h-4 w-4" />
              </div>
              <div class="instance-summary-value progress-card-value metric-panel-value">
                {{ runningCount }}
              </div>
              <div class="instance-summary-helper progress-card-hint metric-panel-helper">
                当前仍在运行、可直接访问的实例数量
              </div>
            </div>
            <div class="instance-summary-item progress-card metric-panel-card">
              <div class="instance-summary-label progress-card-label metric-panel-label">
                <span>等待创建</span>
                <Clock3 class="h-4 w-4" />
              </div>
              <div class="instance-summary-value progress-card-value metric-panel-value">
                {{ waitingCount }}
              </div>
              <div class="instance-summary-helper progress-card-hint metric-panel-helper">
                已经提交创建请求、正在排队或启动中的实例数量
              </div>
            </div>
            <div class="instance-summary-item progress-card metric-panel-card">
              <div class="instance-summary-label progress-card-label metric-panel-label">
                <span>实例上限</span>
                <Server class="h-4 w-4" />
              </div>
              <div class="instance-summary-value progress-card-value metric-panel-value">
                {{ maxInstances }}
              </div>
              <div class="instance-summary-helper progress-card-hint metric-panel-helper">
                当前账号最多可同时保留的实例数量
              </div>
            </div>
          </div>
        </section>

        <section
          class="student-directory-section workspace-directory-section"
          aria-label="实例目录"
        >
          <section class="student-directory-shell instance-directory workspace-directory-list">
            <header class="student-directory-shell__head">
              <div class="student-directory-shell__heading">
                <div class="journal-note-label student-directory-shell__eyebrow">
                  Instance Directory
                </div>
                <h2 class="student-directory-shell__title">实例列表</h2>
              </div>
              <div class="student-directory-shell__meta">共 {{ instances.length }} 个实例</div>
            </header>

            <div
              v-if="loading"
              class="instance-loading student-directory-state workspace-directory-loading"
            >
              <div class="student-directory-spinner" />
            </div>

            <div v-else-if="instances.length === 0" class="instance-empty student-directory-state">
              <div class="instance-empty-title">暂无运行中或等待创建的实例</div>
              <router-link to="/challenges" class="instance-empty-link">
                前往靶场列表创建实例
              </router-link>
            </div>

            <template v-else>
              <div class="workspace-directory-grid-head instance-directory-head">
                <span>题目</span>
                <span>分类</span>
                <span>难度</span>
                <span>访问</span>
                <span>状态</span>
                <span>剩余时间</span>
                <span>操作</span>
              </div>

              <article
                v-for="instance in instances"
                :key="instance.id"
                class="workspace-directory-grid-row instance-row"
              >
                <div class="workspace-directory-cell instance-row-main">
                  <h2
                    class="instance-row-title workspace-directory-row-title"
                    :title="instance.challenge_title"
                  >
                    {{ instance.challenge_title }}
                  </h2>
                </div>

                <div class="instance-row-category">
                  <ChallengeCategoryPill :category="instance.category" />
                </div>

                <div class="instance-row-difficulty">
                  <span
                    class="workspace-directory-status-pill instance-chip instance-chip-difficulty"
                    :style="difficultyPillStyle(instance.difficulty)"
                  >
                    <ChallengeDifficultyText :difficulty="instance.difficulty" />
                  </span>
                </div>

                <div class="workspace-directory-compact-text instance-row-access">
                  <template v-if="instance.status === 'running'">
                    <div
                      class="workspace-directory-mono instance-row-mono instance-row-access-value"
                      :title="formatInstanceAccessDisplay(instance)"
                    >
                      {{ formatInstanceAccessDisplay(instance) }}
                    </div>
                    <div class="instance-row-inline-actions">
                      <button
                        type="button"
                        class="ui-btn ui-btn--link instance-link-btn"
                        @click="copyAddress(formatInstanceAccessDisplay(instance))"
                      >
                        复制
                      </button>
                      <button
                        v-if="canOpenInstanceInBrowser(instance)"
                        type="button"
                        class="ui-btn ui-btn--link instance-link-btn"
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
                  <span class="workspace-directory-status-pill instance-state-chip">
                    <span :class="getInstanceStatusClass(instance.status)">●</span>
                    <span>{{ getInstanceStatusLabel(instance.status) }}</span>
                  </span>
                </div>

                <div class="instance-row-remaining">
                  <span
                    v-if="instance.status === 'running'"
                    class="workspace-directory-mono instance-row-mono"
                    :class="
                      instance.remaining < WARNING_THRESHOLD_SECONDS
                        ? 'instance-row-mono-warning'
                        : ''
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

                <div class="workspace-directory-row-actions instance-row-actions">
                  <button
                    v-if="instance.status === 'running' && isInstanceManualActionAllowed(instance)"
                    :disabled="instance.remaining_extends <= 0"
                    class="ui-btn ui-btn--sm ui-btn--primary"
                    @click="extendTime(instance.id)"
                  >
                    延时 +{{ EXTEND_DURATION_SECONDS / 60 }}min
                  </button>
                  <button
                    v-if="isInstanceManualActionAllowed(instance)"
                    class="ui-btn ui-btn--sm ui-btn--danger"
                    @click="destroyInstance(instance.id)"
                  >
                    销毁
                  </button>
                  <span v-if="!isInstanceManualActionAllowed(instance)" class="instance-row-note">
                    系统托管
                  </span>
                </div>
              </article>
            </template>
          </section>
        </section>
      </div>
    </main>

    <div
      v-if="showWarning"
      class="instance-warning-overlay"
      role="presentation"
      @click.self="closeWarning"
    >
      <div
        class="warning-dialog"
        role="dialog"
        aria-modal="true"
        aria-labelledby="instance-warning-title"
        aria-describedby="instance-warning-description"
      >
        <div class="warning-dialog__header">
          <h3 id="instance-warning-title" class="warning-dialog__title">实例即将过期</h3>
          <button
            ref="warningCloseButton"
            type="button"
            class="ui-btn ui-btn--sm ui-btn--secondary"
            aria-label="关闭实例过期提醒"
            @click="closeWarning"
          >
            关闭
          </button>
        </div>
        <p id="instance-warning-description" class="warning-dialog__description">
          实例 "{{ warningInstance?.challenge_title }}" 剩余时间不足 5 分钟，是否延长？
        </p>
        <div class="warning-dialog__actions">
          <button class="ui-btn ui-btn--secondary" @click="closeWarning">取消</button>
          <button class="ui-btn ui-btn--primary" @click="extendFromWarning">
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
}

.instance-empty {
  display: grid;
  align-content: center;
  justify-items: center;
  padding: var(--space-8) var(--space-4);
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
  --workspace-directory-grid-columns: minmax(0, 1.2fr) 7rem 7rem minmax(13.75rem, 1.2fr) 8.75rem
    10rem 13.75rem;
}

.instance-row-title {
  font-size: var(--font-size-15);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.instance-row-category,
.instance-row-difficulty {
  min-width: 0;
}

.instance-chip-difficulty {
  border-color: var(--instance-difficulty-pill-border);
  background: var(--instance-difficulty-pill-bg);
}

.instance-row-access,
.instance-row-remaining {
  min-width: 0;
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

.instance-row-note {
  font-size: var(--font-size-13);
  line-height: 1.6;
  color: var(--journal-muted);
}

.instance-state-chip {
  gap: var(--space-2);
  border-color: color-mix(in srgb, var(--journal-accent) 22%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 10%, transparent);
  color: var(--journal-accent);
}

.instance-status-dot--warning {
  color: var(--color-warning);
}

.instance-status-dot--success {
  color: var(--color-success);
}

.instance-status-dot--danger {
  color: var(--color-danger);
}

.instance-status-dot--muted {
  color: var(--color-text-muted);
}

.instance-row-actions {
  flex-wrap: wrap;
  justify-content: flex-start;
}

.instance-warning-overlay {
  position: fixed;
  inset: 0;
  z-index: 50;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: var(--space-4);
  background: color-mix(in srgb, var(--color-bg-base) 64%, transparent);
}

.warning-dialog {
  width: min(100%, 28rem);
  max-height: min(32rem, calc(100vh - var(--space-12)));
  overflow: auto;
  padding: var(--space-6);
  border: 1px solid var(--journal-border);
  border-radius: var(--ui-dialog-radius-wide);
  border-color: var(--journal-border);
  background: color-mix(in srgb, var(--journal-surface) 98%, var(--color-bg-base));
  box-shadow: var(--ui-dialog-shadow);
}

.warning-dialog__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-4);
}

.warning-dialog__title {
  font-size: var(--font-size-18);
  font-weight: 700;
  color: var(--journal-ink);
}

.warning-dialog__description {
  margin-top: var(--space-2);
  font-size: var(--font-size-14);
  line-height: 1.7;
  color: var(--journal-muted);
}

.warning-dialog__actions {
  display: flex;
  justify-content: flex-end;
  gap: var(--space-3);
  margin-top: var(--space-6);
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
