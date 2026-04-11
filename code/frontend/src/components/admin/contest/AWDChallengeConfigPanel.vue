<script setup lang="ts">
import { computed } from 'vue'

import type { AdminContestChallengeData } from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'

const props = defineProps<{
  challengeLinks: AdminContestChallengeData[]
}>()

const emit = defineEmits<{
  create: []
  edit: [challenge: AdminContestChallengeData]
}>()

const sortedChallengeLinks = computed(() =>
  [...props.challengeLinks].sort(
    (left, right) => left.order - right.order || left.challenge_id.localeCompare(right.challenge_id)
  )
)

const summaryItems = computed(() => [
  {
    key: 'total',
    label: '已关联题目',
    value: String(sortedChallengeLinks.value.length),
    hint: '当前 AWD 赛事中可参与攻防的服务题目数量',
  },
  {
    key: 'configured',
    label: '已配 Checker',
    value: String(
      sortedChallengeLinks.value.filter(
        (item) =>
          Boolean(item.awd_checker_type) ||
          Object.keys(item.awd_checker_config || {}).length > 0
      ).length
    ),
    hint: '已写入 checker 类型或 checker 配置的题目数',
  },
  {
    key: 'http-standard',
    label: 'HTTP Standard',
    value: String(
      sortedChallengeLinks.value.filter((item) => item.awd_checker_type === 'http_standard').length
    ),
    hint: '已切到 HTTP 标准 Checker 的题目数',
  },
  {
    key: 'hidden',
    label: '隐藏题目',
    value: String(sortedChallengeLinks.value.filter((item) => !item.is_visible).length),
    hint: '当前不会直接对选手展示的赛事题目数',
  },
])

function getCheckerTypeLabel(value?: string): string {
  switch (value) {
    case 'legacy_probe':
      return '基础探活'
    case 'http_standard':
      return 'HTTP 标准 Checker'
    default:
      return '未配置'
  }
}

function getConfigSummary(item: AdminContestChallengeData): string {
  const config = item.awd_checker_config || {}
  const putFlag = readActionSummary(config.put_flag, 'PUT')
  const getFlag = readActionSummary(config.get_flag, 'GET')
  const havoc = readActionSummary(config.havoc, 'Havoc')
  const healthPath =
    typeof config.health_path === 'string' && config.health_path.trim() !== ''
      ? `Health ${config.health_path.trim()}`
      : ''

  return [putFlag, getFlag, havoc, healthPath].filter(Boolean).join(' · ') || '未配置动作摘要'
}

function readActionSummary(value: unknown, label: string): string {
  if (!value || typeof value !== 'object') {
    return ''
  }
  const item = value as Record<string, unknown>
  const path = typeof item.path === 'string' ? item.path : ''
  if (!path) {
    return label
  }
  return `${label} ${path}`
}

function getChallengeTitle(item: AdminContestChallengeData): string {
  return item.title?.trim() || `Challenge #${item.challenge_id}`
}
</script>

