<script setup lang="ts">
import { computed, reactive, watch } from 'vue'

import AdminSurfaceModal from '@/components/common/modal-templates/AdminSurfaceModal.vue'
import type {
  AdminContestChallengeData,
  AdminContestTeamData,
  AWDAttackLogData,
} from '@/api/contracts'

const props = defineProps<{
  open: boolean
  teams: AdminContestTeamData[]
  challengeLinks: AdminContestChallengeData[]
  saving: boolean
}>()

const emit = defineEmits<{
  'update:open': [value: boolean]
  save: [
    value: {
      attacker_team_id: number
      victim_team_id: number
      challenge_id: number
      attack_type: AWDAttackLogData['attack_type']
      submitted_flag?: string
      is_success: boolean
    },
  ]
}>()

const form = reactive({
  attacker_team_id: '',
  victim_team_id: '',
  challenge_id: '',
  attack_type: 'flag_capture' as AWDAttackLogData['attack_type'],
  submitted_flag: '',
  is_success: true,
})

const fieldErrors = reactive({
  attacker_team_id: '',
  victim_team_id: '',
  challenge_id: '',
})

const challengeOptions = computed(() =>
  [...props.challengeLinks].sort(
    (a, b) => a.order - b.order || Number(a.challenge_id) - Number(b.challenge_id)
  )
)
const hasTargets = computed(() => props.teams.length >= 2 && challengeOptions.value.length > 0)

function getChallengeLabel(challenge: AdminContestChallengeData): string {
  const prefix = challenge.title?.trim()
    ? challenge.title.trim()
    : `Challenge #${challenge.challenge_id}`
  return `${prefix} · ${challenge.is_visible ? '可见' : '隐藏'}`
}

watch(
  () => props.open,
  (open) => {
    if (!open) {
      return
    }
    form.attacker_team_id = props.teams[0]?.id || ''
    form.victim_team_id = props.teams[1]?.id || props.teams[0]?.id || ''
    form.challenge_id = challengeOptions.value[0]?.challenge_id || ''
    form.attack_type = 'flag_capture'
    form.submitted_flag = ''
    form.is_success = true
    clearErrors()
  },
  { immediate: true }
)

function clearErrors() {
  fieldErrors.attacker_team_id = ''
  fieldErrors.victim_team_id = ''
  fieldErrors.challenge_id = ''
}

function closeDialog() {
  emit('update:open', false)
}

function handleSubmit() {
  clearErrors()

  if (!form.attacker_team_id) {
    fieldErrors.attacker_team_id = '请选择攻击队伍'
  }
  if (!form.victim_team_id) {
    fieldErrors.victim_team_id = '请选择受害队伍'
  }
  if (!form.challenge_id) {
    fieldErrors.challenge_id = '请选择题目'
  }
  if (
    form.attacker_team_id &&
    form.victim_team_id &&
    form.attacker_team_id === form.victim_team_id
  ) {
    fieldErrors.victim_team_id = '攻击队伍和受害队伍不能相同'
  }

  if (fieldErrors.attacker_team_id || fieldErrors.victim_team_id || fieldErrors.challenge_id) {
    return
  }

  emit('save', {
    attacker_team_id: Number(form.attacker_team_id),
    victim_team_id: Number(form.victim_team_id),
    challenge_id: Number(form.challenge_id),
    attack_type: form.attack_type,
    submitted_flag: form.submitted_flag.trim() || undefined,
    is_success: form.is_success,
  })
}
</script>

