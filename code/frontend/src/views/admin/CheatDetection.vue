<script setup lang="ts">
import { useRouter } from 'vue-router'

const router = useRouter()

const riskSignals = [
  {
    title: '高频提交',
    description: '结合审计日志中的 submit 动作，排查短时间内异常密集的答题提交。',
    badge: '审计日志',
  },
  {
    title: '批量账号行为',
    description: '对比 login 与 submit 记录，识别同一时间段内相似操作轨迹的账号。',
    badge: '人工研判',
  },
  {
    title: '异常资源波动',
    description: '回看系统概览中的容器告警，辅助判断是否存在脚本化爆破或滥用资源。',
    badge: '系统概览',
  },
] as const

const quickActions = [
  {
    title: '查看提交记录',
    description: '直接打开审计日志中的提交动作，优先筛查挑战提交异常。',
    query: { action: 'submit' },
  },
  {
    title: '查看登录记录',
    description: '回看登录日志，核对是否存在短时间多账号集中登录。',
    query: { action: 'login' },
  },
  {
    title: '按 challenge 资源过滤',
    description: '定位 challenge 相关操作，交叉验证题目提交与配置变更。',
    query: { resource_type: 'challenge' },
  },
] as const

function openAudit(query: Record<string, string>) {
  return router.push({ name: 'AuditLog', query })
}

function openDashboard() {
  return router.push({ name: 'AdminDashboard' })
}
</script>

<template>
  <div class="space-y-6">
    <section class="rounded-[28px] border border-[var(--color-border-default)] bg-[linear-gradient(135deg,rgba(127,29,29,0.08),rgba(217,119,6,0.14))] p-7 shadow-sm">
      <p class="text-xs font-semibold uppercase tracking-[0.28em] text-red-600/85">Risk Triage</p>
      <h1 class="mt-3 text-3xl font-semibold tracking-tight text-[var(--color-text-primary)]">作弊检测</h1>
      <p class="mt-2 max-w-3xl text-sm leading-6 text-[var(--color-text-secondary)]">
        当前版本尚未接入独立的作弊检测 API，这里提供基于审计日志与系统概览的研判入口，不展示伪造风险分数。
      </p>
    </section>

    <section class="grid gap-4 lg:grid-cols-3">
      <article
        v-for="signal in riskSignals"
        :key="signal.title"
        class="rounded-2xl border border-[var(--color-border-default)] bg-[var(--color-bg-surface)] p-5 shadow-sm"
      >
        <span class="rounded-full bg-amber-500/10 px-3 py-1 text-xs font-semibold text-amber-700">{{ signal.badge }}</span>
        <h2 class="mt-4 text-lg font-semibold text-[var(--color-text-primary)]">{{ signal.title }}</h2>
        <p class="mt-2 text-sm leading-6 text-[var(--color-text-secondary)]">{{ signal.description }}</p>
      </article>
    </section>

    <section class="grid gap-6 xl:grid-cols-[1.1fr_0.9fr]">
      <div class="rounded-2xl border border-[var(--color-border-default)] bg-[var(--color-bg-surface)] p-6 shadow-sm">
        <div class="flex items-start justify-between gap-4">
          <div>
            <h2 class="text-lg font-semibold text-[var(--color-text-primary)]">快速排查入口</h2>
            <p class="mt-1 text-sm text-[var(--color-text-secondary)]">
              通过预置筛选直达审计日志，减少人工排查的首屏准备成本。
            </p>
          </div>
          <button
            type="button"
            class="rounded-xl border border-[var(--color-border-default)] px-4 py-2 text-sm font-medium text-[var(--color-text-primary)] transition hover:border-[var(--color-primary)]"
            @click="openDashboard"
          >
            查看系统概览
          </button>
        </div>

        <div class="mt-5 space-y-3">
          <button
            v-for="action in quickActions"
            :key="action.title"
            type="button"
            class="flex w-full items-start justify-between gap-4 rounded-xl border border-[var(--color-border-default)] bg-[var(--color-bg-base)] px-4 py-4 text-left transition hover:-translate-y-0.5 hover:border-[var(--color-primary)]"
            @click="openAudit(action.query)"
          >
            <div>
              <p class="font-medium text-[var(--color-text-primary)]">{{ action.title }}</p>
              <p class="mt-1 text-sm leading-6 text-[var(--color-text-secondary)]">{{ action.description }}</p>
            </div>
            <span class="mt-0.5 text-sm font-medium text-[var(--color-primary)]">打开</span>
          </button>
        </div>
      </div>

      <div class="rounded-2xl border border-[var(--color-border-default)] bg-[var(--color-bg-surface)] p-6 shadow-sm">
        <h2 class="text-lg font-semibold text-[var(--color-text-primary)]">当前边界</h2>
        <ul class="mt-4 space-y-3 text-sm leading-6 text-[var(--color-text-secondary)]">
          <li>暂未提供独立作弊检测接口，无法直接输出风险名单或相似度分析。</li>
          <li>本页只提供运营研判入口，最终结论仍需结合审计日志和容器告警人工确认。</li>
          <li>后续如果后端补充专用检测接口，可在此页无缝替换为自动化结果面板。</li>
        </ul>
      </div>
    </section>
  </div>
</template>
