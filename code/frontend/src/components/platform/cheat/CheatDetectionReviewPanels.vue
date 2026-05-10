<script setup lang="ts">
import { ArrowRight } from 'lucide-vue-next'

import type { AdminCheatDetectionData } from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'

type CheatQuickAction = {
  title: string
  description: string
  actionLabel: string
  query: Record<string, string>
}

defineProps<{
  riskData: AdminCheatDetectionData
  quickActions: ReadonlyArray<CheatQuickAction>
  formatDateTime: (value: string) => string
}>()

const emit = defineEmits<{
  openAudit: [query: Record<string, string>]
}>()

function handleOpenAudit(query: Record<string, string>): void {
  emit('openAudit', query)
}
</script>

<template>
  <section class="workspace-directory-section cheat-directory-section">
    <header class="list-heading">
      <div>
        <div class="workspace-overline">Compliance Risk / Burst</div>
        <h2 class="list-heading__title">高频提交账号</h2>
      </div>
    </header>

    <AppEmpty
      v-if="!riskData.suspects.length"
      class="cheat-empty-state"
      icon="ShieldCheck"
      title="当前没有超过阈值的高频提交账号"
      description="说明最近统计窗口内还没有明显的提交样本超过安全阈值。"
    />

    <div v-else class="cheat-directory-list">
      <button
        v-for="suspect in riskData.suspects"
        :key="suspect.user_id"
        type="button"
        class="cheat-directory-row"
        @click="handleOpenAudit({ action: 'submit', actor_user_id: String(suspect.user_id) })"
      >
        <div class="cheat-directory-row-main">
          <div class="cheat-directory-row-title">{{ suspect.username }}</div>
        </div>
        <div class="cheat-directory-row-copy">{{ suspect.reason }}</div>
        <div class="cheat-directory-row-meta">
          <span class="cheat-badge cheat-badge--warning">{{ suspect.submit_count }} 次提交</span>
          <span class="cheat-meta-text">最近出现 {{ formatDateTime(suspect.last_seen_at) }}</span>
          <span class="cheat-link-hint">
            审计复核
            <ArrowRight class="h-3 w-3" />
          </span>
        </div>
      </button>
    </div>
  </section>

  <section class="workspace-directory-section cheat-directory-section">
    <header class="list-heading">
      <div>
        <div class="workspace-overline">Compliance Risk / Network</div>
        <h2 class="list-heading__title">共享 IP 线索</h2>
      </div>
    </header>

    <AppEmpty
      v-if="!riskData.shared_ips.length"
      class="cheat-empty-state"
      icon="ShieldCheck"
      title="当前没有共享 IP 线索"
      description="最近 24 小时内未监测到不同账号从同一公网地址密集登录。"
    />

    <div v-else class="cheat-directory-list">
      <button
        v-for="group in riskData.shared_ips"
        :key="group.ip"
        type="button"
        class="cheat-directory-row"
        @click="handleOpenAudit({ action: 'login' })"
      >
        <div class="cheat-directory-row-main">
          <div class="cheat-directory-row-title cheat-directory-row-title--mono">
            {{ group.ip }}
          </div>
        </div>
        <div class="cheat-directory-row-copy">涉及账号：{{ group.usernames.join('、') }}</div>
        <div class="cheat-directory-row-meta">
          <span class="cheat-badge">{{ group.user_count }} 个账号</span>
          <span class="cheat-meta-text">多见于短时集中登录行为</span>
          <span class="cheat-link-hint">
            追踪登录
            <ArrowRight class="h-3 w-3" />
          </span>
        </div>
      </button>
    </div>
  </section>

  <section class="workspace-directory-section cheat-directory-section">
    <header class="list-heading">
      <div>
        <div class="workspace-overline">Analysis Shortcuts</div>
        <h2 class="list-heading__title">审计联动</h2>
      </div>
    </header>

    <div class="quick-action-directory">
      <button
        v-for="action in quickActions"
        :key="action.title"
        type="button"
        class="quick-action-row"
        @click="handleOpenAudit(action.query)"
      >
        <div class="cheat-directory-row-main">
          <div class="cheat-directory-row-title">{{ action.title }}</div>
        </div>
        <div class="cheat-directory-row-copy">{{ action.description }}</div>
        <div class="cheat-directory-row-meta">
          <span class="cheat-badge cheat-badge--muted">{{ action.actionLabel }}</span>
          <span class="cheat-link-hint">
            打开详情
            <ArrowRight class="h-3 w-3" />
          </span>
        </div>
      </button>
    </div>
  </section>
