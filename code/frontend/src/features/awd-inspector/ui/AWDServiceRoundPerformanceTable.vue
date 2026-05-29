<script setup lang="ts">
import type { AWDRoundSummaryData } from '@/api/contracts'

defineProps<{
  summary: AWDRoundSummaryData
}>()
</script>

<template>
  <section class="round-performance-area mt-12">
    <header class="performance-header">
      <h3 class="viewer-title">
        本轮得分与健康表现
      </h3>
      <div class="filter-summary">
        Round Performance Summary
      </div>
    </header>

    <div class="log-table-wrap mt-4">
      <table class="studio-table">
        <thead>
          <tr>
            <th>队伍节点</th>
            <th class="text-right">
              本轮得分
            </th>
            <th class="text-right">
              SLA / ATK / DEF
            </th>
            <th class="text-right">
              服务健康
            </th>
            <th class="text-right">
              被攻破统计
            </th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="item in summary.items"
            :key="item.team_id"
            class="studio-row"
          >
            <td class="performance-team-name font-bold">
              {{ item.team_name }}
            </td>
            <td class="performance-total-score text-right font-mono font-black">
              {{ item.total_score }}
            </td>
            <td class="performance-score-breakdown text-right font-mono">
              {{ item.sla_score ?? 0 }} / {{ item.attack_score }} / {{ item.defense_score }}
            </td>
            <td class="text-right">
              <div class="health-stack">
                <span class="health-stack__up">{{ item.service_up_count }} UP</span>
                <span class="health-stack__separator">/</span>
                <span class="health-stack__down">{{ item.service_down_count }} OFF</span>
                <span class="health-stack__separator">/</span>
                <span class="health-stack__compromised">{{ item.service_compromised_count }} EXP</span>
              </div>
            </td>
            <td class="performance-breach-count text-right">
              攻破 {{ item.successful_breach_count }} 次 · {{ item.unique_attackers_against }} 攻击方
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </section>
</template>
