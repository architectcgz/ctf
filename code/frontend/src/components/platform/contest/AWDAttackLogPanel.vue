<script setup lang="ts">
import { ShieldCheck, FileDown, ArrowRight, Sword, Target, History } from 'lucide-vue-next'

import type {
  AWDAttackLogPanelEmits,
  AWDAttackLogPanelProps,
} from '@/components/platform/contest/awdInspector.types'

const props = defineProps<AWDAttackLogPanelProps>()
const emit = defineEmits<AWDAttackLogPanelEmits>()

function updateAttackResultFilter(value: string): void {
  if (value !== 'all' && value !== 'success' && value !== 'failed') {
    return
  }
  emit('updateAttackResultFilter', value)
}

function updateAttackSourceFilter(value: string): void {
  if (value !== 'all' && !props.attackSourceOptions.includes(value as any)) {
    return
  }
  emit('updateAttackSourceFilter', value as any)
}
</script>

<template>
  <div class="studio-attack-log">
    <!-- 1. Sophisticated Log Toolbar -->
    <header class="log-toolbar">
      <div class="toolbar-left">
        <div class="flex items-center gap-3">
          <div class="toolbar-icon">
            <Sword class="h-4 w-4" />
          </div>
          <div>
            <h3 class="toolbar-title">
              攻击流量审计流水
            </h3>
            <p class="toolbar-hint">
              Real-time Attack Vector Monitoring
            </p>
          </div>
        </div>
      </div>

      <div class="toolbar-right">
        <div class="filter-actions-group">
          <div class="filter-pills">
            <select
              :value="attackTeamFilter"
              class="log-select"
              @change="emit('updateAttackTeamFilter', ($event.target as HTMLSelectElement).value)"
            >
              <option value="">
                全部参与队
              </option>
              <option
                v-for="team in attackTeamOptions"
                :key="team.id"
                :value="team.id"
              >
                {{ team.name }}
              </option>
            </select>
            <select
              :value="attackResultFilter"
              class="log-select"
              @change="updateAttackResultFilter(($event.target as HTMLSelectElement).value)"
            >
              <option value="all">
                所有交互结果
              </option>
              <option value="success">
                攻击成功 (BREACH)
              </option>
              <option value="failed">
                尝试失败 (DROP)
              </option>
            </select>
          </div>
          <button
            type="button"
            class="ops-btn ops-btn--neutral"
            :disabled="filteredAttacks.length === 0"
            @click="emit('exportAttacks')"
          >
            <FileDown class="h-3.5 w-3.5" />
            <span>导出审计日志</span>
          </button>
        </div>
      </div>
    </header>

    <!-- 2. High-Density Event Feed -->
    <div class="event-feed-container custom-scrollbar">
      <div
        v-if="filteredAttacks.length === 0"
        class="empty-feed"
      >
        <History class="h-10 w-10 opacity-10 mb-4" />
        <p>当前时段未监测到符合条件的攻击矢量</p>
      </div>

      <div
        v-else
        class="event-list"
      >
        <article 
          v-for="attack in filteredAttacks" 
          :key="attack.id" 
          class="event-row"
          :class="{ 'is-success': attack.is_success }"
        >
          <!-- Time & Status Marker -->
          <div class="event-marker">
            <div class="marker-time font-mono">
              {{ formatDateTime(attack.created_at).split(' ')[1] }}
            </div>
            <div class="marker-dot" />
          </div>

          <!-- Interaction Detail -->
          <div class="event-body">
            <div class="event-primary">
              <div class="vector-wrap">
                <span
                  class="actor attacker"
                  title="攻击方"
                >{{ attack.attacker_team }}</span>
                <div class="vector-line">
                  <div class="line-path" />
                  <ArrowRight class="h-3 w-3 line-head" />
                </div>
                <span
                  class="actor victim"
                  title="受害方"
                >{{ attack.victim_team }}</span>
              </div>

              <div class="event-meta">
                <div class="meta-item">
                  <Target class="h-3 w-3" />
                  <span>{{ getChallengeTitle(attack.challenge_id) }}</span>
                </div>
                <div class="meta-divider" />
                <div class="meta-item opacity-60">
                  <span>{{ getAttackSourceLabel(attack.source) }}</span>
                </div>
              </div>
            </div>

            <div class="event-secondary">
              <div
                v-if="attack.is_success"
                class="result-badge success"
              >
                <ShieldCheck class="h-3 w-3" />
                <span>SUCCESS</span>
                <span class="score-delta font-mono">+{{ attack.score_gained }}</span>
              </div>
              <div
                v-else
                class="result-badge failed"
              >
                <span>DROPPED</span>
              </div>
            </div>
          </div>
        </article>
      </div>
    </div>
  </div>
</template>

<style scoped>
.studio-attack-log { display: flex; flex-direction: column; gap: 2rem; background: transparent; }

