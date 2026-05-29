<script setup lang="ts">
import type { AWDTrafficIntelligenceGridProps } from './awdTrafficPanel.types'

defineProps<AWDTrafficIntelligenceGridProps>()
</script>

<template>
  <div class="intelligence-grid">
    <div class="intel-column">
      <header class="intel-header">
        热点实体分析
      </header>
      <div class="intel-cards">
        <div class="intel-sub-card">
          <div class="sub-card-label">
            主力攻击队
          </div>
          <div class="sub-card-list">
            <div
              v-for="item in summary.top_attackers.slice(0, 3)"
              :key="item.team_id"
              class="list-row"
            >
              <span class="row-name">{{ item.team_name }}</span>
              <span class="row-count font-mono">{{ item.request_count }}</span>
            </div>
          </div>
        </div>
        <div class="intel-sub-card">
          <div class="sub-card-label">
            高频受害队
          </div>
          <div class="sub-card-list">
            <div
              v-for="item in summary.top_victims.slice(0, 3)"
              :key="item.team_id"
              class="list-row"
            >
              <span class="row-name">{{ item.team_name }}</span>
              <span class="row-count font-mono">{{ item.request_count }}</span>
            </div>
          </div>
        </div>
        <div class="intel-sub-card">
          <div class="sub-card-label">
            异常交互路径
          </div>
          <div class="sub-card-list">
            <div
              v-for="item in summary.top_error_paths.slice(0, 3)"
              :key="item.path"
              class="list-row"
            >
              <span class="row-name truncate font-mono text-[10px]">{{ item.path }}</span>
              <span class="row-count row-count--danger font-mono">{{ item.error_count }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div class="intel-column">
      <header class="intel-header">
        流量趋势 (12-Bucket Trend)
      </header>
      <div class="trend-canvas">
        <div
          v-for="bucket in trendRows"
          :key="bucket.bucket_start_at"
          class="trend-unit"
        >
          <div class="trend-meta">
            <span class="trend-label">{{ bucket.label }}</span>
            <span class="trend-data">{{ bucket.request_count }} REQS</span>
          </div>
          <div class="trend-bar-track">
            <div
              class="trend-bar-fill"
              :style="{ width: `${bucket.ratio}%` }"
            />
          </div>
        </div>
        <p
          v-if="trendRows.length === 0"
          class="traffic-empty-hint py-4"
        >
          等待数据注入趋势桶...
        </p>
        <p
          v-else
          class="traffic-empty-hint"
        >
          {{ trendNarrative }}
        </p>
      </div>
    </div>
  </div>
</template>
