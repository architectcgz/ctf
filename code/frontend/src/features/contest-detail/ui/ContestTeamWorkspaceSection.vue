<template>
  <section
    :id="panelId"
    class="workspace-panel"
    role="tabpanel"
    :aria-labelledby="tabId"
  >
    <section class="contest-section">
      <div class="contest-section__head workspace-tab-heading">
        <div class="workspace-tab-heading__main">
          <div class="workspace-overline">
            Team
          </div>
          <h2 class="contest-section__title workspace-tab-heading__title">
            队伍
          </h2>
        </div>
        <div class="contest-section__hint">
          {{ memberCount }} 人
        </div>
      </div>

      <ContestTeamPanel
        v-if="!team"
        :team="null"
        @create-team="emit('createTeam')"
        @join-team="emit('joinTeam')"
      />

      <div
        v-else
        class="team-board"
      >
        <div class="team-summary">
          <div>
            <div class="workspace-overline">
              Current Team
            </div>
            <h3 class="team-summary__name">
              {{ team.name }}
            </h3>
          </div>
          <span
            v-if="team.invite_code"
            class="team-summary__invite"
          >邀请码: {{ team.invite_code }}</span>
        </div>

        <ContestTeamPanel
          :team="team"
          :is-captain="isCaptain"
          @kick-member="emit('kickMember', $event)"
        />
      </div>
    </section>
  </section>
</template>

<style scoped>
.contest-section__head {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-end;
  justify-content: space-between;
  gap: var(--space-3);
}

.contest-section__title:not(.workspace-tab-heading__title) {
  margin-top: var(--space-1-5);
  font-size: var(--font-size-1-10);
  font-weight: 700;
  color: var(--color-text-primary);
}

.contest-section__hint {
  font-size: var(--font-size-0-82);
  color: var(--color-text-secondary);
}

.team-board {
  margin-top: 1rem;
}

.team-summary {
  display: flex;
  flex-wrap: wrap;
  align-items: baseline;
  justify-content: space-between;
  gap: 0.6rem 1rem;
}

.team-summary__name {
  margin-top: 0.35rem;
  font-size: var(--font-size-1-10);
  font-weight: 700;
  color: var(--color-text-primary);
}

.team-summary__invite {
  font-size: var(--font-size-0-78);
  color: var(--color-text-secondary);
}
</style>

<script setup lang="ts">
import type { TeamData } from '@/api/contracts'

import ContestTeamPanel from './ContestTeamPanel.vue'

defineProps<{
  panelId: string
  tabId: string
  team: TeamData | null
  memberCount: number
  isCaptain: boolean
}>()

const emit = defineEmits<{
  createTeam: []
  joinTeam: []
  kickMember: [userId: string]
}>()
</script>
