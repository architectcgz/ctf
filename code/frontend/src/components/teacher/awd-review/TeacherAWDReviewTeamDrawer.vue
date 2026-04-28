<script setup lang="ts">
import type {
  TeacherAWDReviewAttackItemData,
  TeacherAWDReviewServiceItemData,
  TeacherAWDReviewTeamItemData,
  TeacherAWDReviewTrafficItemData,
} from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'
import AdminSurfaceDrawer from '@/components/common/modal-templates/AdminSurfaceDrawer.vue'
import { formatDate } from '@/utils/format'

defineProps<{
  visible: boolean
  team: TeacherAWDReviewTeamItemData | null
  services: TeacherAWDReviewServiceItemData[]
  attacks: TeacherAWDReviewAttackItemData[]
  traffic: TeacherAWDReviewTrafficItemData[]
}>()

const emit = defineEmits<{
  close: []
}>()

function formatServiceRef(serviceId?: string): string {
  return `Service #${serviceId || '--'}`
}
</script>

<template>
  <AdminSurfaceDrawer
    :open="visible"
    :title="team?.team_name || '队伍详情'"
    subtitle="查看当前队伍在所选轮次内的服务状态、攻击记录和流量证据。"
    eyebrow="Team Focus"
    width="34rem"
    @close="emit('close')"
  >
    <div class="awd-review-drawer teacher-surface-dialog">
      <section
        v-if="team"
        class="awd-review-drawer__summary metric-panel-default-surface"
      >
        <div class="awd-review-drawer__metrics metric-panel-grid metric-panel-default-surface">
          <article class="metric-panel-card">
            <div class="metric-panel-label">
              总分
            </div>
            <div class="metric-panel-value">
              {{ team.total_score }}
            </div>
            <div class="metric-panel-helper">
              队伍当前累计分数
            </div>
          </article>
          <article class="metric-panel-card">
            <div class="metric-panel-label">
              成员数
            </div>
            <div class="metric-panel-value">
              {{ team.member_count }}
            </div>
            <div class="metric-panel-helper">
              当前队伍成员数量
            </div>
          </article>
          <article class="metric-panel-card">
            <div class="metric-panel-label">
              最近命中
            </div>
            <div class="metric-panel-value awd-review-drawer__time">
              {{ team.last_solve_at ? formatDate(team.last_solve_at) : '--' }}
            </div>
            <div class="metric-panel-helper">
              最近一次有效命中时间
            </div>
          </article>
        </div>
      </section>

      <section class="awd-review-drawer__section">
        <div class="awd-review-drawer__section-head">
          <h4>服务视图</h4>
          <span>{{ services.length }} 条</span>
        </div>
        <AppEmpty
          v-if="services.length === 0"
          icon="Server"
          title="暂无服务记录"
          description="当前筛选下还没有可展示的服务状态。"
        />
        <div
          v-else
          class="awd-review-drawer__list"
        >
          <article
            v-for="service in services"
            :key="service.id"
            class="awd-review-drawer__item"
          >
            <div>
              <div class="awd-review-drawer__item-head">
                <strong>{{ service.challenge_title }}</strong>
                <span
                  v-if="service.service_id"
                  class="awd-review-drawer__item-chip"
                  data-testid="awd-review-drawer-service-id"
                >
                  {{ formatServiceRef(service.service_id) }}
                </span>
              </div>
              <p>{{ service.team_name }} · {{ service.service_status }}</p>
            </div>
            <div class="awd-review-drawer__item-meta">
              <span>SLA {{ service.sla_score }}</span>
              <span>Def {{ service.defense_score }}</span>
              <span>Atk {{ service.attack_score }}</span>
            </div>
          </article>
        </div>
      </section>

      <section class="awd-review-drawer__section">
        <div class="awd-review-drawer__section-head">
          <h4>攻击记录</h4>
          <span>{{ attacks.length }} 条</span>
        </div>
        <AppEmpty
          v-if="attacks.length === 0"
          icon="Crosshair"
          title="暂无攻击记录"
          description="当前筛选下还没有与该队伍相关的攻击事件。"
        />
        <div
          v-else
          class="awd-review-drawer__list"
        >
          <article
            v-for="attack in attacks"
            :key="attack.id"
            class="awd-review-drawer__item"
          >
            <div>
              <div class="awd-review-drawer__item-head">
                <strong>{{ attack.attacker_team_name }} → {{ attack.victim_team_name }}</strong>
                <span
                  v-if="attack.service_id"
                  class="awd-review-drawer__item-chip"
                  data-testid="awd-review-drawer-attack-service-id"
                >
                  {{ formatServiceRef(attack.service_id) }}
                </span>
              </div>
              <p>{{ attack.challenge_title }} · {{ attack.attack_type }} · {{ attack.source }}</p>
            </div>
            <div class="awd-review-drawer__item-meta">
              <span>{{ attack.is_success ? '成功' : '失败' }}</span>
              <span>+{{ attack.score_gained }}</span>
            </div>
          </article>
        </div>
      </section>

      <section class="awd-review-drawer__section">
        <div class="awd-review-drawer__section-head">
          <h4>流量证据</h4>
          <span>{{ traffic.length }} 条</span>
        </div>
        <AppEmpty
          v-if="traffic.length === 0"
          icon="Waypoints"
          title="暂无流量证据"
          description="当前筛选下还没有与该队伍相关的流量事件。"
        />
        <div
          v-else
          class="awd-review-drawer__list"
        >
          <article
            v-for="event in traffic"
            :key="event.id"
            class="awd-review-drawer__item"
          >
            <div>
              <div class="awd-review-drawer__item-head">
                <strong>{{ event.method }} {{ event.path }}</strong>
                <span
                  v-if="event.service_id"
                  class="awd-review-drawer__item-chip"
                  data-testid="awd-review-drawer-traffic-service-id"
                >
                  {{ formatServiceRef(event.service_id) }}
                </span>
              </div>
              <p>
                {{ event.attacker_team_name }} → {{ event.victim_team_name }} ·
                {{ event.challenge_title }}
              </p>
            </div>
            <div class="awd-review-drawer__item-meta">
              <span>{{ event.status_code }}</span>
              <span>{{ event.source }}</span>
            </div>
          </article>
        </div>
      </section>
    </div>
  </AdminSurfaceDrawer>
