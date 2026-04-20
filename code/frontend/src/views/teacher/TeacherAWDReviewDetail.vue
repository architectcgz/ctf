<script setup lang="ts">
import { computed, ref } from 'vue'
import { 
  ArrowLeft, 
  Download, 
  FileDown, 
  Shield, 
  Waypoints, 
  Clock, 
  Target, 
  Zap, 
  Activity,
  ChevronRight,
  TrendingUp,
  Award
} from 'lucide-vue-next'

import AppEmpty from '@/components/common/AppEmpty.vue'
import TeacherAWDReviewTeamDrawer from '@/components/teacher/awd-review/TeacherAWDReviewTeamDrawer.vue'
import { useTeacherAwdReviewDetail } from '@/composables/useTeacherAwdReviewDetail'
import { formatDate } from '@/utils/format'

const {
  router,
  polling,
  loading,
  error,
  review,
  exporting,
  selectedRoundNumber,
  selectedRound,
  selectedTeam,
  selectedTeamServices,
  selectedTeamAttacks,
  selectedTeamTraffic,
  canExportReport,
  loadReview,
  setRound,
  openTeam,
  closeTeam,
  exportArchive,
  exportReport,
} = useTeacherAwdReviewDetail()

const activeTitle = computed(() => review.value?.contest.title || 'AWD复盘')
const activeSummaryTitle = computed(() =>
  selectedRoundNumber.value ? `第 ${selectedRoundNumber.value} 轮` : '整场总览'
)

const summaryStats = computed(() => {
  if (selectedRound.value) {
    return {
      roundCount: 1,
      teamCount: selectedRound.value.teams.length,
      serviceCount: selectedRound.value.round.service_count,
      attackCount: selectedRound.value.round.attack_count,
      trafficCount: selectedRound.value.round.traffic_count,
    }
  }

  return {
    roundCount: review.value?.overview?.round_count ?? 0,
    teamCount: review.value?.overview?.team_count ?? 0,
    serviceCount: review.value?.overview?.service_count ?? 0,
    attackCount: review.value?.overview?.attack_count ?? 0,
    trafficCount: review.value?.overview?.traffic_count ?? 0,
  }
})

function contestStatusLabel(status: string): string {
  switch (status) {
    case 'running': return '进行中'
    case 'ended': return '已结束'
    case 'frozen': return '冻结中'
    default: return status || '未开始'
  }
}

const timelineRounds = computed(() => review.value?.rounds || [])
</script>