<template>
  <AdminSurfaceModal
    :open="open"
    title="补录攻击日志"
    subtitle="将线下核实过的攻击事件补录进复盘记录，不直接改写正式排行榜。"
    eyebrow="AWD Operations"
    width="35rem"
    @close="closeDialog"
    @update:open="emit('update:open', $event)"
  >
    <form class="space-y-5" @submit.prevent="handleSubmit">
      <div class="grid gap-4 sm:grid-cols-2">
        <div class="ui-field awd-attack-field">
          <label class="ui-field__label" for="awd-attack-team">攻击队伍</label>
          <span class="ui-control-wrap" :class="{ 'is-error': !!fieldErrors.attacker_team_id }">
            <select id="awd-attack-team" v-model="form.attacker_team_id" class="ui-control">
              <option value="" disabled>请选择攻击队伍</option>
              <option v-for="team in teams" :key="team.id" :value="team.id">
                {{ team.name }}
              </option>
            </select>
          </span>
          <p v-if="fieldErrors.attacker_team_id" class="ui-field__error">
            {{ fieldErrors.attacker_team_id }}
          </p>
        </div>

        <div class="ui-field awd-attack-field">
          <label class="ui-field__label" for="awd-victim-team">受害队伍</label>
          <span class="ui-control-wrap" :class="{ 'is-error': !!fieldErrors.victim_team_id }">
            <select id="awd-victim-team" v-model="form.victim_team_id" class="ui-control">
              <option value="" disabled>请选择受害队伍</option>
              <option v-for="team in teams" :key="team.id" :value="team.id">
                {{ team.name }}
              </option>
            </select>
          </span>
          <p v-if="fieldErrors.victim_team_id" class="ui-field__error">
            {{ fieldErrors.victim_team_id }}
          </p>
        </div>
      </div>

      <div class="grid gap-4 sm:grid-cols-2">
        <div class="ui-field awd-attack-field">
          <label class="ui-field__label" for="awd-attack-challenge">题目</label>
          <span class="ui-control-wrap" :class="{ 'is-error': !!fieldErrors.challenge_id }">
            <select id="awd-attack-challenge" v-model="form.challenge_id" class="ui-control">
              <option value="" disabled>请选择题目</option>
              <option
                v-for="challenge in challengeOptions"
                :key="challenge.id"
                :value="challenge.challenge_id"
              >
                {{ getChallengeLabel(challenge) }}
              </option>
            </select>
          </span>
          <p v-if="fieldErrors.challenge_id" class="ui-field__error">
            {{ fieldErrors.challenge_id }}
          </p>
        </div>

        <div class="ui-field awd-attack-field">
          <label class="ui-field__label" for="awd-attack-type">攻击类型</label>
          <span class="ui-control-wrap">
            <select id="awd-attack-type" v-model="form.attack_type" class="ui-control">
              <option value="flag_capture">Flag 获取</option>
              <option value="service_exploit">服务利用</option>
            </select>
          </span>
        </div>
      </div>

      <div class="ui-field awd-attack-field">
        <label class="ui-field__label" for="awd-attack-flag">提交 Flag</label>
        <span class="ui-control-wrap">
          <input
            id="awd-attack-flag"
            v-model="form.submitted_flag"
            type="text"
            class="ui-control"
            placeholder="可选，补录 flag_capture 时填写"
          />
        </span>
      </div>

      <label class="ui-control-wrap awd-attack-check">
        <input v-model="form.is_success" type="checkbox" class="awd-attack-check__box" />
        <span class="awd-attack-check__label">本次攻击判定成功</span>
      </label>
      <p class="ui-field__hint">
        人工补录仅进入当前轮复盘记录，不写入正式排行榜与实时竞赛得分。
      </p>
      <p v-if="!hasTargets" class="ui-field__hint awd-attack-field__warning">
        至少需要 2 支队伍且已关联题目后，才能补录攻击日志。
      </p>
    </form>

    <template #footer>
      <div class="awd-attack-dialog__footer">
        <button
          type="button"
          class="ui-btn ui-btn--secondary"
          @click="closeDialog"
        >
          取消
        </button>
        <button
          id="awd-attack-log-submit"
          type="button"
          class="ui-btn ui-btn--primary"
          :disabled="saving || !hasTargets"
          @click="handleSubmit"
        >
          {{ saving ? '保存中...' : '保存攻击日志' }}
        </button>
      </div>
    </template>
  </AdminSurfaceModal>
</template>

<style scoped>
.awd-attack-field {
  --ui-field-gap: var(--space-2);
}

.awd-attack-field__warning {
  color: var(--color-warning);
}

.awd-attack-check {
  gap: var(--space-3);
  padding-inline: 1rem;
  color: var(--color-text-primary);
}

.awd-attack-check__box {
  width: 1rem;
  height: 1rem;
  flex-shrink: 0;
}

.awd-attack-check__label {
  font-size: var(--font-size-13);
  font-weight: 600;
}

.awd-attack-dialog__footer {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: var(--space-2);
}

@media (max-width: 767px) {
  .awd-attack-dialog__footer {
    flex-direction: column-reverse;
  }
}
</style>
