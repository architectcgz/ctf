<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { getCheatDetection } from '@/api/admin'
import type { AdminCheatDetectionData } from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'
import AppLoading from '@/components/common/AppLoading.vue'

type CheatPanelKey = 'overview' | 'suspects' | 'shared-ip' | 'actions'
const validPanelKeys = new Set<CheatPanelKey>(['overview', 'suspects', 'shared-ip', 'actions'])

const route = useRoute()
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

const panelTabs: Array<{ key: CheatPanelKey; label: string; panelId: string; tabId: string }> = [
  { key: 'overview', label: '风险总览', panelId: 'cheat-panel-overview', tabId: 'cheat-tab-overview' },
  { key: 'suspects', label: '高频提交账号', panelId: 'cheat-panel-suspects', tabId: 'cheat-tab-suspects' },
  { key: 'shared-ip', label: '共享 IP 线索', panelId: 'cheat-panel-shared-ip', tabId: 'cheat-tab-shared-ip' },
  { key: 'actions', label: '快速排查入口', panelId: 'cheat-panel-actions', tabId: 'cheat-tab-actions' },
]

const activePanel = computed<CheatPanelKey>(() => {
  const panel = route.query.panel
  if (typeof panel === 'string' && validPanelKeys.has(panel as CheatPanelKey)) {
    return panel as CheatPanelKey
  }
  return 'overview'
})

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

async function switchPanel(panelKey: CheatPanelKey): Promise<void> {
  const nextQuery = { ...route.query, panel: panelKey }
  if (panelKey === 'overview') {
    delete nextQuery.panel
  }
  await router.replace({ name: 'CheatDetection', query: nextQuery })
}

function focusTabByIndex(index: number): void {
  const safeIndex = Math.max(0, Math.min(index, panelTabs.length - 1))
  const targetTab = panelTabs[safeIndex]
  if (!targetTab) return
  document.getElementById(targetTab.tabId)?.focus()
}

function handleTabKeydown(event: KeyboardEvent, index: number): void {
  if (event.key !== 'ArrowRight' && event.key !== 'ArrowLeft' && event.key !== 'Home' && event.key !== 'End') {
    return
  }

  event.preventDefault()

  if (event.key === 'Home') {
    void switchPanel(panelTabs[0].key)
    focusTabByIndex(0)
    return
  }

  if (event.key === 'End') {
    const endIndex = panelTabs.length - 1
    void switchPanel(panelTabs[endIndex].key)
    focusTabByIndex(endIndex)
    return
  }

  const direction = event.key === 'ArrowRight' ? 1 : -1
  const nextIndex = (index + direction + panelTabs.length) % panelTabs.length
  void switchPanel(panelTabs[nextIndex].key)
  focusTabByIndex(nextIndex)
}

onMounted(() => {
  void loadRiskData()
})

const lastSyncText = computed(() => {
  if (!riskData.value?.generated_at) return '近 24 小时'
  return formatDateTime(riskData.value.generated_at)
})

function formatDateTime(value: string): string {
  return new Date(value).toLocaleString('zh-CN')
}
</script>

