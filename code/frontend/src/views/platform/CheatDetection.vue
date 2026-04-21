<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import {
  AlertCircle,
  Fingerprint,
  RefreshCw,
  SearchCheck,
  ShieldAlert,
  Users,
  ShieldQuestion,
  History,
  ArrowRight,
} from 'lucide-vue-next'

import { getCheatDetection } from '@/api/admin'
import type { AdminCheatDetectionData } from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'
import AppLoading from '@/components/common/AppLoading.vue'

const router = useRouter()
const loading = ref(false)
const error = ref('')
const riskData = ref<AdminCheatDetectionData | null>(null)

const quickActions = [
  {
    title: '查看提交记录',
    description: '直接打开审计日志中的 submit 动作，复核高频提交账号。',
    query: { action: 'submit' },
  },
  {
    title: '查看登录记录',
    description: '回看 login 日志，继续确认共享 IP 或短时集中登录。',
    query: { action: 'login' },
  },
] as const

async function loadRiskData() {
  loading.value = true
  error.value = ''
  try {
    riskData.value = await getCheatDetection()
  } catch (err) {
    console.error(err)
    error.value = '加载作弊检测结果失败，请稍后重试。'
  } finally {
    loading.value = false
  }
}

function openAudit(query: Record<string, string>) {
  return router.push({ name: 'AuditLog', query })
}

function formatDateTime(value: string): string {
  return new Date(value).toLocaleString('zh-CN')
}

onMounted(() => {
  void loadRiskData()
})
</script>

