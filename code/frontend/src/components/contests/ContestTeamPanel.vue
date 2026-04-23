<script setup lang="ts">
import type { TeamData } from '@/api/contracts'

const props = withDefaults(
  defineProps<{
    team: TeamData | null
    isCaptain?: boolean
  }>(),
  {
    isCaptain: false,
  }
)

const emit = defineEmits<{
  createTeam: []
  joinTeam: []
  kickMember: [userId: string]
}>()
</script>

<template>
  <div
    v-if="!props.team"
    class="team-empty"
  >
    <div class="contest-inline-note">
      当前账号尚未加入队伍。
    </div>
    <div class="team-actions">
      <button
        type="button"
        class="ui-btn ui-btn--primary"
        @click="emit('createTeam')"
      >
        创建队伍
      </button>
      <button
        type="button"
        class="ui-btn ui-btn--ghost"
        @click="emit('joinTeam')"
      >
        加入队伍
      </button>
    </div>
  </div>

  <div
    v-else
    class="team-member-list"
  >
    <div
      v-for="member in props.team.members"
      :key="member.user_id"
      class="team-member"
    >
      <span class="team-member__name">{{ member.username }}</span>
      <div class="team-member__actions">
        <span
          v-if="member.user_id === props.team.captain_user_id"
          class="team-member__captain"
        >
          队长
        </span>
        <button
          v-if="props.isCaptain && member.user_id !== props.team.captain_user_id"
          type="button"
          class="team-member__kick"
          @click="emit('kickMember', member.user_id)"
        >
          踢出
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.contest-inline-note {
  border-inline-start: 2px solid color-mix(in srgb, var(--color-border-default) 84%, transparent);
  padding-inline-start: 0.85rem;
  font-size: var(--font-size-0-88);
  line-height: 1.7;
  color: var(--color-text-secondary);
}

.team-empty {
  margin-top: 1rem;
  display: grid;
  gap: 0.9rem;
}

.team-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.6rem;
}

.team-member-list {
  margin-top: 1rem;
  display: grid;
  gap: 0.55rem;
}

.team-member {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 0.45rem 1rem;
  border-bottom: 1px solid color-mix(in srgb, var(--color-border-default) 82%, transparent);
  padding-bottom: 0.75rem;
}

.team-member__name {
  font-size: var(--font-size-0-90);
  color: var(--color-text-primary);
}

.team-member__actions {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
}

.team-member__captain {
  border-radius: 999px;
  border: 1px solid color-mix(in srgb, var(--contest-accent) 28%, transparent);
  background: color-mix(in srgb, var(--contest-accent) 10%, transparent);
  padding: 0.2rem 0.55rem;
  font-size: var(--font-size-0-72);
  font-weight: 700;
  color: color-mix(in srgb, var(--contest-accent) 84%, var(--color-text-primary));
}

.team-member__kick {
  border: 0;
  background: transparent;
  padding: 0;
  font-size: var(--font-size-0-78);
  color: color-mix(in srgb, var(--color-danger) 88%, var(--color-text-primary));
}

.team-member__kick:hover,
.team-member__kick:focus-visible {
  text-decoration: underline;
  outline: none;
}
</style>