<template>
  <section class="journal-shell journal-hero flex min-h-full flex-1 flex-col rounded-[30px] border px-6 py-6 md:px-8">
      <header class="workspace-topbar">
        <div class="topbar-leading">
          <span class="workspace-overline">Integrity Workspace</span>
          <span class="class-chip">风险排查</span>
        </div>
        <div class="top-note">
          <span>数据窗口: 近 24 小时</span>
          <span>最后同步: {{ lastSyncText }}</span>
        </div>
      </header>

      <nav class="top-tabs" role="tablist" aria-label="作弊检测视图切换">
        <button
          v-for="(tab, index) in panelTabs"
          :id="tab.tabId"
          :key="tab.tabId"
          type="button"
          role="tab"
          class="top-tab"
          :class="{ active: activePanel === tab.key }"
          :aria-selected="activePanel === tab.key ? 'true' : 'false'"
          :aria-controls="tab.panelId"
          :tabindex="activePanel === tab.key ? 0 : -1"
          @click="switchPanel(tab.key)"
          @keydown="handleTabKeydown($event, index)"
        >
          {{ tab.label }}
        </button>
      </nav>

      <main class="content-pane">
        <div v-if="loading" class="flex justify-center py-10">
          <AppLoading>正在加载风险线索...</AppLoading>
        </div>

        <template v-else-if="riskData">
          <section
            id="cheat-panel-overview"
            class="workspace-hero tab-panel"
            role="tabpanel"
            aria-labelledby="cheat-tab-overview"
            :aria-hidden="activePanel === 'overview' ? 'false' : 'true'"
            v-show="activePanel === 'overview'"
          >
            <div class="overview-grid">
              <div>
                <div class="workspace-overline">Risk Triage</div>
                <h1 class="hero-title">作弊检测</h1>
                <p class="hero-summary">
                  查看高频提交账号、共享 IP 线索和快速排查入口，并继续下钻到审计日志。
                </p>
              </div>

              <article class="journal-brief rounded-[24px] border px-5 py-5">
                <div class="journal-note-label">风险概况</div>
                <div class="mt-5 grid gap-3 sm:grid-cols-2">
                  <div class="journal-note">
                    <div class="journal-note-label">提交突增</div>
                    <div class="journal-note-value">{{ riskData.summary.submit_burst_users }}</div>
                    <div class="journal-note-helper">最近窗口内提交次数异常的账号</div>
                  </div>
                  <div class="journal-note">
                    <div class="journal-note-label">共享 IP</div>
                    <div class="journal-note-value">{{ riskData.summary.shared_ip_groups }}</div>
                    <div class="journal-note-helper">最近 24 小时出现多账号复用的 IP 组</div>
                  </div>
                </div>
              </article>
            </div>

            <div class="journal-divider" />

            <div class="grid gap-3 md:grid-cols-3">
              <div class="journal-note">
                <div class="journal-note-label">Submit Burst</div>
                <div class="journal-note-value">{{ riskData.summary.submit_burst_users }}</div>
                <div class="journal-note-helper">高频提交账号</div>
              </div>
              <div class="journal-note">
                <div class="journal-note-label">Shared IP</div>
                <div class="journal-note-value">{{ riskData.summary.shared_ip_groups }}</div>
                <div class="journal-note-helper">共享 IP 组数</div>
              </div>
              <div class="journal-note">
                <div class="journal-note-label">Affected Users</div>
                <div class="journal-note-value">{{ riskData.summary.affected_users }}</div>
                <div class="journal-note-helper">受影响账号数</div>
              </div>
            </div>

          </section>

        <section
          id="cheat-panel-suspects"
          class="tab-panel space-y-3"
          role="tabpanel"
          aria-labelledby="cheat-tab-suspects"
          :aria-hidden="activePanel === 'suspects' ? 'false' : 'true'"
          v-show="activePanel === 'suspects'"
        >
          <div class="admin-section-head">
            <div>
              <div class="journal-note-label">Burst Accounts</div>
              <h2 class="mt-2 text-xl font-semibold text-[var(--journal-ink)]">高频提交账号</h2>
            </div>
          </div>

          <AppEmpty
            v-if="!riskData?.suspects.length"
            class="cheat-empty-state"
            icon="UsersRound"
            title="当前没有超过阈值的高频提交账号"
            description="说明最近窗口内还没有明显的提交爆发样本。"
          />

          <div v-else class="space-y-3">
            <article
              v-for="suspect in riskData.suspects"
              :key="suspect.user_id"
              class="risk-row"
            >
              <div class="flex items-start justify-between gap-4">
                <div>
                  <p class="font-medium text-[var(--color-text-primary)]">{{ suspect.username }}</p>
                  <p class="mt-1 text-sm text-[var(--color-text-secondary)]">
                    {{ suspect.reason }}
                  </p>
                </div>
                <span
                  class="rounded-full bg-[var(--color-warning)]/10 px-3 py-1 text-xs font-semibold text-[var(--color-warning)]"
                >
                  {{ suspect.submit_count }} 次
                </span>
              </div>
              <p class="mt-3 text-xs text-[var(--color-text-secondary)]">
                最近出现时间：{{ new Date(suspect.last_seen_at).toLocaleString('zh-CN') }}
              </p>
            </article>
          </div>
        </section>

        <section
          id="cheat-panel-shared-ip"
          class="tab-panel space-y-3"
          role="tabpanel"
          aria-labelledby="cheat-tab-shared-ip"
          :aria-hidden="activePanel === 'shared-ip' ? 'false' : 'true'"
          v-show="activePanel === 'shared-ip'"
        >
          <div class="admin-section-head">
            <div>
              <div class="journal-note-label">Shared IP</div>
              <h2 class="mt-2 text-xl font-semibold text-[var(--journal-ink)]">共享 IP 线索</h2>
            </div>
          </div>

          <AppEmpty
            v-if="!riskData?.shared_ips.length"
            class="cheat-empty-state"
            icon="UsersRound"
            title="当前没有共享 IP 线索"
            description="最近 24 小时内还没有发现明显的多账号复用 IP。"
          />

          <div v-else class="space-y-3">
            <article
              v-for="group in riskData.shared_ips"
              :key="group.ip"
              class="risk-row"
            >
              <div class="flex items-start justify-between gap-4">
                <div>
                  <p class="font-mono text-sm text-[var(--color-text-primary)]">{{ group.ip }}</p>
                  <p class="mt-1 text-sm text-[var(--color-text-secondary)]">
                    {{ group.usernames.join('、') || '匿名记录' }}
                  </p>
                </div>
                <span
                  class="rounded-full bg-[var(--color-primary)]/10 px-3 py-1 text-xs font-semibold text-[var(--color-primary)]"
                >
                  {{ group.user_count }} 账号
                </span>
              </div>
            </article>
          </div>
        </section>

        <section
          id="cheat-panel-actions"
          class="tab-panel space-y-3"
          role="tabpanel"
          aria-labelledby="cheat-tab-actions"
          :aria-hidden="activePanel === 'actions' ? 'false' : 'true'"
          v-show="activePanel === 'actions'"
        >
          <div class="admin-section-head">
            <div>
              <div class="journal-note-label">Quick Actions</div>
              <h2 class="mt-2 text-xl font-semibold text-[var(--journal-ink)]">快速排查入口</h2>
            </div>
          </div>

        <div class="grid gap-3 lg:grid-cols-2">
          <button
            v-for="action in quickActions"
            :key="action.title"
            type="button"
            class="quick-action-row"
            @click="openAudit(action.query)"
          >
            <div>
              <p class="font-medium text-[var(--color-text-primary)]">{{ action.title }}</p>
              <p class="mt-1 text-sm leading-6 text-[var(--color-text-secondary)]">
                {{ action.description }}
              </p>
            </div>
            <span class="mt-0.5 text-sm font-medium text-[var(--color-primary)]">打开</span>
          </button>
          </div>
        </section>
        </template>

        <div
          v-else-if="error"
          class="rounded-2xl border border-[var(--color-danger)]/20 bg-[var(--color-danger)]/10 px-5 py-4 text-sm text-[var(--color-danger)]"
        >
          {{ error }}
        </div>

        <div v-else class="admin-empty">
          当前没有风险线索。
        </div>
      </main>
    </section>
