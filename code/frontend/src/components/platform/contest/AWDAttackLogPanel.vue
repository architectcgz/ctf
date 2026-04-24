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
              id="awd-attack-filter-team"
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
              id="awd-attack-filter-result"
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
            <select
              id="awd-attack-filter-source"
              :value="attackSourceFilter"
              class="log-select"
              @change="updateAttackSourceFilter(($event.target as HTMLSelectElement).value)"
            >
              <option value="all">
                所有来源
              </option>
              <option
                v-for="source in attackSourceOptions"
                :key="source"
                :value="source"
              >
                {{ getAttackSourceLabel(source) }}
              </option>
            </select>
          </div>
          <button
            id="awd-export-attacks"
            type="button"
            class="ui-btn ui-btn--secondary attack-log-export-button"
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
.studio-attack-log {
  --attack-log-surface: color-mix(in srgb, var(--color-bg-surface) 94%, var(--color-bg-base));
  --attack-log-surface-subtle: color-mix(in srgb, var(--color-bg-surface) 84%, var(--color-bg-base));
  --attack-log-border: color-mix(in srgb, var(--color-border-default) 86%, transparent);
  --attack-log-border-strong: color-mix(in srgb, var(--color-border-default) 94%, transparent);
  --attack-log-text: var(--color-text-primary);
  --attack-log-muted: var(--color-text-secondary);
  --attack-log-faint: var(--color-text-muted);
  --attack-log-success-surface: color-mix(in srgb, var(--color-success) 10%, var(--color-bg-surface));
  --attack-log-success-border: color-mix(in srgb, var(--color-success) 22%, var(--attack-log-border));
  --attack-log-success-text: color-mix(in srgb, var(--color-success) 78%, var(--color-text-primary));
  --attack-log-success-glow: color-mix(in srgb, var(--color-success) 42%, transparent);
  --attack-log-success-divider: color-mix(in srgb, var(--color-success) 24%, transparent);
  display: flex;
  flex-direction: column;
  gap: 2rem;
  background: transparent;
}

/* Toolbar Enhancement */
.log-toolbar { display: flex; justify-content: space-between; align-items: flex-end; padding: 0 0 1.5rem; border-bottom: 1px solid color-mix(in srgb, var(--workspace-line-soft, var(--attack-log-border)) 60%, transparent); }
.toolbar-icon { width: 2.75rem; height: 2.75rem; border-radius: 0.85rem; background: var(--attack-log-surface-subtle); color: var(--attack-log-muted); display: flex; align-items: center; justify-content: center; border: 1px solid var(--attack-log-border); }
.toolbar-title { font-size: 15px; font-weight: 900; color: var(--attack-log-text); margin: 0; letter-spacing: -0.01em; }
.toolbar-hint { font-size: 11px; color: var(--attack-log-faint); font-weight: 700; text-transform: uppercase; letter-spacing: 0.1em; margin-top: 0.15rem; }

.filter-actions-group { display: flex; align-items: center; gap: 1.5rem; }
.filter-pills { display: flex; gap: 0.5rem; }
.log-select { height: 2.25rem; padding: 0 0.75rem; font-size: 12px; font-weight: 700; border-radius: 0.6rem; border: 1px solid var(--attack-log-border); background: var(--attack-log-surface); color: var(--attack-log-muted); outline: none; transition: all 0.2s ease; }
.log-select:hover,
.log-select:focus-visible { border-color: var(--attack-log-border-strong); color: var(--attack-log-text); }

/* Event Feed Design */
.event-feed-container { flex: 1; position: relative; }

.event-list { display: flex; flex-direction: column; }

.event-row { display: flex; gap: 2rem; padding: 1rem 0; transition: all 0.2s ease; border-bottom: 1px solid var(--attack-log-border); }
.event-row:hover { background: color-mix(in srgb, var(--attack-log-surface-subtle) 72%, transparent); }
.event-row:last-child { border-bottom: none; }

