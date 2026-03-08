<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'

import { getCheatDetection } from '@/api/admin'
import type { AdminCheatDetectionData } from '@/api/contracts'
import AppCard from '@/components/common/AppCard.vue'
import AppEmpty from '@/components/common/AppEmpty.vue'
import AppLoading from '@/components/common/AppLoading.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import SectionCard from '@/components/common/SectionCard.vue'

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
  <div class="space-y-6">
    <PageHeader
      eyebrow="Risk Triage"
      title="作弊检测"
      description="当前页已接入真实的 `/admin/cheat-detection` 接口，展示最近一轮基于审计日志聚合的高频提交账号和共享 IP 线索。"
    />

    <div v-if="loading" class="flex justify-center py-10">
      <AppLoading>正在加载风险线索...</AppLoading>
    </div>

    <div v-else class="space-y-6">
      <div
        v-if="error"
        class="rounded-2xl border border-rose-500/20 bg-rose-500/10 px-5 py-4 text-sm text-rose-200"
      >
        {{ error }}
      </div>

      <section v-if="riskData" class="grid gap-4 lg:grid-cols-3">
        <AppCard variant="metric" accent="warning" eyebrow="Submit Burst" :title="String(riskData.summary.submit_burst_users)" subtitle="最近窗口内提交次数超过阈值的账号数。" />
        <AppCard variant="metric" accent="primary" eyebrow="Shared IP" :title="String(riskData.summary.shared_ip_groups)" subtitle="最近 24 小时内存在多账号复用的 IP 组数。" />
        <AppCard variant="metric" accent="danger" eyebrow="Affected Users" :title="String(riskData.summary.affected_users)" subtitle="当前聚合结果覆盖到的可疑账号总数。" />
      </section>

      <section class="grid gap-6 xl:grid-cols-[1.1fr_0.9fr]">
        <SectionCard
          title="高频提交账号"
          subtitle="这些账号在最近窗口内的提交次数超过阈值，建议先结合审计日志复核。"
        >
          <AppEmpty
            v-if="!riskData?.suspects.length"
            icon="UsersRound"
            title="当前没有超过阈值的高频提交账号"
            description="说明最近窗口内还没有明显的提交爆发样本。"
          />

          <div v-else class="space-y-3">
            <AppCard
              v-for="suspect in riskData.suspects"
              :key="suspect.user_id"
              variant="action"
              accent="warning"
            >
              <div class="flex items-start justify-between gap-4">
                <div>
                  <p class="font-medium text-[var(--color-text-primary)]">{{ suspect.username }}</p>
                  <p class="mt-1 text-sm text-[var(--color-text-secondary)]">
                    {{ suspect.reason }}
                  </p>
                </div>
                <span
                  class="rounded-full bg-amber-500/10 px-3 py-1 text-xs font-semibold text-amber-700"
                >
                  {{ suspect.submit_count }} 次
                </span>
              </div>
              <p class="mt-3 text-xs text-[var(--color-text-secondary)]">
                最近出现时间：{{ new Date(suspect.last_seen_at).toLocaleString('zh-CN') }}
              </p>
            </AppCard>
          </div>
        </SectionCard>

        <SectionCard
          title="共享 IP 线索"
          subtitle="同一 IP 在最近 24 小时内出现多个账号登录，适合作为第二层排查线索。"
        >
          <AppEmpty
            v-if="!riskData?.shared_ips.length"
            icon="UsersRound"
            title="当前没有共享 IP 线索"
            description="最近 24 小时内还没有发现明显的多账号复用 IP。"
          />

          <div v-else class="space-y-3">
            <AppCard
              v-for="group in riskData.shared_ips"
              :key="group.ip"
              variant="action"
              accent="primary"
            >
              <div class="flex items-start justify-between gap-4">
                <div>
                  <p class="font-mono text-sm text-[var(--color-text-primary)]">{{ group.ip }}</p>
                  <p class="mt-1 text-sm text-[var(--color-text-secondary)]">
                    {{ group.usernames.join('、') || '匿名记录' }}
                  </p>
                </div>
                <span
                  class="rounded-full bg-cyan-500/10 px-3 py-1 text-xs font-semibold text-cyan-700"
                >
                  {{ group.user_count }} 账号
                </span>
              </div>
            </AppCard>
          </div>
        </SectionCard>
      </section>

      <SectionCard
        title="快速排查入口"
        subtitle="保留直接跳转审计日志的入口，便于把自动聚合结果继续下钻到原始记录。"
      >
        <div class="grid gap-3 lg:grid-cols-2">
          <AppCard
            v-for="action in quickActions"
            :key="action.title"
            as="button"
            variant="action"
            accent="warning"
            interactive
            class="cursor-pointer"
            @click="openAudit(action.query)"
          >
            <div>
              <p class="font-medium text-[var(--color-text-primary)]">{{ action.title }}</p>
              <p class="mt-1 text-sm leading-6 text-[var(--color-text-secondary)]">
                {{ action.description }}
              </p>
            </div>
            <span class="mt-0.5 text-sm font-medium text-[var(--color-primary)]">打开</span>
          </AppCard>
        </div>
      </SectionCard>
    </div>
  </div>
</template>
