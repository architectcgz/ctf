<script setup lang="ts">
import { computed } from 'vue'

import type { AWDReadinessData, AWDReadinessItemData } from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'
import { 
  ShieldCheck, 
  AlertTriangle, 
  AlertCircle 
} from 'lucide-vue-next'

const props = withDefaults(
  defineProps<{
    readiness: AWDReadinessData | null
    loading: boolean
    actionLabel?: string
  }>(),
  {
    actionLabel: '编辑配置',
  }
)

const emit = defineEmits<{
  editConfig: [challengeId: string]
}>()

const summaryItems = computed(() => {
  const readiness = props.readiness
  return [
    { key: 'passed', label: '最近通过', value: readiness?.passed_challenges ?? 0 },
    { key: 'pending', label: '未验证', value: readiness?.pending_challenges ?? 0 },
    { key: 'failed', label: '最近失败', value: readiness?.failed_challenges ?? 0 },
    { key: 'stale', label: '待重新验证', value: readiness?.stale_challenges ?? 0 },
    { key: 'missing', label: '未配 Checker', value: readiness?.missing_checker_challenges ?? 0 },
  ]
})

const readinessDecision = computed(() => {
  if (!props.readiness) return { key: 'pending', title: '正在审计...', description: '请稍候，系统正在扫描题目状态。' }
  if (props.readiness.ready) return { key: 'ready', title: '环境已就绪', description: '所有服务 Checker 均已通过验证，可以安全开启竞赛。' }
  return { key: 'blocked', title: '存在阻塞风险', description: '部分题目仍有阻塞或校验失败，将限制运维操作。' }
})

const blockingActionLabels = computed(() => {
  if (!props.readiness) return []
  const labels: string[] = []
  const actions = props.readiness.blocking_actions || []
  if (actions.includes('start_contest')) labels.push('开启比赛')
  if (actions.includes('create_round')) labels.push('创建轮次')
  if (actions.includes('run_current_round_check')) labels.push('即时巡检')
  return labels
})

const hasGlobalBlockingReasons = computed(() => (props.readiness?.global_blocking_reasons?.length ?? 0) > 0)
const globalBlockingReasons = computed(() => props.readiness?.global_blocking_reasons || [])

const blockingItems = computed(() => props.readiness?.items || [])
const blockingEmptyDescription = computed(() => props.readiness?.ready ? '所有题目均已通过自动审计。' : '题目级别暂无直接阻塞，请检查系统级配置。')

function getGlobalReasonCopy(reason: string): string {
  switch (reason) {
    case 'no_challenges': return '当前赛事还没有关联题目，无法执行开赛关键动作。'
    case 'missing_teams': return '竞赛中尚未加入任何参赛队伍。'
    case 'missing_challenges': return '题目池为空，至少需要关联一道题目。'
    case 'invalid_schedule': return '赛程时间设置有误或尚未开始。'
    default: return reason
  }
}

function getValidationStateLabel(item: AWDReadinessItemData): string {
  switch (item.validation_state) {
    case 'passed': return '最近通过'
    case 'failed': return '最近失败'
    case 'stale': return '待重新验证'
    default: return '未验证'
  }
}

function getBlockingReasonLabel(item: AWDReadinessItemData): string {
  switch (item.blocking_reason) {
    case 'missing_checker': return '未配置 Checker'
    case 'invalid_checker_config': return 'Checker 配置不可用'
    case 'pending_validation': return '还没有试跑结果'
    case 'last_preview_failed': return '最近一次试跑失败'
    case 'validation_stale': return '配置变更后待重新验证'
    default: return item.blocking_reason
  }
}

function formatDateTime(value?: string): string {
  if (!value) return '未记录'
  return new Date(value).toLocaleString('zh-CN', {
    month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit',
  })
}
</script>

