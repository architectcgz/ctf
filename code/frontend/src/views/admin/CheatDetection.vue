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
  routeName: 'CheatDetection',
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

function formatDateTime(value: string): string {
  return new Date(value).toLocaleString('zh-CN')
}

onMounted(() => {
  void loadRiskData()
})
</script>

<template>
  <section
    class="workspace-shell journal-shell journal-shell-admin journal-notes-card journal-hero flex min-h-full flex-1 flex-col rounded-[30px] border px-6 py-6 md:px-8"
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
          class="tab-panel"
          :class="{ active: activePanel === 'overview' }"
          role="tabpanel"
          aria-labelledby="cheat-tab-overview"
          :aria-hidden="activePanel === 'overview' ? 'false' : 'true'"
          v-show="activePanel === 'overview'"
        >
          <header class="cheat-overview-head">
            <div class="workspace-tab-heading__main">
              <div class="workspace-overline">Risk Triage</div>
              <h1 class="workspace-page-title">作弊检测</h1>
              <p class="workspace-page-copy">
                先看高频提交和共享 IP 的风险焦点，再切到账号、线索和审计入口继续复核。
              </p>
            </div>

            <div class="cheat-overview-meta">
              <span class="cheat-overview-meta__label">最近生成</span>
              <span class="cheat-overview-meta__value">
                {{ formatDateTime(riskData.generated_at) }}
              </span>
            </div>
          </header>

          <div
            class="admin-summary-grid cheat-kpi-summary progress-strip metric-panel-grid metric-panel-default-surface metric-panel-workspace-surface"
          >
            <article class="journal-note progress-card metric-panel-card">
              <div class="journal-note-label progress-card-label metric-panel-label">
                Submit Burst
              </div>
              <div class="journal-note-value progress-card-value metric-panel-value">
                {{ riskData.summary.submit_burst_users }}
              </div>
              <div class="journal-note-helper progress-card-hint metric-panel-helper">
                高频提交账号
              </div>
            </article>
            <article class="journal-note progress-card metric-panel-card">
              <div class="journal-note-label progress-card-label metric-panel-label">Shared IP</div>
              <div class="journal-note-value progress-card-value metric-panel-value">
                {{ riskData.summary.shared_ip_groups }}
              </div>
              <div class="journal-note-helper progress-card-hint metric-panel-helper">
                共享 IP 组数
              </div>
            </article>
            <article class="journal-note progress-card metric-panel-card">
              <div class="journal-note-label progress-card-label metric-panel-label">
                Affected Users
              </div>
              <div class="journal-note-value progress-card-value metric-panel-value">
                {{ riskData.summary.affected_users }}
              </div>
              <div class="journal-note-helper progress-card-hint metric-panel-helper">
                受影响账号数
              </div>
            </article>
          </div>

          <div class="journal-divider" />

          <section class="workspace-directory-section cheat-directory-section">
            <header class="list-heading">
              <div>
                <div class="journal-note-label">Risk Focus</div>
                <h2 class="list-heading__title">风险焦点</h2>
              </div>
              <div class="cheat-directory-caption">按优先级切换到对应排查面板</div>
            </header>

            <div class="cheat-directory-list">
              <button type="button" class="cheat-directory-row" @click="switchPanel('suspects')">
                <div class="cheat-directory-row-main">
                  <h3 class="cheat-directory-row-title">高频提交账号</h3>
                  <p class="cheat-directory-row-copy">
                    先复核短时间内提交异常偏高的账号，再决定是否继续追踪其登录和操作轨迹。
                  </p>
                </div>
                <div class="cheat-directory-row-meta">
                  <span class="cheat-directory-row-chip cheat-directory-row-chip-warning">
                    {{ riskData.summary.submit_burst_users }} 个账号
                  </span>
                  <span class="cheat-directory-row-cta">查看账号</span>
                </div>
              </button>

              <button type="button" class="cheat-directory-row" @click="switchPanel('shared-ip')">
                <div class="cheat-directory-row-main">
                  <h3 class="cheat-directory-row-title">共享 IP 线索</h3>
                  <p class="cheat-directory-row-copy">
                    汇总最近 24 小时内多账号复用的 IP 组，适合先做登录轨迹和时间段交叉复核。
                  </p>
                </div>
                <div class="cheat-directory-row-meta">
                  <span class="cheat-directory-row-chip"
                    >{{ riskData.summary.shared_ip_groups }} 组</span
                  >
                  <span class="cheat-directory-row-cta">查看线索</span>
                </div>
              </button>

              <button
                type="button"
                class="cheat-directory-row"
                @click="openAudit({ action: 'submit' })"
              >
                <div class="cheat-directory-row-main">
                  <h3 class="cheat-directory-row-title">提交审计流水</h3>
                  <p class="cheat-directory-row-copy">
                    直接跳到 submit 日志，结合账号和时间窗口继续核对风险样本是否真实异常。
                  </p>
                </div>
                <div class="cheat-directory-row-meta">
                  <span class="cheat-directory-row-chip cheat-directory-row-chip-muted">
                    审计联动
                  </span>
                  <span class="cheat-directory-row-cta">打开日志</span>
                </div>
              </button>
            </div>
          </section>

          <section class="workspace-directory-section cheat-directory-section">
            <header class="list-heading">
              <div>
                <div class="journal-note-label">Review Actions</div>
                <h2 class="list-heading__title">复核动作</h2>
              </div>
              <div class="cheat-directory-caption">常用排查入口按动作整理</div>
            </header>

            <div class="quick-action-directory">
              <button
                v-for="action in quickActions"
                :key="action.title"
                type="button"
                class="quick-action-row"
                @click="openAudit(action.query)"
              >
                <div class="cheat-directory-row-main">
                  <h3 class="cheat-directory-row-title">{{ action.title }}</h3>
                  <p class="cheat-directory-row-copy">
                    {{ action.description }}
                  </p>
                </div>
                <div class="cheat-directory-row-meta">
                  <span class="cheat-directory-row-chip cheat-directory-row-chip-muted">
                    Audit
                  </span>
                  <span class="cheat-directory-row-cta">立即排查</span>
                </div>
              </button>
            </div>
          </section>
        </section>

        <section
          id="cheat-panel-suspects"
          class="tab-panel"
          :class="{ active: activePanel === 'suspects' }"
          role="tabpanel"
          aria-labelledby="cheat-tab-suspects"
          :aria-hidden="activePanel === 'suspects' ? 'false' : 'true'"
          v-show="activePanel === 'suspects'"
        >
          <section class="workspace-directory-section cheat-directory-section">
            <header class="list-heading">
              <div>
                <div class="journal-note-label">Burst Accounts</div>
                <h2 class="list-heading__title">高频提交账号</h2>
              </div>
              <div class="cheat-directory-caption">按账号排序查看提交频次和最近出现时间</div>
            </header>

            <AppEmpty
              v-if="!riskData?.suspects.length"
              class="cheat-empty-state"
              icon="UsersRound"
              title="当前没有超过阈值的高频提交账号"
              description="说明最近窗口内还没有明显的提交爆发样本。"
            />

            <div v-else class="cheat-directory-list">
              <button
                v-for="suspect in riskData.suspects"
                :key="suspect.user_id"
                type="button"
                class="cheat-directory-row"
                @click="openAudit({ action: 'submit', actor_user_id: suspect.user_id })"
              >
                <div class="cheat-directory-row-main">
                  <h3 class="cheat-directory-row-title">{{ suspect.username }}</h3>
                  <p class="cheat-directory-row-copy">{{ suspect.reason }}</p>
                </div>
                <div class="cheat-directory-row-meta">
                  <span class="cheat-directory-row-chip cheat-directory-row-chip-warning">
                    {{ suspect.submit_count }} 次
                  </span>
                  <span class="cheat-directory-row-subtle">
                    最近出现 {{ formatDateTime(suspect.last_seen_at) }}
                  </span>
                  <span class="cheat-directory-row-cta">查看提交日志</span>
                </div>
              </button>
            </div>
          </section>
        </section>

        <section
          id="cheat-panel-shared-ip"
          class="tab-panel"
          :class="{ active: activePanel === 'shared-ip' }"
          role="tabpanel"
          aria-labelledby="cheat-tab-shared-ip"
          :aria-hidden="activePanel === 'shared-ip' ? 'false' : 'true'"
          v-show="activePanel === 'shared-ip'"
        >
          <section class="workspace-directory-section cheat-directory-section">
            <header class="list-heading">
              <div>
                <div class="journal-note-label">Shared IP</div>
                <h2 class="list-heading__title">共享 IP 线索</h2>
              </div>
              <div class="cheat-directory-caption">按 IP 聚合查看复用账号范围</div>
            </header>

            <AppEmpty
              v-if="!riskData?.shared_ips.length"
              class="cheat-empty-state"
              icon="UsersRound"
              title="当前没有共享 IP 线索"
              description="最近 24 小时内还没有发现明显的多账号复用 IP。"
            />

            <div v-else class="cheat-directory-list">
              <button
                v-for="group in riskData.shared_ips"
                :key="group.ip"
                type="button"
                class="cheat-directory-row"
                @click="openAudit({ action: 'login' })"
              >
                <div class="cheat-directory-row-main">
                  <h3 class="cheat-directory-row-title cheat-directory-row-title-mono">
                    {{ group.ip }}
                  </h3>
                  <p class="cheat-directory-row-copy">
                    {{ group.usernames.join('、') || '匿名记录' }}
                  </p>
                </div>
                <div class="cheat-directory-row-meta">
                  <span class="cheat-directory-row-chip">{{ group.user_count }} 账号</span>
                  <span class="cheat-directory-row-subtle">建议先复核登录时间段</span>
                  <span class="cheat-directory-row-cta">查看登录日志</span>
                </div>
              </button>
            </div>
          </section>
        </section>

        <section
          id="cheat-panel-actions"
          class="tab-panel"
          :class="{ active: activePanel === 'actions' }"
          role="tabpanel"
          aria-labelledby="cheat-tab-actions"
          :aria-hidden="activePanel === 'actions' ? 'false' : 'true'"
          v-show="activePanel === 'actions'"
        >
          <section class="workspace-directory-section cheat-directory-section">
            <header class="list-heading">
              <div>
                <div class="journal-note-label">Quick Actions</div>
                <h2 class="list-heading__title">快速排查入口</h2>
              </div>
              <div class="cheat-directory-caption">直接跳转到审计日志的高频动作视图</div>
            </header>

            <div class="quick-action-directory">
              <button
                v-for="action in quickActions"
                :key="action.title"
                type="button"
                class="quick-action-row"
                @click="openAudit(action.query)"
              >
                <div class="cheat-directory-row-main">
                  <h3 class="cheat-directory-row-title">{{ action.title }}</h3>
                  <p class="cheat-directory-row-copy">
                    {{ action.description }}
                  </p>
                </div>
                <div class="cheat-directory-row-meta">
                  <span class="cheat-directory-row-chip cheat-directory-row-chip-muted">
                    审计联动
                  </span>
                  <span class="cheat-directory-row-cta">打开日志</span>
                </div>
              </button>
            </div>
          </section>
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
  --workspace-brand: var(--journal-accent);
  --workspace-brand-ink: color-mix(in srgb, var(--journal-accent) 74%, var(--journal-ink));
  --workspace-brand-soft: color-mix(in srgb, var(--journal-accent) 10%, transparent);
  --workspace-panel: color-mix(in srgb, var(--color-bg-surface) 90%, var(--color-bg-base));
  --workspace-panel-soft: color-mix(in srgb, var(--color-bg-surface) 82%, var(--color-bg-base));
  --workspace-line-soft: color-mix(in srgb, var(--color-text-primary) 10%, transparent);
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

