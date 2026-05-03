<script setup lang="ts">
import { Activity, Target, Waypoints } from 'lucide-vue-next'

import AppEmpty from '@/components/common/AppEmpty.vue'

interface ServiceEvidenceItem {
  id: string | number
  team_name: string
  challenge_title: string
  service_id?: string
  service_status: string
  sla_score: number
}

interface AttackEvidenceItem {
  id: string | number
  attacker_team_name: string
  victim_team_name: string
  service_id?: string
  challenge_title: string
  attack_type: string
}

interface TrafficEvidenceItem {
  id: string | number
  method: string
  path: string
  service_id?: string
  attacker_team_name: string
  victim_team_name: string
  status_code: string | number
}

interface SelectedRoundEvidence {
  services: ServiceEvidenceItem[]
  attacks: AttackEvidenceItem[]
  traffic: TrafficEvidenceItem[]
}

defineProps<{
  selectedRound: SelectedRoundEvidence
  formatServiceRef: (serviceId?: string) => string
}>()
</script>

<template>
  <section class="awd-review-evidence-grid">
    <article class="workspace-directory-section teacher-directory-section awd-review-evidence-panel">
      <header class="list-heading awd-review-evidence-head">
        <div>
          <div class="journal-note-label">
            Service Evidence
          </div>
          <h3 class="list-heading__title">
            服务运行
          </h3>
        </div>
        <Activity class="awd-review-evidence-icon awd-review-evidence-icon--service h-4 w-4" />
      </header>
      <div class="awd-review-evidence-list custom-scrollbar">
        <AppEmpty
          v-if="selectedRound.services.length === 0"
          icon="Shield"
          title="无服务数据"
          class="teacher-empty-state workspace-directory-empty awd-review-compact-empty"
        />
        <article
          v-for="service in selectedRound.services"
          v-else
          :key="service.id"
          class="awd-review-evidence-item"
        >
          <div class="awd-review-evidence-item__head">
            <strong>{{ service.team_name }} · {{ service.challenge_title }}</strong>
            <span
              v-if="service.service_id"
              data-testid="awd-review-service-id"
              class="teacher-directory-chip awd-review-service-chip"
            >
              {{ formatServiceRef(service.service_id) }}
            </span>
          </div>
          <p>{{ service.service_status }} · SLA {{ service.sla_score }}</p>
        </article>
      </div>
    </article>

    <article class="workspace-directory-section teacher-directory-section awd-review-evidence-panel">
      <header class="list-heading awd-review-evidence-head">
        <div>
          <div class="journal-note-label">
            Attack Evidence
          </div>
          <h3 class="list-heading__title">
            攻击记录
          </h3>
        </div>
        <Target class="awd-review-evidence-icon awd-review-evidence-icon--attack h-4 w-4" />
      </header>
      <div class="awd-review-evidence-list custom-scrollbar">
        <AppEmpty
          v-if="selectedRound.attacks.length === 0"
          icon="Target"
          title="无攻击记录"
          class="teacher-empty-state workspace-directory-empty awd-review-compact-empty"
        />
        <article
          v-for="attack in selectedRound.attacks"
          v-else
          :key="attack.id"
          class="awd-review-evidence-item"
        >
          <div class="awd-review-evidence-item__head">
            <strong>{{ attack.attacker_team_name }} → {{ attack.victim_team_name }}</strong>
            <span
              v-if="attack.service_id"
              data-testid="awd-review-attack-service-id"
              class="teacher-directory-chip awd-review-service-chip"
            >
              {{ formatServiceRef(attack.service_id) }}
            </span>
          </div>
          <p>{{ attack.challenge_title }} · {{ attack.attack_type }}</p>
        </article>
      </div>
    </article>

    <article class="workspace-directory-section teacher-directory-section awd-review-evidence-panel">
      <header class="list-heading awd-review-evidence-head">
        <div>
          <div class="journal-note-label">
            Traffic Evidence
          </div>
          <h3 class="list-heading__title">
            流量审计
          </h3>
        </div>
        <Waypoints class="awd-review-evidence-icon awd-review-evidence-icon--traffic h-4 w-4" />
      </header>
      <div class="awd-review-evidence-list custom-scrollbar">
        <AppEmpty
          v-if="selectedRound.traffic.length === 0"
          icon="Activity"
          title="无流量证据"
          class="teacher-empty-state workspace-directory-empty awd-review-compact-empty"
        />
        <article
          v-for="event in selectedRound.traffic"
          v-else
          :key="event.id"
          class="awd-review-evidence-item"
        >
          <div class="awd-review-evidence-item__head">
            <strong>{{ event.method }} {{ event.path }}</strong>
            <span
              v-if="event.service_id"
              data-testid="awd-review-traffic-service-id"
              class="teacher-directory-chip awd-review-service-chip"
            >
              {{ formatServiceRef(event.service_id) }}
            </span>
          </div>
          <p>{{ event.attacker_team_name }} → {{ event.victim_team_name }} · {{ event.status_code }}</p>
        </article>
      </div>
    </article>
  </section>
</template>

<style scoped>
.awd-review-evidence-panel {
  gap: var(--space-4);
  border-color: var(--awd-review-line);
  background: color-mix(in srgb, var(--awd-review-surface-subtle) 76%, transparent);
}

.awd-review-evidence-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: var(--space-4);
}

.awd-review-evidence-head {
  align-items: center;
}

.awd-review-evidence-icon {
  color: var(--awd-review-faint);
}

.awd-review-evidence-icon--service {
  color: color-mix(in srgb, var(--awd-review-primary) 72%, transparent);
}

.awd-review-evidence-icon--attack {
  color: color-mix(in srgb, var(--awd-review-danger) 72%, transparent);
}

.awd-review-evidence-icon--traffic {
  color: color-mix(in srgb, var(--awd-review-blue) 72%, transparent);
}

.awd-review-evidence-list {
  display: grid;
  gap: var(--space-3);
  max-height: 22rem;
  overflow: auto;
  padding-right: var(--space-1);
}

.awd-review-evidence-item {
  display: grid;
  gap: var(--space-2);
  padding: var(--space-3) var(--space-3-5);
  border-radius: var(--radius-lg);
  border: 1px solid color-mix(in srgb, var(--awd-review-line) 78%, transparent);
  background: color-mix(in srgb, var(--awd-review-surface) 90%, transparent);
}

.awd-review-evidence-item__head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: var(--space-2);
}

.awd-review-evidence-item__head strong {
  font-size: var(--font-size-0-88);
  color: var(--awd-review-text);
}

.awd-review-evidence-item p {
  margin: 0;
  font-size: var(--font-size-0-82);
  color: var(--awd-review-muted);
}

.awd-review-service-chip {
  white-space: nowrap;
}

.awd-review-compact-empty {
  padding: var(--space-4);
}

@media (max-width: 1024px) {
  .awd-review-evidence-grid {
    grid-template-columns: repeat(1, minmax(0, 1fr));
  }

  .awd-review-evidence-item__head {
    flex-direction: column;
    align-items: flex-start;
  }
}
</style>
