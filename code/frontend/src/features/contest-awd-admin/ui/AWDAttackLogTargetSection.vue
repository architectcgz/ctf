<script setup lang="ts">
import type { AdminContestChallengeViewData, AdminContestTeamData } from '@/api/contracts'

import { formatAwdChallengeLabel } from './awdOperationsDialogOptions'

defineProps<{
  teams: AdminContestTeamData[]
  challengeOptions: AdminContestChallengeViewData[]
  form: {
    attacker_team_id: string
    victim_team_id: string
    challenge_id: string
  }
  fieldErrors: {
    attacker_team_id: string
    victim_team_id: string
    challenge_id: string
  }
}>()
</script>

<template>
  <div class="grid gap-4 sm:grid-cols-2">
    <div class="ui-field awd-operations-field">
      <label
        class="ui-field__label"
        for="awd-attack-team"
      >攻击队伍</label>
      <span
        class="ui-control-wrap"
        :class="{ 'is-error': !!fieldErrors.attacker_team_id }"
      >
        <select
          id="awd-attack-team"
          v-model="form.attacker_team_id"
          class="ui-control"
        >
          <option
            value=""
            disabled
          >请选择攻击队伍</option>
          <option
            v-for="team in teams"
            :key="team.id"
            :value="team.id"
          >
            {{ team.name }}
          </option>
        </select>
      </span>
      <p
        v-if="fieldErrors.attacker_team_id"
        class="ui-field__error"
      >
        {{ fieldErrors.attacker_team_id }}
      </p>
    </div>

    <div class="ui-field awd-operations-field">
      <label
        class="ui-field__label"
        for="awd-victim-team"
      >受害队伍</label>
      <span
        class="ui-control-wrap"
        :class="{ 'is-error': !!fieldErrors.victim_team_id }"
      >
        <select
          id="awd-victim-team"
          v-model="form.victim_team_id"
          class="ui-control"
        >
          <option
            value=""
            disabled
          >请选择受害队伍</option>
          <option
            v-for="team in teams"
            :key="team.id"
            :value="team.id"
          >
            {{ team.name }}
          </option>
        </select>
      </span>
      <p
        v-if="fieldErrors.victim_team_id"
        class="ui-field__error"
      >
        {{ fieldErrors.victim_team_id }}
      </p>
    </div>
  </div>

  <div class="grid gap-4 sm:grid-cols-2">
    <div class="ui-field awd-operations-field">
      <label
        class="ui-field__label"
        for="awd-attack-challenge"
      >题目</label>
      <span
        class="ui-control-wrap"
        :class="{ 'is-error': !!fieldErrors.challenge_id }"
      >
        <select
          id="awd-attack-challenge"
          v-model="form.challenge_id"
          class="ui-control"
        >
          <option
            value=""
            disabled
          >请选择题目</option>
          <option
            v-for="challenge in challengeOptions"
            :key="challenge.id"
            :value="challenge.challenge_id"
          >
            {{ formatAwdChallengeLabel(challenge) }}
          </option>
        </select>
      </span>
      <p
        v-if="fieldErrors.challenge_id"
        class="ui-field__error"
      >
        {{ fieldErrors.challenge_id }}
      </p>
    </div>
  </div>
</template>