.content-pane {
  display: flex;
  flex: 1 1 auto;
  flex-direction: column;
  gap: var(--space-4);
}

.cheat-overview-head {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-end;
  justify-content: space-between;
  gap: var(--space-5);
  margin-bottom: var(--space-4);
}

.cheat-overview-meta {
  display: grid;
  gap: var(--space-1);
  justify-items: end;
  min-width: 12rem;
}

.cheat-overview-meta__label {
  font-size: var(--font-size-0-72);
  font-weight: 700;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.cheat-overview-meta__value {
  font-size: var(--font-size-0-88);
  color: var(--journal-ink);
}

.list-heading {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-end;
  justify-content: space-between;
  gap: var(--space-3);
}

.list-heading__title {
  margin: var(--space-1) 0 0;
  font-size: var(--font-size-1-20);
  font-weight: 700;
  color: var(--journal-ink);
}

.cheat-risk-summary {
  --admin-summary-grid-columns: repeat(2, minmax(0, 1fr));
}

.cheat-kpi-summary {
  --admin-summary-grid-columns: repeat(3, minmax(0, 1fr));
}

.cheat-directory-section + .cheat-directory-section {
  margin-top: var(--space-5);
}

.cheat-directory-caption {
  font-size: var(--font-size-0-82);
  color: var(--journal-muted);
}

.cheat-directory-list,
.quick-action-directory {
  display: flex;
  flex-direction: column;
}

.cheat-directory-row,
.quick-action-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-4);
  width: 100%;
  border: 1px solid var(--cheat-card-border);
  border-radius: 18px;
  background: color-mix(in srgb, var(--journal-surface) 94%, transparent);
  padding: var(--space-4);
  text-align: left;
  transition:
    border-color 160ms ease,
    background 160ms ease,
    transform 160ms ease;
}

