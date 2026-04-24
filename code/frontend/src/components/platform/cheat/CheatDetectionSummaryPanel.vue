<script setup lang="ts">
import { Fingerprint, ShieldAlert, Users } from 'lucide-vue-next'

defineProps<{
  summary: {
    submit_burst_users: number
    shared_ip_groups: number
    affected_users: number
  }
}>()
</script>

<template>
  <div
    class="admin-summary-grid cheat-kpi-summary progress-strip metric-panel-grid metric-panel-default-surface metric-panel-workspace-surface"
  >
    <article class="journal-note progress-card metric-panel-card">
      <div class="journal-note-label progress-card-label metric-panel-label">
        <span>提交风险账号</span>
        <ShieldAlert class="h-4 w-4" />
      </div>
      <div class="journal-note-value progress-card-value metric-panel-value">
        {{ summary.submit_burst_users.toString().padStart(2, '0') }}
      </div>
      <div class="journal-note-helper progress-card-hint metric-panel-helper">
        超过高频阈值的账号数量
      </div>
    </article>

    <article class="journal-note progress-card metric-panel-card">
      <div class="journal-note-label progress-card-label metric-panel-label">
        <span>共享 IP 线索</span>
        <Fingerprint class="h-4 w-4" />
      </div>
      <div class="journal-note-value progress-card-value metric-panel-value">
        {{ summary.shared_ip_groups.toString().padStart(2, '0') }}
      </div>
      <div class="journal-note-helper progress-card-hint metric-panel-helper">
        最近窗口内出现的共享公网地址分组
      </div>
    </article>

    <article class="journal-note progress-card metric-panel-card">
      <div class="journal-note-label progress-card-label metric-panel-label">
        <span>涉及用户总数</span>
        <Users class="h-4 w-4" />
      </div>
      <div class="journal-note-value progress-card-value metric-panel-value">
        {{ summary.affected_users.toString().padStart(2, '0') }}
      </div>
      <div class="journal-note-helper progress-card-hint metric-panel-helper">
        当前所有线索累计覆盖的账号规模
      </div>
    </article>
  </div>
</template>

<style scoped>
.cheat-kpi-summary {
  --metric-panel-columns: repeat(3, minmax(0, 1fr));
}

.cheat-kpi-summary .metric-panel-label {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-2);
}

@media (max-width: 1100px) {
  .cheat-kpi-summary {
    --metric-panel-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 720px) {
  .cheat-kpi-summary {
    --metric-panel-columns: 1fr;
  }
}
</style>
