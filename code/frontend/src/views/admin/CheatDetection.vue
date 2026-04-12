<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { getCheatDetection } from '@/api/admin'
import type { AdminCheatDetectionData } from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'
import AppLoading from '@/components/common/AppLoading.vue'
import { useRouteQueryTabs } from '@/composables/useRouteQueryTabs'

type CheatPanelKey = 'overview' | 'suspects' | 'shared-ip' | 'actions'

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
  {
    key: 'overview',
    label: '风险总览',
    panelId: 'cheat-panel-overview',
    tabId: 'cheat-tab-overview',
  },
  {
    key: 'suspects',
    label: '高频提交账号',
    panelId: 'cheat-panel-suspects',
    tabId: 'cheat-tab-suspects',
  },
  {
    key: 'shared-ip',
    label: '共享 IP 线索',
    panelId: 'cheat-panel-shared-ip',
    tabId: 'cheat-tab-shared-ip',
  },
  {
    key: 'actions',
    label: '快速排查入口',
    panelId: 'cheat-panel-actions',
    tabId: 'cheat-tab-actions',
  },
]
const panelTabOrder = panelTabs.map((tab) => tab.key) as CheatPanelKey[]
const {
  activeTab: activePanel,
  setTabButtonRef,
  selectTab: switchPanel,
  handleTabKeydown,
} = useRouteQueryTabs<CheatPanelKey>({
  route,
  router,
  orderedTabs: panelTabOrder,
  defaultTab: 'overview',
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

onMounted(() => {
  void loadRiskData()
})
</script>

<template>
  <section
    class="journal-shell journal-shell-admin journal-notes-card journal-hero flex min-h-full flex-1 flex-col rounded-[30px] border px-6 py-6 md:px-8"
  >
    <header class="workspace-topbar">
      <div class="topbar-leading">
        <span class="workspace-overline">Integrity Workspace</span>
        <span class="class-chip">风险排查</span>
      </div>
    </header>

    <nav class="top-tabs" role="tablist" aria-label="作弊检测视图切换">
      <button
        v-for="(tab, index) in panelTabs"
        :id="tab.tabId"
        :key="tab.tabId"
        :ref="(element) => setTabButtonRef(tab.key, element as HTMLButtonElement | null)"
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
              <div class="admin-summary-grid cheat-risk-summary mt-5 progress-strip metric-panel-grid metric-panel-default-surface">
                <div class="journal-note progress-card metric-panel-card">
                  <div class="journal-note-label progress-card-label metric-panel-label">提交突增</div>
                  <div class="journal-note-value progress-card-value metric-panel-value">
                    {{ riskData.summary.submit_burst_users }}
                  </div>
                  <div class="journal-note-helper progress-card-hint metric-panel-helper">
                    最近窗口内提交次数异常的账号
                  </div>
                </div>
                <div class="journal-note progress-card metric-panel-card">
                  <div class="journal-note-label progress-card-label metric-panel-label">共享 IP</div>
                  <div class="journal-note-value progress-card-value metric-panel-value">
                    {{ riskData.summary.shared_ip_groups }}
                  </div>
                  <div class="journal-note-helper progress-card-hint metric-panel-helper">
                    最近 24 小时出现多账号复用的 IP 组
                  </div>
                </div>
              </div>
            </article>
          </div>

          <div class="journal-divider" />

          <div class="admin-summary-grid cheat-kpi-summary progress-strip metric-panel-grid metric-panel-default-surface">
            <div class="journal-note progress-card metric-panel-card">
              <div class="journal-note-label progress-card-label metric-panel-label">Submit Burst</div>
              <div class="journal-note-value progress-card-value metric-panel-value">
                {{ riskData.summary.submit_burst_users }}
              </div>
              <div class="journal-note-helper progress-card-hint metric-panel-helper">
                高频提交账号
              </div>
            </div>
            <div class="journal-note progress-card metric-panel-card">
              <div class="journal-note-label progress-card-label metric-panel-label">Shared IP</div>
              <div class="journal-note-value progress-card-value metric-panel-value">
                {{ riskData.summary.shared_ip_groups }}
              </div>
              <div class="journal-note-helper progress-card-hint metric-panel-helper">
                共享 IP 组数
              </div>
            </div>
            <div class="journal-note progress-card metric-panel-card">
              <div class="journal-note-label progress-card-label metric-panel-label">
                Affected Users
              </div>
              <div class="journal-note-value progress-card-value metric-panel-value">
                {{ riskData.summary.affected_users }}
              </div>
              <div class="journal-note-helper progress-card-hint metric-panel-helper">
                受影响账号数
              </div>
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
          <div class="workspace-tab-heading">
            <div class="workspace-tab-heading__main">
              <div class="journal-note-label">Burst Accounts</div>
              <h2 class="workspace-tab-heading__title">高频提交账号</h2>
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
            <article v-for="suspect in riskData.suspects" :key="suspect.user_id" class="risk-row">
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
          <div class="workspace-tab-heading">
            <div class="workspace-tab-heading__main">
              <div class="journal-note-label">Shared IP</div>
              <h2 class="workspace-tab-heading__title">共享 IP 线索</h2>
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
            <article v-for="group in riskData.shared_ips" :key="group.ip" class="risk-row">
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
          <div class="workspace-tab-heading">
            <div class="workspace-tab-heading__main">
              <div class="journal-note-label">Quick Actions</div>
              <h2 class="workspace-tab-heading__title">快速排查入口</h2>
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

      <div v-else class="admin-empty">当前没有风险线索。</div>
    </main>
  </section>
</template>

<style scoped>
.journal-shell {
  --cheat-card-border: color-mix(in srgb, var(--journal-border) 74%, transparent);
  --cheat-divider: color-mix(in srgb, var(--journal-border) 68%, transparent);
  --journal-topbar-padding-bottom: var(--space-3-5);
  --journal-overline-font-size: var(--font-size-0-72);
  --journal-overline-letter-spacing: 0.18em;
  --page-top-tabs-gap: 28px;
  --page-top-tabs-margin: var(--space-2-5) calc(var(--space-6) * -1) 0;
  --page-top-tabs-padding: 0 var(--space-6);
  --page-top-tabs-border: color-mix(in srgb, var(--journal-ink) 10%, transparent);
  --page-top-tab-min-height: 52px;
  --page-top-tab-padding: var(--space-2-5) 0 var(--space-3-5);
  --page-top-tab-font-size: var(--font-size-15);
  --page-top-tab-active-color: color-mix(in srgb, var(--journal-accent) 74%, var(--journal-ink));
  --page-top-tab-active-border: color-mix(in srgb, var(--journal-accent) 86%, var(--journal-ink));
  --journal-divider-border: 1px dashed var(--cheat-divider);
  --journal-shell-dark-accent: var(--color-primary-hover);
}

.journal-brief {
  background: var(--journal-surface-subtle);
  border-color: var(--journal-border);
  border-radius: 16px !important;
}

.content-pane {
  display: flex;
  flex: 1 1 auto;
  flex-direction: column;
  gap: var(--space-4);
}

.workspace-hero {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}

.cheat-risk-summary {
  --admin-summary-grid-columns: repeat(2, minmax(0, 1fr));
}

.cheat-kpi-summary {
  --admin-summary-grid-columns: repeat(3, minmax(0, 1fr));
}

.overview-grid {
  display: grid;
  gap: var(--space-6);
}

.hero-summary {
  margin-top: var(--space-3);
  max-width: 48rem;
  font-size: var(--font-size-0-92);
  line-height: 1.7;
  color: var(--journal-muted);
}

.risk-row,
.quick-action-row {
  border: 1px solid var(--cheat-card-border);
  border-radius: 18px;
  background: color-mix(in srgb, var(--journal-surface) 94%, transparent);
  padding: var(--space-4);
}

.cheat-empty-state {
  border-top-color: var(--cheat-divider);
  border-bottom-color: var(--cheat-divider);
}

.admin-empty {
  border: 1px dashed color-mix(in srgb, var(--journal-border) 72%, transparent);
  border-radius: 16px;
  padding: var(--space-4);
  font-size: var(--font-size-0-875);
  color: var(--journal-muted);
}

.quick-action-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
  text-align: left;
}

@media (min-width: 1280px) {
  .overview-grid {
    grid-template-columns: 1.06fr 0.94fr;
    align-items: start;
  }
}

@media (max-width: 720px) {
  .cheat-risk-summary,
  .cheat-kpi-summary {
    --admin-summary-grid-columns: 1fr;
  }

  .top-tabs {
    gap: var(--space-5-5);
  }

  .top-note {
    width: 100%;
    flex-direction: column;
    gap: var(--space-1-5);
  }
}

@media (min-width: 768px) {
  .top-tabs {
    margin-left: calc(var(--space-8) * -1);
    margin-right: calc(var(--space-8) * -1);
    padding-left: var(--space-8);
    padding-right: var(--space-8);
  }
}
</style>