<template>
  <div class="studio-readiness-flow">
    <!-- 1. Global Metric Band -->
    <div v-if="readiness" class="studio-metric-band">
      <div v-for="item in summaryItems" :key="item.key" class="metric-pill">
        <span class="metric-pill__label">{{ item.label }}</span>
        <span class="metric-pill__value">{{ item.value }}</span>
      </div>
    </div>

    <!-- 2. Decision HUD -->
    <div v-if="readiness" class="decision-hud" :class="readinessDecision.key">
      <div class="decision-main">
        <div class="decision-icon">
          <ShieldCheck v-if="readinessDecision.key === 'ready'" class="h-6 w-6" />
          <AlertTriangle v-else class="h-6 w-6" />
        </div>
        <div class="decision-text">
          <h3 class="decision-title">{{ readinessDecision.title }}</h3>
          <p class="decision-description">{{ readinessDecision.description }}</p>
        </div>
      </div>
      <div class="decision-meta">
        <div class="impact-label">受影响动作</div>
        <div class="impact-tags">
          <span v-for="label in blockingActionLabels" :key="label" class="impact-tag">{{ label }}</span>
          <span v-if="blockingActionLabels.length === 0" class="impact-tag neutral">无阻塞</span>
        </div>
      </div>
    </div>

    <!-- 3. Global Blockers -->
    <section v-if="hasGlobalBlockingReasons" class="global-blockers">
      <header class="section-header">
        <h3 class="section-title">系统级阻塞项</h3>
      </header>
      <div class="blocker-list">
        <div v-for="reason in globalBlockingReasons" :key="reason" class="blocker-item">
          <AlertCircle class="h-4 w-4 text-red-500" />
          <span>{{ getGlobalReasonCopy(reason) }}</span>
        </div>
      </div>
    </section>

    <!-- 4. Challenge Blockers Directory -->
    <section class="challenge-blockers">
      <header class="directory-header">
        <h3 class="directory-title">题目级就绪明细</h3>
        <div class="directory-meta">发现 {{ readiness?.blocking_count ?? 0 }} 个阻塞点</div>
      </header>

      <AppEmpty
        v-if="blockingItems.length === 0"
        title="题目校验通过"
        :description="blockingEmptyDescription"
        icon="ShieldCheck"
        class="py-12"
      />

      <div v-else class="studio-table-wrap">
        <table class="studio-table">
          <thead>
            <tr>
              <th class="col-identity">题目资源</th>
              <th class="col-status">当前状态</th>
              <th class="col-reason">阻塞原因</th>
              <th class="col-meta">最近校验</th>
              <th class="col-actions">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in blockingItems" :key="item.challenge_id" class="studio-row">
              <td class="col-identity">
                <div class="challenge-identity">
                  <div class="challenge-title">{{ item.title }}</div>
                  <div class="challenge-subtitle">{{ item.checker_type === 'http_standard' ? 'HTTP Standard' : '基础探活' }}</div>
                </div>
              </td>
              <td class="col-status">
                <span class="status-pill" :class="item.validation_state">{{ getValidationStateLabel(item) }}</span>
              </td>
              <td class="col-reason">
                <div class="reason-text">{{ getBlockingReasonLabel(item) }}</div>
              </td>
              <td class="col-meta text-[11px] text-slate-500">{{ formatDateTime(item.last_preview_at) }}</td>
              <td class="col-actions">
                <button
                  :id="`awd-readiness-edit-${item.challenge_id}`"
                  class="action-btn"
                  @click="emit('editConfig', item.challenge_id)"
                >
                  {{ props.actionLabel }}
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>
  </div>
</template>

<style scoped>
.studio-readiness-flow { display: flex; flex-direction: column; gap: 2rem; }