</template>

<style scoped>
.journal-shell {
  --journal-ink: var(--color-text-primary);
  --journal-muted: var(--color-text-secondary);
  --journal-accent: var(--color-primary);
  --journal-border: color-mix(in srgb, var(--color-border-default) 84%, transparent);
  --journal-surface: color-mix(in srgb, var(--color-bg-surface) 92%, var(--color-bg-base));
  --journal-surface-subtle: color-mix(in srgb, var(--color-bg-surface) 78%, var(--color-bg-base));
  --cheat-card-border: color-mix(in srgb, var(--journal-border) 74%, transparent);
  --cheat-divider: color-mix(in srgb, var(--journal-border) 68%, transparent);
}

.journal-hero,
.journal-panel {
  border-color: var(--journal-border);
  background:
    radial-gradient(circle at top right, color-mix(in srgb, var(--journal-accent) 12%, transparent), transparent 18rem),
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--journal-surface) 96%, var(--color-bg-base)),
      color-mix(in srgb, var(--journal-surface-subtle) 94%, var(--color-bg-base))
    );
  border-radius: 16px !important;
  box-shadow: 0 18px 40px var(--color-shadow-soft);
}

.journal-brief {
  background: var(--journal-surface-subtle);
  border-color: var(--journal-border);
  border-radius: 16px !important;
}

.journal-eyebrow,
.journal-note-label {
  font-size: 0.7rem;
  font-weight: 700;
  letter-spacing: 0.2em;
  text-transform: uppercase;
  color: var(--journal-accent);
}

.journal-note {
  border-radius: 14px;
  border: 1px solid var(--journal-border);
  background: var(--journal-surface);
  padding: 0.75rem 0.875rem;
}

.journal-note-value {
  margin-top: 0.35rem;
  font-size: 1rem;
  font-weight: 600;
  color: var(--journal-ink);
}

.journal-note-helper {
  margin-top: 0.55rem;
  font-size: 0.78rem;
  line-height: 1.5;
  color: var(--journal-muted);
}

.journal-divider {
  margin-block: 1rem;
  border-top: 1px dashed var(--cheat-divider);
}

.workspace-topbar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem 1rem;
  padding-bottom: 0.85rem;
}