<template>
  <section class="space-y-6">
    <header class="panel-head panel-head--config">
      <div class="panel-copy workspace-tab-heading__main">
        <div class="journal-eyebrow">AWD Service Config</div>
        <h2 class="workspace-tab-heading__title">题目配置</h2>
        <p class="admin-page-copy">
          在这里管理赛事题目的 Checker 类型、动作配置和每轮 SLA / 防守分。
        </p>
      </div>

      <div class="metric-panel-grid metric-panel-default-surface">
        <article v-for="item in summaryItems" :key="item.key" class="journal-note metric-panel-card">
          <div class="journal-note-label metric-panel-label">{{ item.label }}</div>
          <div class="journal-note-value metric-panel-value">{{ item.value }}</div>
          <div class="journal-note-helper metric-panel-helper">{{ item.hint }}</div>
        </article>
      </div>
    </header>

    <section class="workspace-directory-section">
      <header class="config-list-head">
        <div class="workspace-tab-heading__main">
          <div class="journal-note-label">Contest Challenges</div>
          <h3 class="workspace-tab-heading__title">已关联题目</h3>
        </div>
        <button
          id="awd-challenge-config-create"
          type="button"
          class="rounded-xl bg-primary px-4 py-2 text-sm font-medium text-white transition hover:opacity-90"
          @click="emit('create')"
        >
          新增题目
        </button>
      </header>

      <AppEmpty
        v-if="sortedChallengeLinks.length === 0"
        title="当前赛事还没有关联题目"
        description="先关联赛事题目，再为每道服务配置 checker 与分值。"
        icon="FileChartColumnIncreasing"
      />

      <template v-else>
        <div class="config-directory-head" aria-hidden="true">
          <span>题目</span>
          <span>可见性</span>
          <span>分值</span>
          <span>Checker</span>
          <span>规则摘要</span>
          <span class="config-directory-head__actions">操作</span>
        </div>

        <article
          v-for="item in sortedChallengeLinks"
          :key="item.id"
          class="config-row"
        >
          <div class="config-row__identity">
            <h4 class="config-row__title">{{ getChallengeTitle(item) }}</h4>
            <p class="config-row__meta">
              {{ item.category || '未分类' }} · {{ item.difficulty || '未标记难度' }} · 顺序
              {{ item.order }}
            </p>
          </div>
          <div class="config-row__visibility">
            {{ item.is_visible ? '可见' : '隐藏' }}
          </div>
          <div class="config-row__scores">
            <p>{{ item.points }} 分</p>
            <p class="config-row__scores-sub">SLA {{ item.awd_sla_score ?? 0 }} / 防守 {{ item.awd_defense_score ?? 0 }}</p>
          </div>
          <div class="config-row__checker">
            {{ getCheckerTypeLabel(item.awd_checker_type) }}
          </div>
          <div class="config-row__summary">
            {{ getConfigSummary(item) }}
          </div>
          <div class="config-row__actions" role="group" :aria-label="`题目 ${getChallengeTitle(item)} 操作`">
            <button
              :id="`awd-challenge-config-edit-${item.id}`"
              type="button"
              class="rounded-xl border border-border px-3 py-2 text-sm font-medium text-[var(--color-text-primary)] transition hover:border-primary"
              @click="emit('edit', item)"
            >
              编辑配置
            </button>
          </div>
        </article>
      </template>
    </section>
  </section>
</template>

<style scoped>
.panel-head--config {
  display: grid;
  gap: 1.5rem;
}

.config-list-head {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 1rem;
  margin-bottom: 1.25rem;
}

.config-directory-head,
.config-row {
  display: grid;
  grid-template-columns: minmax(0, 1.4fr) 0.7fr 0.9fr 0.9fr minmax(0, 1.2fr) auto;
  gap: 1rem;
  align-items: center;
}

.config-directory-head {
  padding: 0 0 0.9rem;
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 85%, transparent);
  color: var(--journal-muted);
  font-size: 0.72rem;
  font-weight: 600;
  letter-spacing: 0.16em;
  text-transform: uppercase;
}

.config-directory-head__actions {
  text-align: right;
}

.config-row {
  padding: 1rem 0;
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 72%, transparent);
}

.config-row:last-child {
  border-bottom: none;
}

.config-row__title {
  margin: 0;
  color: var(--color-text-primary);
  font-size: 0.98rem;
  font-weight: 600;
}

.config-row__meta,
.config-row__scores-sub {
  margin: 0.35rem 0 0;
  color: var(--color-text-muted);
  font-size: 0.8rem;
}

.config-row__visibility,
.config-row__scores,
.config-row__checker,
.config-row__summary {
  color: var(--color-text-secondary);
  font-size: 0.9rem;
}

.config-row__scores p {
  margin: 0;
}

.config-row__summary {
  line-height: 1.6;
}

.config-row__actions {
  display: flex;
  justify-content: flex-end;
}

@media (max-width: 1100px) {
  .config-directory-head {
    display: none;
  }

  .config-row {
    grid-template-columns: 1fr;
    gap: 0.6rem;
    padding: 1rem 0;
  }

  .config-row__actions {
    justify-content: flex-start;
  }
}
</style>