/* Metric Band */
.studio-metric-band { display: flex; gap: 0.5rem; background: #f1f5f9; padding: 1rem; border-radius: 1rem; border: 1px solid #e2e8f0; }
.metric-pill { background: white; border: 1px solid #e2e8f0; padding: 0.45rem 1rem; border-radius: 0.75rem; display: flex; align-items: baseline; gap: 0.75rem; }
.metric-pill__label { font-size: 8px; font-weight: 800; text-transform: uppercase; color: #64748b; letter-spacing: 0.05em; }
.metric-pill__value { font-size: 13px; font-weight: 900; color: #1e293b; font-family: var(--font-family-mono); }

/* Decision HUD */
.decision-hud { display: flex; justify-content: space-between; align-items: center; padding: 2rem; border-radius: 1.25rem; }
.decision-hud.ready { background: #f0fdf4; color: #166534; }
.decision-hud.blocked { background: #fef2f2; color: #991b1b; }
.decision-hud.override { background: #fffbeb; color: #92400e; }

.decision-main { display: flex; align-items: center; gap: 1.5rem; }
.decision-icon { width: 3.5rem; height: 3.5rem; border-radius: 1rem; background: white; display: flex; align-items: center; justify-content: center; box-shadow: 0 4px 12px rgba(0,0,0,0.05); }
.decision-title { font-size: 1.15rem; font-weight: 900; margin: 0; }
.decision-description { font-size: 13px; margin-top: 0.25rem; opacity: 0.8; font-weight: 500; }

.decision-meta { text-align: right; }
.impact-label { font-size: 9px; font-weight: 800; text-transform: uppercase; letter-spacing: 0.1em; margin-bottom: 0.5rem; opacity: 0.6; }
.impact-tags { display: flex; gap: 0.35rem; justify-content: flex-end; }
.impact-tag { font-size: 10px; font-weight: 800; padding: 0.15rem 0.6rem; border-radius: 4px; background: rgba(0,0,0,0.05); }

/* Global Blockers */
.global-blockers { background: #fff1f2; border-radius: 1.25rem; padding: 2rem; }
.section-title { font-size: 13px; font-weight: 900; color: #991b1b; text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 1rem; }
.blocker-list { display: flex; flex-direction: column; gap: 0.75rem; }
.blocker-item { display: flex; align-items: center; gap: 0.75rem; font-size: 13px; font-weight: 700; color: #7f1d1d; }

/* Directory */
.directory-header { display: flex; justify-content: space-between; align-items: flex-end; margin-bottom: 1rem; }
.directory-title { font-size: 13px; font-weight: 900; color: #0f172a; text-transform: uppercase; letter-spacing: 0.1em; }
.directory-meta { font-size: 11px; font-weight: 600; color: #94a3b8; }

/* Table Styles */
.studio-table-wrap { border: none; border-radius: 0; background: transparent; overflow: hidden; }
.studio-table { width: 100%; border-collapse: collapse; background: white; }
.studio-table th { background: #f8fafc; padding: 0.75rem 1rem; text-align: left; font-size: 10px; font-weight: 800; text-transform: uppercase; color: #94a3b8; border-bottom: 1px solid #e2e8f0; border-top: 1px solid #e2e8f0; }
.studio-table td { padding: 1rem; border-bottom: 1px solid #f1f5f9; }

.challenge-title { font-size: 14px; font-weight: 800; color: #1e293b; }
.challenge-subtitle { font-size: 11px; color: #94a3b8; margin-top: 0.15rem; }

.status-pill { font-size: 9px; font-weight: 800; padding: 0.15rem 0.5rem; border-radius: 99px; }
.status-pill.passed { background: #dcfce7; color: #166534; }
.status-pill.failed { background: #fee2e2; color: #991b1b; }
.status-pill.pending, .status-pill.stale { background: #fef3c7; color: #92400e; }

.reason-text { font-size: 12px; font-weight: 700; color: #475569; }

.action-btn { font-size: 11px; font-weight: 800; color: #2563eb; background: #eff6ff; padding: 0.35rem 0.75rem; border-radius: 0.5rem; cursor: pointer; transition: all 0.2s ease; border: none; }
.action-btn:hover { background: #dbeafe; }
</style>