/* Timeline Marker */
.event-marker { display: flex; flex-direction: column; align-items: flex-end; width: 5rem; shrink: 0; padding-top: 0.25rem; }
.marker-time { font-size: 11px; font-weight: 800; color: var(--attack-log-faint); }
.marker-dot { width: 8px; height: 8px; border-radius: 50%; background: var(--attack-log-border-strong); margin-top: 0.5rem; position: relative; }
.marker-dot::after { content: ''; position: absolute; top: 8px; left: 50%; transform: translateX(-50%); width: 1px; height: 2.5rem; background: var(--attack-log-border); }
.event-row:last-child .marker-dot::after { display: none; }

.is-success .marker-dot { background: var(--color-success); box-shadow: 0 0 10px var(--attack-log-success-glow); }

/* Event Body */
.event-body { flex: 1; display: flex; justify-content: space-between; align-items: center; }

.event-primary { display: flex; flex-direction: column; gap: 0.75rem; }

.vector-wrap { display: flex; align-items: center; gap: 1rem; }
.actor { font-size: 14px; font-weight: 900; letter-spacing: -0.01em; }
.actor.attacker { color: var(--attack-log-text); }
.actor.victim { color: var(--attack-log-muted); }

.vector-line { position: relative; width: 4rem; display: flex; align-items: center; }
.line-path { height: 2px; width: 100%; background: var(--attack-log-border-strong); border-radius: 1px; }
.line-head { position: absolute; right: -4px; }
.is-success .line-path { background: color-mix(in srgb, var(--color-success) 24%, var(--attack-log-border-strong)); }
.is-success .line-head { color: var(--attack-log-success-text); }

.event-meta { display: flex; align-items: center; gap: 1rem; }
.meta-item { display: flex; align-items: center; gap: 0.5rem; font-size: 11px; font-weight: 700; color: var(--attack-log-muted); }
.meta-divider { width: 3px; height: 3px; border-radius: 50%; background: var(--attack-log-border-strong); }

/* Result Badges */
.event-secondary { display: flex; align-items: center; }
.result-badge { display: flex; align-items: center; gap: 0.65rem; padding: 0.45rem 1rem; border-radius: 0.75rem; font-size: 10px; font-weight: 900; letter-spacing: 0.05em; }
.result-badge.success { background: var(--attack-log-success-surface); color: var(--attack-log-success-text); border: 1px solid var(--attack-log-success-border); }
.result-badge.failed { background: var(--attack-log-surface); color: var(--attack-log-faint); border: 1px solid var(--attack-log-border); }

.score-delta { font-size: 13px; color: var(--attack-log-success-text); margin-left: 0.5rem; border-left: 1px solid var(--attack-log-success-divider); padding-left: 0.75rem; }

.attack-log-export-button {
  --ui-btn-height: 2.25rem;
  --ui-btn-padding: 0 1.25rem;
  --ui-btn-radius: 0.75rem;
  --ui-btn-font-size: var(--font-size-12);
  --ui-btn-font-weight: 800;
  --ui-btn-secondary-border: var(--attack-log-border);
  --ui-btn-secondary-background: var(--attack-log-surface);
  --ui-btn-secondary-color: var(--attack-log-muted);
  --ui-btn-secondary-hover-border: var(--attack-log-border-strong);
  --ui-btn-secondary-hover-background: var(--attack-log-surface-subtle);
  --ui-btn-secondary-hover-color: var(--attack-log-text);
}

.empty-feed { display: flex; flex-direction: column; align-items: center; justify-content: center; padding: 5rem 0; color: var(--attack-log-faint); font-size: 14px; font-weight: 600; }

.custom-scrollbar::-webkit-scrollbar { width: 4px; }
.custom-scrollbar::-webkit-scrollbar-thumb { background: var(--attack-log-border-strong); border-radius: 10px; }
</style>