<template>
  <div class="teacher-review-workspace">
    <!-- 1. Academy Header -->
    <header class="academy-header">
      <div class="academy-header__identity">
        <div class="academy-overline">AWD Instructional Review / Academy Workspace</div>
        <div class="flex items-center gap-4">
          <h1 class="academy-title">{{ activeTitle }}</h1>
          <div class="academy-badge" :class="review?.contest.status">
            {{ contestStatusLabel(review?.contest.status || '') }}
          </div>
        </div>
        <p class="academy-description">
          多维复盘攻防实战过程。通过轮次下钻与流量回溯，协助教师评估学生的防御加固能力与漏洞挖掘表现。
        </p>
      </div>

      <div class="academy-header__actions">
        <button type="button" class="academy-btn academy-btn--neutral" @click="router.push({ name: 'TeacherAWDReviewIndex' })">
          <ArrowLeft class="h-3.5 w-3.5" />
          <span>返回列表</span>
        </button>
        <button type="button" class="academy-btn academy-btn--neutral" :disabled="loading || !review || exporting === 'archive'" @click="exportArchive">
          <Download class="h-3.5 w-3.5" />
          <span>归档导出</span>
        </button>
        <button type="button" class="academy-btn academy-btn--primary" :disabled="loading || !review || exporting === 'report' || !canExportReport" @click="exportReport">
          <FileDown class="h-3.5 w-3.5" />
          <span>生成评估报告</span>
        </button>
      </div>
    </header>

    <!-- 2. Insight Strip -->
    <section class="academy-insight-strip">
      <div class="insight-card">
        <div class="insight-icon insight-icon--blue"><TrendingUp class="h-4 w-4" /></div>
        <div class="insight-body">
          <div class="insight-label">视图焦点</div>
          <div class="insight-value">{{ activeSummaryTitle }}</div>
        </div>
      </div>
      <div class="insight-card">
        <div class="insight-icon insight-icon--emerald"><Award class="h-4 w-4" /></div>
        <div class="insight-body">
          <div class="insight-label">参与规模</div>
          <div class="insight-value">{{ summaryStats.teamCount }} <small>TEAMS</small> / {{ summaryStats.roundCount }} <small>ROUNDS</small></div>
        </div>
      </div>
      <div class="insight-card">
        <div class="insight-icon insight-icon--purple"><Zap class="h-4 w-4" /></div>
        <div class="insight-body">
          <div class="insight-label">证据总量 (SRV/ATK/TRF)</div>
          <div class="insight-value">{{ summaryStats.serviceCount }} / {{ summaryStats.attackCount }} / {{ summaryStats.trafficCount }}</div>
        </div>
      </div>
      <div class="insight-card">
        <div class="insight-icon insight-icon--slate"><Clock class="h-4 w-4" /></div>
        <div class="insight-body">
          <div class="insight-label">导出状态</div>
          <div class="insight-value text-slate-500">{{ polling ? '后台处理中...' : '链路就绪' }}</div>
        </div>
      </div>
    </section>

    <!-- 3. Timeline Scroller -->
    <nav class="academy-timeline">
      <button 
        class="timeline-node" 
        :class="{ active: !selectedRoundNumber }"
        @click="setRound(undefined)"
      >
        <div class="node-circle"></div>
        <div class="node-label">整场总览</div>
      </button>
      <div class="timeline-line"></div>
      <div class="timeline-nodes custom-scrollbar">
        <button 
          v-for="round in timelineRounds" 
          :key="round.id"
          class="timeline-node"
          :class="{ active: selectedRoundNumber === round.round_number }"
          @click="setRound(round.round_number)"
        >
          <div class="node-circle"></div>
          <div class="node-label">R{{ round.round_number }}</div>
        </button>
      </div>
    </nav>

    <!-- 4. Main Canvas -->
    <div v-if="loading" class="academy-loading-canvas">
      <div class="academy-spinner"></div>
      <p>正在载入复盘分析数据...</p>
    </div>

    <AppEmpty
      v-else-if="error"
      title="复盘详情加载失败"
      :description="error"
      icon="AlertCircle"
      class="academy-empty"
    >
      <template #action>
        <button type="button" class="academy-btn academy-btn--primary" @click="loadReview">重新加载</button>
      </template>
    </AppEmpty>

    <div v-else-if="review" class="academy-canvas">
      <!-- Round Overview / Team Analysis -->
      <section class="academy-section">
        <header class="section-header">
          <h3 class="section-title">{{ activeSummaryTitle }} 表现分析</h3>
          <div class="section-meta">共 {{ summaryStats.teamCount }} 支参与队伍</div>
        </header>

        <div v-if="!selectedRound" class="round-card-grid">
          <article v-for="round in review.rounds" :key="round.id" class="round-summary-card">
            <div class="round-summary-card__header">
              <div class="round-number">ROUND {{ round.round_number }}</div>
              <button type="button" class="academy-btn academy-btn--ghost academy-btn--xs" @click="setRound(round.round_number)">
                下钻分析 <ChevronRight class="h-3 w-3" />
              </button>
            </div>
            <div class="round-summary-card__metrics">
              <div class="metric-item">
                <div class="metric-label">服务</div>
                <div class="metric-value">{{ round.service_count }}</div>
              </div>
              <div class="metric-item">
                <div class="metric-label">攻击</div>
                <div class="metric-value">{{ round.attack_count }}</div>
              </div>
              <div class="metric-item">
                <div class="metric-label">流量</div>
                <div class="metric-value">{{ round.traffic_count }}</div>
              </div>
            </div>
            <div class="round-summary-card__footer">
              {{ round.started_at ? formatDate(round.started_at) : '未开始' }}
            </div>
          </article>
        </div>

        <div v-else class="academy-team-grid">
          <table class="academy-table">
            <thead>
              <tr>
                <th>队伍</th>
                <th>得分表现</th>
                <th>命中记录</th>
                <th>成员结构</th>
                <th class="text-right">操作</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="team in selectedRound.teams" :key="team.team_id" class="group">
                <td class="font-bold text-slate-900">
                  <div class="flex items-center gap-3">
                    <div class="h-2 w-2 rounded-full bg-emerald-400"></div>
                    {{ team.team_name }}
                  </div>
                </td>
                <td class="font-mono text-emerald-600 font-bold">
                  {{ team.total_score }} <small class="text-slate-400 font-sans ml-1">pts</small>
                </td>
                <td class="text-xs text-slate-500">
                  {{ team.last_solve_at ? formatDate(team.last_solve_at) : '暂无命中' }}
                </td>
                <td class="text-xs text-slate-500">
                  {{ team.member_count }} 成员 (UID: {{ team.captain_id }})
                </td>
                <td class="text-right">
                  <button type="button" class="academy-btn academy-btn--ghost academy-btn--xs" @click="openTeam(team)">
                    调阅细节
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>

      <!-- Evidence Grid -->
      <section v-if="selectedRound" class="academy-evidence-grid">
        <article class="evidence-panel">
          <header class="panel-header">
            <Activity class="h-4 w-4 text-emerald-600" />
            <span>服务运行</span>
          </header>
          <div class="evidence-list custom-scrollbar">
            <div v-for="service in selectedRound.services" :key="service.id" class="evidence-item">
              <div class="evidence-item__title">{{ service.team_name }} · {{ service.challenge_title }}</div>
              <div class="evidence-item__meta">Status: {{ service.service_status }} · SLA: {{ service.sla_score }}</div>
            </div>
            <AppEmpty v-if="selectedRound.services.length === 0" icon="Shield" title="无服务数据" class="compact-empty" />
          </div>
        </article>

        <article class="evidence-panel">
          <header class="panel-header">
            <Sword class="h-4 w-4 text-red-600" />
            <span>攻击记录</span>
          </header>
          <div class="evidence-list custom-scrollbar">
            <div v-for="attack in selectedRound.attacks" :key="attack.id" class="evidence-item">
              <div class="evidence-item__title">{{ attack.attacker_team_name }} → {{ attack.victim_team_name }}</div>
              <div class="evidence-item__meta">{{ attack.challenge_title }} · {{ attack.attack_type }}</div>
            </div>
            <AppEmpty v-if="selectedRound.attacks.length === 0" icon="Target" title="无攻击记录" class="compact-empty" />
          </div>
        </article>

        <article class="evidence-panel">
          <header class="panel-header">
            <Waypoints class="h-4 w-4 text-blue-600" />
            <span>流量审计</span>
          </header>
          <div class="evidence-list custom-scrollbar">
            <div v-for="event in selectedRound.traffic" :key="event.id" class="evidence-item">
              <div class="evidence-item__title">{{ event.method }} {{ event.path }}</div>
              <div class="evidence-item__meta">{{ event.attacker_team_name }} → {{ event.victim_team_name }} · {{ event.status_code }}</div>
            </div>
            <AppEmpty v-if="selectedRound.traffic.length === 0" icon="Activity" title="无流量证据" class="compact-empty" />
          </div>
        </article>
      </section>
    </div>

    <TeacherAWDReviewTeamDrawer
      :visible="Boolean(selectedTeam)"
      :team="selectedTeam"
      :services="selectedTeamServices"
      :attacks="selectedTeamAttacks"
      :traffic="selectedTeamTraffic"
      @close="closeTeam"
    />
  </div>
