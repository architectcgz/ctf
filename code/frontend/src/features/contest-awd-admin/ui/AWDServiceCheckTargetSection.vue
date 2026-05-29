<script setup lang="ts">
import type {
  AdminContestChallengeViewData,
  AdminContestTeamData,
  AWDTeamServiceData,
} from '@/api/contracts'

import { formatAwdChallengeLabel } from './awdOperationsDialogOptions'

defineProps<{
  teams: AdminContestTeamData[]
  challengeOptions: AdminContestChallengeViewData[]
  form: {
    team_id: string
    challenge_id: string
    service_status: AWDTeamServiceData['service_status']
  }
  fieldErrors: {
    team_id: string
    challenge_id: string
  }
}>()
</script>

<template>
  <div class="grid gap-4 sm:grid-cols-2">
    <div class="ui-field awd-operations-field">
      <label
        class="ui-field__label"
        for="awd-service-team"
      >队伍</label>
      <span
        class="ui-control-wrap"
        :class="{ 'is-error': !!fieldErrors.team_id }"
      >
        <select
          id="awd-service-team"
          v-model="form.team_id"
          class="ui-control"
        >
          <option
            value=""
            disabled
          >请选择队伍</option>
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
        v-if="fieldErrors.team_id"
        class="ui-field__error"
      >
        {{ fieldErrors.team_id }}
      </p>
    </div>

    <div class="ui-field awd-operations-field">
      <label
        class="ui-field__label"
        for="awd-service-challenge"
      >题目</label>
      <span
        class="ui-control-wrap"
        :class="{ 'is-error': !!fieldErrors.challenge_id }"
      >
        <select
          id="awd-service-challenge"
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

  <div class="ui-field awd-operations-field">
    <label
      class="ui-field__label"
      for="awd-service-status"
    >服务状态</label>
    <span class="ui-control-wrap">
      <select
        id="awd-service-status"
        v-model="form.service_status"
        class="ui-control"
      >
        <option value="up">正常</option>
        <option value="down">下线</option>
        <option value="compromised">已失陷</option>
      </select>
    </span>
  </div>
</template>