.topbar-leading {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.65rem;
}

.workspace-overline {
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  color: var(--journal-accent);
}

.class-chip {
  display: inline-flex;
  align-items: center;
  min-height: 30px;
  border-radius: 999px;
  border: 1px solid color-mix(in srgb, var(--journal-accent) 26%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 10%, transparent);
  padding: 0.25rem 0.7rem;
  font-size: 0.76rem;
  font-weight: 600;
  color: var(--journal-accent);
}

.top-note {
  display: flex;
  flex-wrap: wrap;
  gap: 0.55rem 1rem;
  font-size: 0.82rem;
  color: var(--journal-muted);
}

.top-tabs {
  display: flex;
  flex-wrap: wrap;
  gap: 0.6rem;
  padding: 0.2rem 0 1rem;
}

.top-tab {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 36px;
  border-radius: 10px;
  border: 1px solid color-mix(in srgb, var(--journal-border) 74%, transparent);
  background: color-mix(in srgb, var(--journal-surface) 92%, var(--color-bg-base));
  padding: 0.5rem 0.9rem;
  font-size: 0.82rem;
  font-weight: 600;
  color: var(--journal-muted);
  transition:
    border-color 0.16s ease,
    background-color 0.16s ease,
    color 0.16s ease;
}

.top-tab:hover {
  border-color: color-mix(in srgb, var(--journal-accent) 30%, var(--journal-border));
  color: var(--journal-ink);
}

.top-tab.active {
  border-color: color-mix(in srgb, var(--journal-accent) 44%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 10%, transparent);
  color: var(--journal-accent);
}

.top-tab:focus-visible {
  outline: 2px solid color-mix(in srgb, var(--journal-accent) 58%, white);
  outline-offset: 2px;
}

.tab-panel {
  min-width: 0;
}

.content-pane {
  display: flex;
  flex: 1 1 auto;
  flex-direction: column;
  gap: 1rem;
}

.workspace-hero {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.overview-grid {
  display: grid;
  gap: 1.5rem;
}

.hero-title {
  margin-top: 0.75rem;
  font-size: clamp(32px, 4vw, 46px);
  line-height: 1.02;
  letter-spacing: -0.04em;
  font-weight: 600;
  color: var(--journal-ink);
}

.hero-summary {
  margin-top: 0.75rem;
  max-width: 48rem;
  font-size: 0.92rem;
  line-height: 1.7;
  color: var(--journal-muted);
}

.admin-section-head {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
}

.risk-row,
.quick-action-row {
  border: 1px solid var(--cheat-card-border);
  border-radius: 18px;
  background: color-mix(in srgb, var(--journal-surface) 94%, transparent);
  padding: 1rem;
}

.cheat-empty-state {
  border-top-color: var(--cheat-divider);
  border-bottom-color: var(--cheat-divider);
}

.admin-empty {
  border: 1px dashed rgba(148, 163, 184, 0.72);
  border-radius: 16px;
  padding: 1rem;
  font-size: 0.875rem;
  color: var(--journal-muted);
}

.quick-action-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  text-align: left;
}

:global([data-theme='dark']) .journal-shell {
  --journal-ink: color-mix(in srgb, var(--color-text-primary) 88%, var(--color-text-secondary));
  --journal-muted: var(--color-text-secondary);
  --journal-accent: #60a5fa;
  --journal-border: color-mix(in srgb, var(--color-border-default) 84%, transparent);
  --journal-surface: color-mix(in srgb, var(--color-bg-surface) 90%, var(--color-bg-base));
  --journal-surface-subtle: color-mix(in srgb, var(--color-bg-surface) 76%, var(--color-bg-base));
}

:global([data-theme='dark']) .journal-hero,
:global([data-theme='dark']) .journal-panel {
  background:
    radial-gradient(circle at top right, color-mix(in srgb, var(--journal-accent) 16%, transparent), transparent 18rem),
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--journal-surface) 97%, var(--color-bg-base)),
      color-mix(in srgb, var(--journal-surface-subtle) 95%, var(--color-bg-base))
    );
}

@media (min-width: 1280px) {
  .overview-grid {
    grid-template-columns: 1.06fr 0.94fr;
    align-items: start;
  }
}

@media (max-width: 720px) {
  .top-tabs {
    gap: 0.45rem;
  }

  .top-tab {
    min-height: 38px;
    padding-inline: 0.8rem;
  }

  .top-note {
    width: 100%;
    flex-direction: column;
    gap: 0.3rem;
  }
}
</style>