</template>

<style scoped>
.awd-review-drawer {
  overflow-y: auto;
  padding-top: var(--space-1);
}

.awd-review-drawer__summary {
  margin-top: var(--space-6);
}

.awd-review-drawer__metrics {
  margin-top: var(--space-4);
}

.awd-review-drawer__time {
  font-size: var(--font-size-0-90);
}

.awd-review-drawer__section {
  margin-top: var(--space-6);
  padding-top: var(--space-5);
  border-top: 1px dashed var(--teacher-divider);
}

.awd-review-drawer__section-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
  margin-bottom: var(--space-4);
  color: var(--journal-ink);
}

.awd-review-drawer__section-head h4 {
  margin: 0;
  font-size: var(--font-size-1-00);
  font-weight: 700;
}

.awd-review-drawer__section-head span {
  color: var(--journal-muted);
  font-size: var(--font-size-0-82);
}

.awd-review-drawer__list {
  display: grid;
  gap: var(--space-3);
}

.awd-review-drawer__item {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-4);
  padding: var(--space-4);
  border: 1px solid var(--teacher-card-border);
  border-radius: 16px;
  background: color-mix(in srgb, var(--journal-surface-subtle) 82%, transparent);
}

.awd-review-drawer__item-head {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-start;
  gap: var(--space-2);
}

.awd-review-drawer__item strong {
  display: block;
  color: var(--journal-ink);
}

.awd-review-drawer__item-chip {
  display: inline-flex;
  align-items: center;
  min-height: 1.55rem;
  padding: 0 var(--space-2-5);
  border-radius: 999px;
  background: color-mix(in srgb, var(--journal-accent) 10%, transparent);
  color: var(--journal-accent-strong);
  font-family: var(--font-family-mono);
  font-size: var(--font-size-0-74);
  font-weight: 700;
  white-space: nowrap;
}

.awd-review-drawer__item p {
  margin: var(--space-1-5) 0 0;
  color: var(--journal-muted);
  line-height: 1.6;
}

.awd-review-drawer__item-meta {
  display: grid;
  gap: var(--space-1);
  justify-items: end;
  color: var(--journal-muted);
  font-size: var(--font-size-0-80);
  text-align: right;
}

@media (max-width: 720px) {
  .awd-review-drawer {
    padding-top: 0;
  }

  .awd-review-drawer__item {
    flex-direction: column;
  }

  .awd-review-drawer__item-meta {
    justify-items: start;
    text-align: left;
  }
}
</style>