</template>

<style scoped>
.teacher-review-workspace {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
  padding: 2rem;
  background: #fdfdfd;
  min-height: 100%;
}

.academy-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  padding-bottom: 1.5rem;
  border-bottom: 1px solid #edf2f7;
}

.academy-overline {
  font-size: 10px;
  font-weight: 800;
  text-transform: uppercase;
  letter-spacing: 0.15em;
  color: #a0aec0;
  margin-bottom: 0.5rem;
}

.academy-title {
  font-size: 1.75rem;
  font-weight: 900;
  color: #2d3748;
  margin: 0;
}

.academy-badge {
  padding: 0.25rem 0.75rem;
  border-radius: 999px;
  font-size: 11px;
  font-weight: 800;
  background: #edf2f7;
  color: #4a5568;
}

.academy-badge.running { background: #ebf8ff; color: #2b6cb0; }
.academy-badge.ended { background: #f7fafc; color: #718096; }

.academy-description {
  margin: 0.75rem 0 0;
  font-size: 14px;
  color: #718096;
  max-width: 45rem;
  line-height: 1.6;
}

.academy-header__actions {
  display: flex;
  gap: 0.75rem;
}

.academy-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  height: 2.5rem;
  padding: 0 1.25rem;
  border-radius: 0.75rem;
  font-size: 13px;
  font-weight: 700;
  transition: all 0.2s ease;
  cursor: pointer;
}

.academy-btn--neutral {
  background: white;
  border: 1px solid #e2e8f0;
  color: #4a5568;
}

.academy-btn--neutral:hover {
  background: #f7fafc;
  border-color: #cbd5e0;
}

.academy-btn--primary {
  background: #3182ce;
  color: white;
  border: none;
}

.academy-btn--primary:hover {
  background: #2b6cb0;
  transform: translateY(-1px);
}

.academy-btn--ghost {
  background: transparent;
  color: #3182ce;
  border: none;
}

.academy-btn--xs {
  height: 1.75rem;
  padding: 0 0.75rem;
  font-size: 11px;
}

.academy-insight-strip {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 1.25rem;
}

.insight-card {
  background: white;
  border: 1px solid #edf2f7;
  border-radius: 1rem;
  padding: 1.25rem;
  display: flex;
  align-items: center;
  gap: 1.25rem;
  box-shadow: 0 1px 3px rgba(0,0,0,0.02);
}

.insight-icon {
  width: 2.75rem;
  height: 2.75rem;
  border-radius: 0.75rem;
  display: flex;
  align-items: center;
  justify-content: center;
}

.insight-icon--blue { background: #ebf8ff; color: #3182ce; }
.insight-icon--emerald { background: #f0fff4; color: #38a169; }
.insight-icon--purple { background: #faf5ff; color: #805ad5; }
.insight-icon--slate { background: #f7fafc; color: #718096; }

.insight-label {
  font-size: 10px;
  font-weight: 800;
  text-transform: uppercase;
  color: #a0aec0;
}

.insight-value {
  font-size: 1.15rem;
  font-weight: 900;
  color: #2d3748;
}

.insight-value small {
  font-size: 10px;
  color: #a0aec0;
}

.academy-timeline {
  display: flex;
  align-items: center;
  gap: 1.5rem;
  background: white;
  padding: 1rem 1.5rem;
  border: 1px solid #edf2f7;
  border-radius: 1rem;
}

.timeline-node {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.5rem;
  cursor: pointer;
  transition: all 0.2s ease;
  min-width: 4rem;
}

.node-circle {
  width: 0.75rem;
  height: 0.75rem;
  border-radius: 50%;
  border: 2px solid #e2e8f0;
  background: white;
  transition: all 0.2s ease;
}

.node-label {
  font-size: 11px;
  font-weight: 700;
  color: #718096;
}

.timeline-node.active .node-circle {
  border-color: #3182ce;
  background: #3182ce;
  transform: scale(1.2);
}

.timeline-node.active .node-label {
  color: #2b6cb0;
}

.timeline-line {
  width: 2rem;
  height: 2px;
  background: #edf2f7;
}

.timeline-nodes {
  flex: 1;
  display: flex;
  gap: 1.5rem;
  overflow-x: auto;
  padding: 0.25rem 0;
}

.academy-section {
  background: white;
  border: 1px solid #edf2f7;
  border-radius: 1rem;
  padding: 1.5rem;
}

.section-header {
  margin-bottom: 1.5rem;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.section-title {
  font-size: 1.15rem;
  font-weight: 900;
  color: #2d3748;
}

.section-meta {
  font-size: 12px;
  color: #a0aec0;
}

.round-card-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(16rem, 1fr));
  gap: 1.25rem;
}

.round-summary-card {
  border: 1px solid #edf2f7;
  border-radius: 1rem;
  padding: 1.25rem;
  transition: all 0.2s ease;
}

.round-summary-card:hover {
  border-color: #3182ce;
  box-shadow: 0 4px 12px rgba(0,0,0,0.03);
}

.round-summary-card__header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
}

.round-number {
  font-size: 13px;
  font-weight: 900;
  color: #4a5568;
}

.round-summary-card__metrics {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 0.75rem;
  margin-bottom: 1rem;
}

.metric-label {
  font-size: 10px;
  color: #a0aec0;
  text-transform: uppercase;
}

.metric-value {
  font-size: 1rem;
  font-weight: 800;
  color: #2d3748;
}

.round-summary-card__footer {
  font-size: 11px;
  color: #cbd5e0;
}

.academy-table {
  width: 100%;
  border-collapse: collapse;
}

.academy-table th {
  text-align: left;
  padding: 0.75rem 1rem;
  font-size: 10px;
  font-weight: 800;
  text-transform: uppercase;
  color: #a0aec0;
  border-bottom: 2px solid #edf2f7;
}

.academy-table td {
  padding: 1rem;
  border-bottom: 1px solid #f7fafc;
}

.academy-evidence-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 1.5rem;
}

.evidence-panel {
  background: white;
  border: 1px solid #edf2f7;
  border-radius: 1rem;
  padding: 1.25rem;
  display: flex;
  flex-direction: column;
  height: 30rem;
}

.panel-header {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding-bottom: 1rem;
  border-bottom: 1px solid #f7fafc;
  margin-bottom: 1rem;
  font-size: 14px;
  font-weight: 900;
  color: #2d3748;
}

.evidence-list {
  flex: 1;
  overflow-y: auto;
}

.evidence-item {
  padding: 0.85rem;
  border-radius: 0.75rem;
  border: 1px solid transparent;
  transition: all 0.2s ease;
}

.evidence-item:hover {
  background: #f7fafc;
  border-color: #edf2f7;
}

.evidence-item__title {
  font-size: 13px;
  font-weight: 700;
  color: #2d3748;
}

.evidence-item__meta {
  font-size: 11px;
  color: #718096;
  margin-top: 0.25rem;
}

.compact-empty :deep(.empty-icon) {
  width: 2rem;
  height: 2rem;
}

.academy-loading-canvas {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 1rem;
  color: #718096;
}

.academy-spinner {
  width: 2.5rem;
  height: 2.5rem;
  border: 3px solid #edf2f7;
  border-top-color: #3182ce;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}
</style>

