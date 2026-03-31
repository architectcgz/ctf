<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'

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

onMounted(() => {
  void loadRiskData()
})
</script>

<template>
  <section class="journal-shell journal-hero flex min-h-full flex-col rounded-[30px] border px-6 py-6 md:px-8">
      <div class="grid gap-6 xl:grid-cols-[1.06fr_0.94fr]">
        <div>
          <div class="journal-eyebrow">Risk Triage</div>
          <h1 class="mt-3 text-3xl font-semibold tracking-tight text-[var(--journal-ink)] md:text-[2.45rem]">作弊检测</h1>
          <p class="mt-3 max-w-2xl text-sm leading-7 text-[var(--journal-muted)]">
            查看高频提交账号和共享 IP 线索，并继续下钻到审计日志。
          </p>
        </div>

        <article v-if="riskData" class="journal-brief rounded-[24px] border px-5 py-5">
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

      <div v-if="loading" class="flex justify-center py-10">
        <AppLoading>正在加载风险线索...</AppLoading>
      </div>

      <template v-else-if="riskData">
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

        <div class="journal-divider" />

        <div class="space-y-3">
          <div class="admin-section-head">
            <div>
              <div class="journal-note-label">Burst Accounts</div>
              <h2 class="mt-2 text-xl font-semibold text-[var(--journal-ink)]">高频提交账号</h2>
            </div>
          </div>

          <AppEmpty
            v-if="!riskData?.suspects.length"
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
        </div>

        <div class="journal-divider" />

        <div class="space-y-3">
          <div class="admin-section-head">
            <div>
              <div class="journal-note-label">Shared IP</div>
              <h2 class="mt-2 text-xl font-semibold text-[var(--journal-ink)]">共享 IP 线索</h2>
            </div>
          </div>

          <AppEmpty
            v-if="!riskData?.shared_ips.length"
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
        </div>

        <div class="journal-divider" />

        <div class="space-y-3">
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
        </div>
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
    </section>
</template>

<style scoped>
.journal-shell {
  --journal-ink: #0f172a;
  --journal-muted: #64748b;
  --journal-accent: #2563eb;
  --journal-border: rgba(226, 232, 240, 0.84);
  --journal-surface: rgba(248, 250, 252, 0.92);
  --journal-surface-subtle: rgba(241, 245, 249, 0.72);
}

.journal-hero,
.journal-panel {
  border-color: var(--journal-border);
  background:
    radial-gradient(circle at top right, rgba(37, 99, 235, 0.08), transparent 18rem),
    linear-gradient(180deg, #ffffff, #f8fafc);
  border-radius: 16px !important;
  box-shadow: 0 18px 40px rgba(15, 23, 42, 0.06);
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
  border-top: 1px dashed rgba(148, 163, 184, 0.7);
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
  border: 1px solid var(--journal-border);
  border-radius: 18px;
  background: rgba(255, 255, 255, 0.74);
  padding: 1rem;
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
  --journal-ink: #e2e8f0;
  --journal-muted: #94a3b8;
  --journal-accent: #60a5fa;
  --journal-border: rgba(71, 85, 105, 0.78);
  --journal-surface: rgba(15, 23, 42, 0.7);
  --journal-surface-subtle: rgba(15, 23, 42, 0.78);
}

:global([data-theme='dark']) .journal-hero,
:global([data-theme='dark']) .journal-panel {
  background:
    radial-gradient(circle at top right, rgba(96, 165, 250, 0.1), transparent 18rem),
    linear-gradient(180deg, rgba(15, 23, 42, 0.96), rgba(15, 23, 42, 0.9));
}
</style>