.cheat-directory-row + .cheat-directory-row,
.quick-action-row + .quick-action-row {
  margin-top: var(--space-3);
}

.cheat-directory-row:hover,
.cheat-directory-row:focus-visible,
.quick-action-row:hover,
.quick-action-row:focus-visible {
  border-color: color-mix(in srgb, var(--journal-accent) 24%, var(--cheat-card-border));
  background: color-mix(in srgb, var(--journal-accent) 4%, var(--journal-surface));
  outline: none;
  transform: translateY(-1px);
}

.cheat-directory-row-main {
  display: grid;
  gap: var(--space-1-5);
  min-width: 0;
}

.cheat-directory-row-title {
  margin: 0;
  font-size: var(--font-size-0-98);
  font-weight: 700;
  color: var(--journal-ink);
}

.cheat-directory-row-title-mono {
  font-family: var(--font-family-mono);
}

.cheat-directory-row-copy {
  margin: 0;
  font-size: var(--font-size-0-84);
  line-height: 1.65;
  color: var(--journal-muted);
}

.cheat-directory-row-meta {
  display: grid;
  gap: var(--space-1-5);
  justify-items: end;
  flex-shrink: 0;
}

.cheat-directory-row-chip {
  display: inline-flex;
  align-items: center;
  min-height: 1.75rem;
  padding: 0 var(--space-2-5);
  border-radius: 999px;
  background: color-mix(in srgb, var(--journal-accent) 10%, transparent);
  font-size: var(--font-size-0-74);
  font-weight: 700;
  color: var(--journal-accent-strong);
}

.cheat-directory-row-chip-warning {
  background: color-mix(in srgb, var(--color-warning) 12%, transparent);
  color: var(--color-warning);
}

.cheat-directory-row-chip-muted {
  background: color-mix(in srgb, var(--journal-muted) 10%, transparent);
  color: var(--journal-muted);
}

.cheat-directory-row-subtle {
  font-size: var(--font-size-0-78);
  color: var(--journal-muted);
}

.cheat-directory-row-cta {
  font-size: var(--font-size-0-80);
  font-weight: 700;
  color: var(--journal-accent-strong);
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

@media (max-width: 720px) {
  .cheat-overview-head,
  .list-heading {
    align-items: flex-start;
    flex-direction: column;
  }

  .cheat-kpi-summary {
    --admin-summary-grid-columns: 1fr;
  }

  .cheat-overview-meta,
  .cheat-directory-row-meta {
    justify-items: start;
  }

  .cheat-directory-row,
  .quick-action-row {
    align-items: flex-start;
    flex-direction: column;
  }

  .top-tabs {
    gap: var(--space-5-5);
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
