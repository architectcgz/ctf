<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { AlertCircle, Fingerprint, RefreshCw, SearchCheck, ShieldAlert, Users } from 'lucide-vue-next'

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
  <section
    class="workspace-shell journal-shell journal-shell-admin journal-notes-card journal-hero flex min-h-full flex-1 flex-col"
  >
    <main class="content-pane">
      <div v-if="loading" class="flex justify-center py-10">
        <AppLoading>正在加载风险线索...</AppLoading>
      </div>

      <template v-else-if="riskData">
        <section class="cheat-workbench">
          <header class="workspace-tab-heading cheat-workbench-head">
            <div class="workspace-tab-heading__main">
              <div class="workspace-overline">Integrity Workspace</div>
              <h1 class="workspace-page-title">作弊检测</h1>
              <p class="workspace-page-copy">
                基于提交爆发、IP 共享及行为指纹的多维度合规分析。
              </p>
            </div>

            <div class="cheat-workbench-actions">
              <div class="cheat-workbench-meta">
                <span class="cheat-workbench-meta__label">最近生成</span>
                <span class="cheat-workbench-meta__value">
                  {{ formatDateTime(riskData.generated_at) }}
                </span>
              </div>
              <button type="button" class="ui-btn ui-btn--ghost" @click="openAudit({})">
                <SearchCheck class="h-4 w-4" />
                打开审计日志
              </button>
              <button type="button" class="ui-btn ui-btn--primary" @click="loadRiskData">
                <RefreshCw class="h-4 w-4" />
                刷新线索
              </button>
            </div>
          </header>

          <div class="metric-panel-grid--premium cols-3">
            <article class="metric-panel-card--premium">
              <div class="metric-panel-label">
                <span>Submit Burst</span>
                <ShieldAlert class="h-4 w-4" />
              </div>
              <div class="metric-panel-value">
                {{ riskData.summary.submit_burst_users.toString().padStart(2, '0') }}
              </div>
              <div class="metric-panel-helper">高频提交风险账号</div>
            </article>

            <article class="metric-panel-card--premium">
              <div class="metric-panel-label">
                <span>Shared IP</span>
                <Fingerprint class="h-4 w-4" />
              </div>
              <div class="metric-panel-value">
                {{ riskData.summary.shared_ip_groups.toString().padStart(2, '0') }}
              </div>
              <div class="metric-panel-helper">多账号共享 IP 组数</div>
            </article>

            <article class="metric-panel-card--premium">
              <div class="metric-panel-label">
                <span>Affected Users</span>
                <Users class="h-4 w-4" />
              </div>
              <div class="metric-panel-value">
                {{ riskData.summary.affected_users.toString().padStart(2, '0') }}
              </div>
              <div class="metric-panel-helper">受风险波及的学生总数</div>
            </article>
          </div>

          <section class="workspace-directory-section cheat-directory-section">
            <header class="list-heading">
              <div>
                <div class="journal-note-label">Burst Accounts</div>
                <h2 class="list-heading__title">高频提交账号</h2>
              </div>
              <div class="cheat-directory-caption">按账号查看提交频次、最近出现时间和审计入口</div>
            </header>

            <AppEmpty
              v-if="!riskData.suspects.length"
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

          <section class="workspace-directory-section cheat-directory-section">
            <header class="list-heading">
              <div>
                <div class="journal-note-label">Shared IP</div>
                <h2 class="list-heading__title">共享 IP 线索</h2>
              </div>
              <div class="cheat-directory-caption">按 IP 聚合查看复用账号范围和登录审计入口</div>
            </header>

            <AppEmpty
              v-if="!riskData.shared_ips.length"
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

          <section class="workspace-directory-section cheat-directory-section">
            <header class="list-heading">
              <div>
                <div class="journal-note-label">Audit Actions</div>
                <h2 class="list-heading__title">审计联动</h2>
              </div>
              <div class="cheat-directory-caption">保留常用的日志入口，作为底部补充动作区</div>
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
  --workspace-brand: var(--journal-accent);
  --workspace-brand-ink: color-mix(in srgb, var(--journal-accent) 74%, var(--journal-ink));
  --workspace-panel: color-mix(in srgb, var(--color-bg-surface) 90%, var(--color-bg-base));
  --workspace-panel-soft: color-mix(in srgb, var(--color-bg-surface) 82%, var(--color-bg-base));
  --workspace-line-soft: color-mix(in srgb, var(--color-text-primary) 10%, transparent);
  --journal-divider-border: 1px dashed var(--cheat-divider);
  --journal-shell-dark-accent: var(--color-primary-hover);
}

.content-pane {
  display: flex;
  flex: 1 1 auto;
  flex-direction: column;
  gap: var(--space-4);
}

.cheat-workbench {
  display: grid;
  gap: var(--space-4);
}

.cheat-workbench-head {
  gap: var(--space-3);
}

.cheat-workbench-actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: flex-end;
  gap: var(--space-3);
}

.cheat-workbench-meta {
  display: grid;
  gap: var(--space-1);
  justify-items: end;
  min-width: 12rem;
}

.cheat-workbench-meta__label {
  font-size: var(--font-size-0-72);
  font-weight: 700;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.cheat-workbench-meta__value {
  font-size: var(--font-size-0-88);
  color: var(--journal-ink);
}

.cheat-directory-section {
  display: grid;
  gap: var(--space-4);
  padding: 0;
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
  transition: all 160ms ease;
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
  .cheat-workbench-head {
    align-items: flex-start;
    flex-direction: column;
  }

  .cheat-workbench-actions {
    justify-content: flex-start;
  }

  .cheat-workbench-meta,
  .cheat-directory-row-meta {
    justify-items: start;
  }

  .cheat-directory-row,
  .quick-action-row {
    align-items: flex-start;
    flex-direction: column;
  }
}
</style>