</template>

<style scoped>
.cheat-directory-section {
  display: grid;
  gap: var(--space-4);
  padding: 0;
}

.cheat-directory-list,
.quick-action-directory {
  display: grid;
  gap: var(--space-3);
}

.cheat-directory-row,
.quick-action-row {
  display: grid;
  grid-template-columns: minmax(10rem, 0.85fr) minmax(16rem, 1.15fr) auto;
  align-items: center;
  gap: var(--space-4);
  width: 100%;
  border: 1px solid var(--cheat-card-border);
  border-radius: 1rem;
  background: color-mix(in srgb, var(--workspace-panel) 88%, transparent);
  padding: var(--space-4);
  text-align: left;
  transition:
    border-color 160ms ease,
    background-color 160ms ease,
    color 160ms ease;
}

.cheat-directory-row:hover,
.quick-action-row:hover,
.cheat-directory-row:focus-visible,
.quick-action-row:focus-visible {
  border-color: color-mix(in srgb, var(--journal-accent) 24%, var(--cheat-card-border));
  background: color-mix(in srgb, var(--workspace-panel-soft) 92%, transparent);
  outline: none;
}

.cheat-directory-row-main {
  min-width: 0;
}

.cheat-directory-row-title {
  font-size: var(--font-size-15);
  font-weight: 700;
  color: var(--journal-ink);
}

.cheat-directory-row-title--mono {
  font-family: var(--font-family-mono);
}

.cheat-directory-row-copy {
  font-size: var(--font-size-13);
  line-height: 1.7;
  color: var(--journal-muted);
}

.cheat-directory-row-meta {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: flex-end;
  gap: var(--space-3);
}

.cheat-badge {
  display: inline-flex;
  align-items: center;
  min-height: 1.9rem;
  border-radius: 999px;
  border: 1px solid color-mix(in srgb, var(--journal-border) 76%, transparent);
  background: color-mix(in srgb, var(--workspace-panel-soft) 84%, transparent);
  padding: 0 var(--space-3);
  font-size: var(--font-size-11);
  font-weight: 700;
  color: var(--journal-muted);
  white-space: nowrap;
}

.cheat-badge--warning {
  border-color: color-mix(in srgb, var(--color-warning) 28%, transparent);
  background: color-mix(in srgb, var(--color-warning) 10%, transparent);
  color: color-mix(in srgb, var(--color-warning) 86%, var(--journal-ink));
}

.cheat-badge--muted {
  border-color: color-mix(in srgb, var(--journal-border) 68%, transparent);
}

.cheat-meta-text {
  font-size: var(--font-size-12);
  color: var(--journal-muted);
  white-space: nowrap;
}

.cheat-link-hint {
  display: inline-flex;
  align-items: center;
  gap: var(--space-1-5);
  font-size: var(--font-size-12);
  font-weight: 700;
  color: var(--workspace-brand-ink);
  white-space: nowrap;
}

.cheat-empty-state {
  border-top-color: var(--cheat-divider);
  border-bottom-color: var(--cheat-divider);
}

@media (max-width: 1100px) {
  .cheat-directory-row,
  .quick-action-row {
    grid-template-columns: 1fr;
  }

  .cheat-directory-row-meta {
    justify-content: flex-start;
  }
}
</style>
