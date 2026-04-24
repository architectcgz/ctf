<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import {
  AlertCircle,
  ArrowRight,
  ShieldQuestion,
} from 'lucide-vue-next'

import { getCheatDetection } from '@/api/admin'
import type { AdminCheatDetectionData } from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'
import AppLoading from '@/components/common/AppLoading.vue'
import CheatDetectionHeroPanel from '@/components/platform/cheat/CheatDetectionHeroPanel.vue'
import CheatDetectionSummaryPanel from '@/components/platform/cheat/CheatDetectionSummaryPanel.vue'

const router = useRouter()
const loading = ref(false)
const error = ref('')
const riskData = ref<AdminCheatDetectionData | null>(null)

const quickActions = [
  {
    title: '查看提交记录',
    description: '直接打开审计日志中的 submit 动作，复核高频提交账号。',
    actionLabel: '提交审计',
    query: { action: 'submit' },
  },
  {
    title: '查看登录记录',
    description: '回看 login 日志，继续确认共享 IP 或短时集中登录。',
    actionLabel: '登录审计',
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
  <div class="workspace-shell journal-shell journal-shell-admin journal-hero cheat-shell">
    <div class="workspace-grid">
      <main class="content-pane">
        <CheatDetectionHeroPanel
          :generated-at-label="riskData ? formatDateTime(riskData.generated_at) : null"
          :loading="loading"
          @open-audit="void openAudit({})"
          @refresh="void loadRiskData()"
        />

        <div class="journal-divider" />

        <AppLoading
          v-if="loading && !riskData"
          class="cheat-loading"
        >
          正在扫描合规风险...
        </AppLoading>

        <div
          v-else-if="riskData"
          class="cheat-workbench"
        >
          <CheatDetectionSummaryPanel :summary="riskData.summary" />

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

            <div
              v-else
              class="cheat-directory-list"
            >
              <button
                v-for="suspect in riskData.suspects"
                :key="suspect.user_id"
                type="button"
                class="cheat-directory-row"
                @click="openAudit({ action: 'submit', actor_user_id: String(suspect.user_id) })"
              >
                <div class="cheat-directory-row-main">
                  <div class="cheat-directory-row-title">{{ suspect.username }}</div>
                  <div class="cheat-directory-row-copy">{{ suspect.reason }}</div>
                </div>
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

            <div
              v-else
              class="cheat-directory-list"
            >
              <button
                v-for="group in riskData.shared_ips"
                :key="group.ip"
                type="button"
                class="cheat-directory-row"
                @click="openAudit({ action: 'login' })"
              >
                <div class="cheat-directory-row-main">
                  <div class="cheat-directory-row-title cheat-directory-row-title--mono">{{ group.ip }}</div>
                  <div class="cheat-directory-row-copy">涉及账号：{{ group.usernames.join('、') }}</div>
                </div>
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
                @click="openAudit(action.query)"
              >
                <div class="cheat-directory-row-main">
                  <div class="cheat-directory-row-title">{{ action.title }}</div>
                  <div class="cheat-directory-row-copy">{{ action.description }}</div>
                </div>
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
        </div>

        <div
          v-else-if="error"
          class="cheat-error-box"
          role="alert"
        >
          <AlertCircle class="h-4 w-4" />
          <span>{{ error }}</span>
          <button
            type="button"
            class="ui-btn ui-btn--ghost ui-btn--sm"
            @click="loadRiskData"
          >
            重试
          </button>
        </div>

        <div
          v-else
          class="cheat-empty-shell"
        >
          <ShieldQuestion class="h-12 w-12" />
          <p>当前没有任何风险检出</p>
        </div>
      </main>
    </div>
  </div>
</template>

<style scoped>
.cheat-shell {
  --journal-shell-dark-accent: var(--color-primary-hover);
  --cheat-card-border: color-mix(in srgb, var(--journal-border) 74%, transparent);
  --cheat-divider: color-mix(in srgb, var(--journal-border) 68%, transparent);
  --journal-divider-border: 1px dashed var(--cheat-divider);
  --page-top-tabs-gap: 28px;
  --page-top-tab-font-size: var(--font-size-15);
  --page-top-tab-active-border: color-mix(in srgb, var(--journal-accent) 84%, var(--journal-ink));
  --workspace-brand: var(--journal-accent);
  --workspace-brand-ink: color-mix(in srgb, var(--journal-accent) 74%, var(--journal-ink));
  --workspace-panel: color-mix(in srgb, var(--color-bg-surface) 90%, var(--color-bg-base));
  --workspace-panel-soft: color-mix(in srgb, var(--color-bg-surface) 82%, var(--color-bg-base));
  --workspace-line-soft: color-mix(in srgb, var(--color-text-primary) 10%, transparent);
}

.cheat-loading {
  padding-block: var(--space-7);
}

.cheat-workbench {
  display: grid;
  gap: var(--space-4);
}

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
  grid-template-columns: minmax(0, 1fr) auto;
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
  display: grid;
  gap: var(--space-1-5);
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

.admin-empty {
  border: 1px dashed color-mix(in srgb, var(--journal-border) 72%, transparent);
}

.cheat-empty-state {
  border-top-color: var(--cheat-divider);
  border-bottom-color: var(--cheat-divider);
}

.cheat-error-box {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--space-3);
  border: 1px solid color-mix(in srgb, var(--color-danger) 22%, var(--cheat-card-border));
  border-radius: 1rem;
  background: color-mix(in srgb, var(--color-danger) 8%, transparent);
  padding: var(--space-4);
  color: color-mix(in srgb, var(--color-danger) 84%, var(--journal-ink));
}

.cheat-empty-shell {
  display: grid;
  justify-items: center;
  gap: var(--space-3);
  padding-block: var(--space-8);
  color: var(--journal-muted);
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