<template>
  <div class="workspace-shell">
    <div class="workspace-grid">
      <main class="content-pane">
        <section class="workspace-hero">
          <div class="workspace-tab-heading__main">
            <div class="workspace-overline">
              Integrity Workspace
            </div>
            <h1 class="hero-title">
              作弊检测
            </h1>
            <p class="hero-summary">
              基于提交爆发、IP 共享及行为指纹的多维度合规分析，维护靶场竞技公平性。
            </p>
          </div>

          <div class="awd-library-hero-actions">
            <div class="quick-actions">
              <div
                v-if="riskData"
                class="hero-meta-badge"
              >
                <span class="hero-meta-badge__label">最近生成</span>
                <span class="hero-meta-badge__value">{{ formatDateTime(riskData.generated_at) }}</span>
              </div>
              <button
                type="button"
                class="ui-btn ui-btn--ghost"
                @click="openAudit({})"
              >
                <SearchCheck class="h-4 w-4" />
                打开审计日志
              </button>
              <button
                type="button"
                class="ui-btn ui-btn--primary"
                @click="loadRiskData"
              >
                <RefreshCw class="h-4 w-4" :class="{ 'animate-spin': loading }" />
                刷新线索
              </button>
            </div>
          </div>
        </section>

        <div
          v-if="loading && !riskData"
          class="flex justify-center py-20"
        >
          <AppLoading>正在扫描合规风险...</AppLoading>
        </div>

        <div
          v-else-if="riskData"
          class="cheat-detection-body mt-10 space-y-10"
        >
          <div class="metric-panel-grid metric-panel-grid--premium cols-3">
            <article class="metric-panel-card metric-panel-card--premium">
              <div class="metric-panel-label">
                <span>提交风险账号</span>
                <ShieldAlert class="h-4 w-4" />
              </div>
              <div class="metric-panel-value">
                {{ riskData.summary.submit_burst_users.toString().padStart(2, '0') }}
              </div>
              <div class="metric-panel-helper">
                检出高频提交爆发
              </div>
            </article>

            <article class="metric-panel-card metric-panel-card--premium">
              <div class="metric-panel-label">
                <span>多账号共享 IP</span>
                <Fingerprint class="h-4 w-4" />
              </div>
              <div class="metric-panel-value">
                {{ riskData.summary.shared_ip_groups.toString().padStart(2, '0') }}
              </div>
              <div class="metric-panel-helper">
                疑似团队线下协作
              </div>
            </article>

            <article class="metric-panel-card metric-panel-card--premium">
              <div class="metric-panel-label">
                <span>涉及用户总数</span>
                <Users class="h-4 w-4" />
              </div>
              <div class="metric-panel-value">
                {{ riskData.summary.affected_users.toString().padStart(2, '0') }}
              </div>
              <div class="metric-panel-helper">
                全站风险波及范围
              </div>
            </article>
          </div>

          <!-- Burst Accounts -->
          <section class="workspace-directory-section">
            <header class="list-heading">
              <div>
                <div class="workspace-overline">
                  Compliance Risk / Burst
                </div>
                <h2 class="list-heading__title">
                  高频提交风险线索
                </h2>
              </div>
            </header>

            <AppEmpty
              v-if="!riskData.suspects.length"
              icon="ShieldCheck"
              title="当前无爆发性提交风险"
              description="说明最近统计窗口内还没有明显的提交样本超过安全阈值。"
              class="py-12"
            />

            <div
              v-else
              class="cheat-list"
            >
              <button
                v-for="suspect in riskData.suspects"
                :key="suspect.user_id"
                class="cheat-row group"
                @click="openAudit({ action: 'submit', actor_user_id: suspect.user_id.toString() })"
              >
                <div class="cheat-row__main">
                  <div class="cheat-row__title">
                    {{ suspect.username }}
                  </div>
                  <div class="cheat-row__reason">
                    {{ suspect.reason }}
                  </div>
                </div>
                <div class="cheat-row__meta">
                  <div class="cheat-badge cheat-badge--warning">
                    {{ suspect.submit_count }} 提交
                  </div>
                  <div class="cheat-time">
                    最近出现 {{ formatDateTime(suspect.last_seen_at) }}
                  </div>
                  <div class="cheat-action">
                    <span>审计复核</span>
                    <ArrowRight class="h-3 w-3" />
                  </div>
                </div>
              </button>
            </div>
          </section>

          <!-- Shared IP -->
          <section class="workspace-directory-section">
            <header class="list-heading">
              <div>
                <div class="workspace-overline">
                  Compliance Risk / Network
                </div>
                <h2 class="list-heading__title">
                  共享 IP 复用线索
                </h2>
              </div>
            </header>

            <AppEmpty
              v-if="!riskData.shared_ips.length"
              icon="ShieldCheck"
              title="未发现 IP 共享行为"
              description="最近 24 小时内未监测到不同账号从同一公网地址密集登录。"
              class="py-12"
            />

            <div
              v-else
              class="cheat-list"
            >
              <button
                v-for="group in riskData.shared_ips"
                :key="group.ip"
                class="cheat-row group"
                @click="openAudit({ action: 'login' })"
              >
                <div class="cheat-row__main">
                  <div class="cheat-row__title font-mono">
                    {{ group.ip }}
                  </div>
                  <div class="cheat-row__reason">
                    涉及账号: {{ group.usernames.join('、') }}
                  </div>
                </div>
                <div class="cheat-row__meta">
                  <div class="cheat-badge">
                    {{ group.user_count }} 账号
                  </div>
                  <div class="cheat-time">
                    多见于短时集中登录行为
                  </div>
                  <div class="cheat-action">
                    <span>追踪登录</span>
                    <ArrowRight class="h-3 w-3" />
                  </div>
                </div>
              </button>
            </div>
          </section>

          <!-- Quick Actions -->
          <section class="workspace-directory-section">
            <header class="list-heading">
              <div>
                <div class="workspace-overline">
                  Analysis Shortcuts
                </div>
                <h2 class="list-heading__title">
                  常用审计联动
                </h2>
              </div>
            </header>

            <div class="cheat-list">
              <button
                v-for="action in quickActions"
                :key="action.title"
                class="cheat-row group"
                @click="openAudit(action.query)"
              >
                <div class="cheat-row__main">
                  <div class="cheat-row__title">
                    {{ action.title }}
                  </div>
                  <div class="cheat-row__reason">
                    {{ action.description }}
                  </div>
                </div>
                <div class="cheat-row__meta">
                  <div class="cheat-badge cheat-badge--muted">
                    跳转入口
                  </div>
                  <div class="cheat-action">
                    <span>打开详情</span>
                    <ArrowRight class="h-3 w-3" />
                  </div>
                </div>
              </button>
            </div>
          </section>
        </div>

        <div
          v-else-if="error"
          class="cheat-error-box"
        >
          <AlertCircle class="h-4 w-4" />
          <span>{{ error }}</span>
          <button @click="loadRiskData" class="underline ml-2">重试</button>
        </div>

        <div
          v-else
          class="py-20 flex flex-col items-center gap-4 opacity-30"
        >
          <ShieldQuestion class="h-12 w-12" />
          <p class="font-bold">当前没有任何风险检出</p>
        </div>
      </main>
    </div>
  </div>