/* Toolbar Enhancement */
.log-toolbar { display: flex; justify-content: space-between; align-items: flex-end; padding: 0 0 1.5rem; border-bottom: 1px solid color-mix(in srgb, var(--workspace-line-soft) 60%, transparent); }
.toolbar-icon { width: 2.75rem; height: 2.75rem; border-radius: 0.85rem; background: #f1f5f9; color: #475569; display: flex; align-items: center; justify-content: center; }
.toolbar-title { font-size: 15px; font-weight: 900; color: #0f172a; margin: 0; letter-spacing: -0.01em; }
.toolbar-hint { font-size: 11px; color: #94a3b8; font-weight: 700; text-transform: uppercase; letter-spacing: 0.1em; margin-top: 0.15rem; }

.filter-actions-group { display: flex; align-items: center; gap: 1.5rem; }
.filter-pills { display: flex; gap: 0.5rem; }
.log-select { height: 2.25rem; padding: 0 0.75rem; font-size: 12px; font-weight: 700; border-radius: 0.6rem; border: 1px solid #e2e8f0; background: white; color: #475569; outline: none; transition: all 0.2s ease; }
.log-select:hover { border-color: #cbd5e1; }

/* Event Feed Design */
.event-feed-container { flex: 1; position: relative; }

.event-list { display: flex; flex-direction: column; }

.event-row { display: flex; gap: 2rem; padding: 1rem 0; transition: all 0.2s ease; border-bottom: 1px solid #f1f5f9; }
.event-row:hover { background: rgba(248, 250, 252, 0.6); }
.event-row:last-child { border-bottom: none; }

/* Timeline Marker */
.event-marker { display: flex; flex-direction: column; align-items: flex-end; width: 5rem; shrink: 0; padding-top: 0.25rem; }
.marker-time { font-size: 11px; font-weight: 800; color: #94a3b8; }
.marker-dot { width: 8px; height: 8px; border-radius: 50%; background: #e2e8f0; margin-top: 0.5rem; position: relative; }
.marker-dot::after { content: ''; position: absolute; top: 8px; left: 50%; transform: translateX(-50%); width: 1px; height: 2.5rem; background: #f1f5f9; }
.event-row:last-child .marker-dot::after { display: none; }

.is-success .marker-dot { background: #22c55e; box-shadow: 0 0 10px rgba(34, 197, 94, 0.4); }

/* Event Body */
.event-body { flex: 1; display: flex; justify-content: space-between; align-items: center; }

.event-primary { display: flex; flex-direction: column; gap: 0.75rem; }

.vector-wrap { display: flex; align-items: center; gap: 1rem; }
.actor { font-size: 14px; font-weight: 900; letter-spacing: -0.01em; }
.actor.attacker { color: #0f172a; }
.actor.victim { color: #64748b; }

.vector-line { position: relative; width: 4rem; display: flex; align-items: center; }
.line-path { height: 2px; width: 100%; background: #e2e8f0; border-radius: 1px; }
.line-head { position: absolute; right: -4px; }
.is-success .line-path { background: #cbd5e1; }
.is-success .line-head { color: #94a3b8; }

.event-meta { display: flex; align-items: center; gap: 1rem; }
.meta-item { display: flex; align-items: center; gap: 0.5rem; font-size: 11px; font-weight: 700; color: #64748b; }
.meta-divider { width: 3px; height: 3px; border-radius: 50%; background: #cbd5e1; }

/* Result Badges */
.event-secondary { display: flex; align-items: center; }
.result-badge { display: flex; align-items: center; gap: 0.65rem; padding: 0.45rem 1rem; border-radius: 0.75rem; font-size: 10px; font-weight: 900; letter-spacing: 0.05em; }
.result-badge.success { background: #f0fdf4; color: #166534; border: 1px solid #bbf7d0; }
.result-badge.failed { background: #f8fafc; color: #94a3b8; border: 1px solid #e2e8f0; }

.score-delta { font-size: 13px; color: #16a34a; margin-left: 0.5rem; border-left: 1px solid rgba(22, 101, 52, 0.1); padding-left: 0.75rem; }

/* Global UI Primitives */
.ops-btn { display: inline-flex; align-items: center; gap: 0.5rem; height: 2.25rem; padding: 0 1.25rem; border-radius: 0.75rem; font-size: 12px; font-weight: 800; cursor: pointer; transition: all 0.2s ease; }
.ops-btn--neutral { background: white; border: 1px solid #e2e8f0; color: #475569; }
.ops-btn--neutral:hover { border-color: #cbd5e1; background: #f8fafc; }

.empty-feed { display: flex; flex-direction: column; align-items: center; justify-content: center; padding: 5rem 0; color: #94a3b8; font-size: 14px; font-weight: 600; }

.custom-scrollbar::-webkit-scrollbar { width: 4px; }
.custom-scrollbar::-webkit-scrollbar-thumb { background: #e2e8f0; border-radius: 10px; }
</style>