</template>

<style scoped>
.workspace-hero {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: var(--space-7);
  padding-bottom: var(--space-6);
  border-bottom: 1px solid var(--color-border-subtle);
}

.hero-title {
  margin: 0.5rem 0 0;
  font-size: var(--workspace-page-title-font-size);
  line-height: var(--workspace-page-title-line-height);
  letter-spacing: var(--workspace-page-title-letter-spacing);
  color: var(--color-text-primary);
}

.hero-summary {
  max-width: 760px;
  margin-top: var(--space-3-5);
  font-size: var(--font-size-15);
  line-height: 1.9;
  color: var(--color-text-secondary);
}

.hero-meta-badge {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: var(--space-1);
}

.hero-meta-badge__label {
  font-size: var(--font-size-10);
  font-weight: 800;
  text-transform: uppercase;
  color: var(--color-text-muted);
  letter-spacing: 0.1em;
}

.hero-meta-badge__value {
  font-size: var(--font-size-13);
  font-weight: 700;
  color: var(--color-text-primary);
}

.awd-library-hero-actions {
  display: flex;
  align-items: flex-end;
  padding-bottom: 0.5rem;
}

.quick-actions {
  display: flex;
  gap: var(--space-5);
  align-items: center;
}

.cheat-list {
  display: flex;
  flex-direction: column;
  border: 1px solid var(--color-border-default);
  border-radius: 1.25rem;
  background: var(--color-bg-surface);
  overflow: hidden;
}

.cheat-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-6);
  padding: 1.5rem 2rem;
  background: transparent;
  border: none;
  border-bottom: 1px solid var(--color-border-subtle);
  width: 100%;
  text-align: left;
  transition: all 0.2s ease;
  cursor: pointer;
}

.cheat-row:last-child {
  border-bottom: none;
}

.cheat-row:hover {
  background: var(--color-bg-elevated);
}

.cheat-row__main {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
  min-width: 0;
}

.cheat-row__title {
  font-size: var(--font-size-15);
  font-weight: 800;
  color: var(--color-text-primary);
  transition: color 0.2s ease;
}

.group:hover .cheat-row__title {
  color: var(--color-primary);
}

.cheat-row__reason {
  font-size: var(--font-size-13);
  color: var(--color-text-secondary);
  line-height: 1.6;
}

.cheat-row__meta {
  display: flex;
  align-items: center;
  gap: var(--space-6);
}

.cheat-badge {
  font-size: var(--font-size-11);
  font-weight: 800;
  padding: 0.25rem 0.75rem;
  border-radius: 99px;
  background: var(--color-bg-elevated);
  color: var(--color-text-secondary);
  white-space: nowrap;
}

.cheat-badge--warning {
  background: var(--color-warning-soft);
  color: var(--color-warning);
}

.cheat-badge--muted {
  opacity: 0.6;
}

.cheat-time {
  font-size: var(--font-size-12);
  color: var(--color-text-muted);
  white-space: nowrap;
}

.cheat-action {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: var(--font-size-12);
  font-weight: 700;
  color: var(--color-primary);
  opacity: 0;
  transform: translateX(-10px);
  transition: all 0.2s ease;
}

.group:hover .cheat-action {
  opacity: 1;
  transform: translateX(0);
}

.cheat-error-box {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 1rem 1.25rem;
  background: var(--color-danger-soft);
  color: var(--color-danger);
  border-radius: 1rem;
  font-size: var(--font-size-13);
  font-weight: 600;
}

@media (max-width: 1024px) {
  .workspace-hero { grid-template-columns: 1fr; }
  .awd-library-hero-actions { padding: 0; }
  .cheat-row { flex-direction: column; align-items: flex-start; gap: var(--space-4); }
  .cheat-row__meta { width: 100%; justify-content: space-between; }
  .cheat-action { opacity: 1; transform: none; }
}
</style>