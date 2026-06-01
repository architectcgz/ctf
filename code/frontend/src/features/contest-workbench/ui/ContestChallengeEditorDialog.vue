<script setup lang="ts">
import { computed, reactive, watch } from 'vue'

import AdminSurfaceModal from '@/shared/ui/common/modal-templates/AdminSurfaceModal.vue'
import type {
  AdminAwdChallengeData,
  AdminChallengeListItem,
  AdminContestChallengeViewData,
  ContestDetailData,
} from '@/api/contracts'
import ContestAwdChallengeSelectorSection from './ContestAwdChallengeSelectorSection.vue'
import ContestChallengeSettingsSection from './ContestChallengeSettingsSection.vue'

type DialogMode = 'create' | 'edit'

const props = defineProps<{
  open: boolean
  mode: DialogMode
  contestMode: ContestDetailData['mode']
  challengeOptions: AdminChallengeListItem[]
  awdChallengeOptions?: AdminAwdChallengeData[]
  awdChallengePage?: number
  awdChallengePageSize?: number
  awdChallengeTotal?: number
  awdChallengeKeyword?: string
  awdChallengeServiceType?: AdminAwdChallengeData['service_type'] | ''
  awdChallengeDeploymentMode?: AdminAwdChallengeData['deployment_mode'] | ''
  awdChallengeReadiness?: AdminAwdChallengeData['readiness_status'] | ''
  awdChallengeLoadError?: string
  existingChallengeIds: string[]
  draft?: AdminContestChallengeViewData | null
  loadingChallengeCatalog: boolean
  loadingAwdChallengeCatalog?: boolean
  saving: boolean
}>()

const emit = defineEmits<{
  'update:open': [value: boolean]
  save: [
    value: {
      challenge_id?: number
      awd_challenge_id?: number
      awd_challenge_ids?: number[]
      points: number
      order: number
      is_visible: boolean
    },
  ]
  'update-awd-challenge-keyword': [value: string]
  'update-awd-challenge-service-type': [value: AdminAwdChallengeData['service_type'] | '']
  'update-awd-challenge-deployment-mode': [value: AdminAwdChallengeData['deployment_mode'] | '']
  'update-awd-challenge-readiness': [value: AdminAwdChallengeData['readiness_status'] | '']
  'change-awd-challenge-page': [page: number]
  'refresh-awd-challenge-catalog': []
}>()

const form = reactive({
  challenge_id: '',
  awd_challenge_id: '',
  awd_challenge_ids: [] as string[],
  points: '100',
  order: '0',
  is_visible: 'true',
})

const fieldErrors = reactive({
  challenge_id: '',
  awd_challenge_id: '',
  points: '',
  order: '',
})

const dialogTitle = computed(() =>
  props.mode === 'create'
    ? isAwdContest.value
      ? '关联 AWD 题目'
      : '关联赛事题目'
    : '编辑赛事题目'
)

const selectableChallenges = computed(() =>
  props.challengeOptions.filter(
    (item) => props.mode === 'edit' || !props.existingChallengeIds.includes(item.id)
  )
)
const selectableAwdChallenges = computed(() =>
  (props.awdChallengeOptions ?? []).filter(
    (item) => props.mode === 'edit' || !props.existingChallengeIds.includes(item.id)
  )
)

const isAwdContest = computed(() => props.contestMode === 'awd')
const isAwdCreateMode = computed(() => isAwdContest.value && props.mode === 'create')
const dialogWidth = computed(() =>
  isAwdCreateMode.value ? 'min(60rem, calc(100vw - (var(--space-4) * 2)))' : '42rem'
)
const showContestSelector = computed(() => !isAwdContest.value || props.mode === 'edit')
const showContestSettings = computed(() => !isAwdCreateMode.value)
const showAwdChallengeSelector = computed(() => isAwdCreateMode.value)
const awdChallengePage = computed(() => props.awdChallengePage ?? 1)
const awdChallengePageSize = computed(() => props.awdChallengePageSize ?? 20)
const awdChallengeTotal = computed(() => props.awdChallengeTotal ?? selectableAwdChallenges.value.length)
const awdChallengeTotalPages = computed(() =>
  Math.max(1, Math.ceil(awdChallengeTotal.value / awdChallengePageSize.value))
)
const selectedAwdChallengeIds = computed(() =>
  isAwdCreateMode.value
    ? form.awd_challenge_ids
    : form.awd_challenge_id
      ? [form.awd_challenge_id]
      : []
)

watch(
  () => [props.open, props.mode, props.draft, selectableChallenges.value, selectableAwdChallenges.value] as const,
  ([open]) => {
    if (!open) {
      return
    }
    form.challenge_id =
      props.mode === 'edit'
        ? props.draft?.challenge_id || ''
        : isAwdContest.value
          ? ''
          : ''
    form.awd_challenge_id = isAwdContest.value
      ? props.draft?.awd_challenge_id || ''
      : ''
    form.awd_challenge_ids = isAwdCreateMode.value && form.awd_challenge_id ? [form.awd_challenge_id] : []
    form.points = String(props.draft?.points ?? 100)
    form.order = String(props.draft?.order ?? 0)
    form.is_visible = props.draft?.is_visible === false ? 'false' : 'true'
    clearErrors()
  },
  { immediate: true, deep: true }
)

function clearErrors() {
  fieldErrors.challenge_id = ''
  fieldErrors.awd_challenge_id = ''
  fieldErrors.points = ''
  fieldErrors.order = ''
}

function closeDialog() {
  emit('update:open', false)
}

function selectAwdChallenge(awdChallengeId: string) {
  if (isAwdCreateMode.value) {
    const selected = new Set(form.awd_challenge_ids)
    if (selected.has(awdChallengeId)) {
      if (selected.size > 1) selected.delete(awdChallengeId)
    } else {
      selected.add(awdChallengeId)
    }
    form.awd_challenge_ids = selectableAwdChallenges.value
      .map((item) => item.id)
      .filter((id) => selected.has(id))
    form.awd_challenge_id = form.awd_challenge_ids[0] || ''
    return
  }

  form.awd_challenge_id = awdChallengeId
}

function submit() {
  if (props.saving) {
    return
  }

  clearErrors()

  if (!isAwdContest.value && !form.challenge_id.trim()) {
    fieldErrors.challenge_id = '请选择题目'
  }
  if (
    isAwdContest.value &&
    (isAwdCreateMode.value ? form.awd_challenge_ids.length === 0 : !form.awd_challenge_id.trim())
  ) {
    fieldErrors.awd_challenge_id = '请选择 AWD 题目'
  }

  const points = Number(form.points)
  if (!Number.isFinite(points) || points < 1) {
    fieldErrors.points = '分值至少为 1'
  }

  const order = Number(form.order)
  if (!Number.isFinite(order) || order < 0) {
    fieldErrors.order = '顺序不能小于 0'
  }

  if (
    fieldErrors.challenge_id ||
    fieldErrors.awd_challenge_id ||
    fieldErrors.points ||
    fieldErrors.order
  ) {
    return
  }

  emit('save', {
    challenge_id: isAwdContest.value
      ? undefined
      : form.challenge_id
        ? Number(form.challenge_id)
        : undefined,
    awd_challenge_id: isAwdContest.value ? Number(form.awd_challenge_id) : undefined,
    awd_challenge_ids: isAwdCreateMode.value
      ? form.awd_challenge_ids.map((id) => Number(id))
      : undefined,
    points,
    order,
    is_visible: form.is_visible === 'true',
  })
}
</script>

<template>
  <AdminSurfaceModal
    :open="open"
    :title="dialogTitle"
    :subtitle="
      isAwdContest
          ? '从 AWD 题库选择题目'
        : '维护赛事题目的关联关系、顺序、分值和可见性。'
    "
    eyebrow="Contest Orchestration"
    :width="dialogWidth"
    @close="closeDialog"
    @update:open="emit('update:open', $event)"
  >
    <form
      class="contest-challenge-dialog"
      @submit.prevent="submit"
    >
      <ContestChallengeSettingsSection
        :mode="mode"
        :contest-mode="contestMode"
        :show-contest-selector="showContestSelector"
        :show-contest-settings="showContestSettings"
        :challenge-options="selectableChallenges"
        :draft="draft"
        :loading-challenge-catalog="loadingChallengeCatalog"
        :challenge-id="form.challenge_id"
        :points="form.points"
        :order="form.order"
        :is-visible="form.is_visible"
        :challenge-error="fieldErrors.challenge_id"
        :points-error="fieldErrors.points"
        :order-error="fieldErrors.order"
        @update:challenge-id="form.challenge_id = $event"
        @update:points="form.points = $event"
        @update:order="form.order = $event"
        @update:is-visible="form.is_visible = $event"
      />

      <ContestAwdChallengeSelectorSection
        v-if="showAwdChallengeSelector"
        :awd-challenge-options="selectableAwdChallenges"
        :awd-challenge-page="awdChallengePage"
        :awd-challenge-total-pages="awdChallengeTotalPages"
        :awd-challenge-keyword="awdChallengeKeyword"
        :awd-challenge-service-type="awdChallengeServiceType"
        :awd-challenge-deployment-mode="awdChallengeDeploymentMode"
        :awd-challenge-readiness="awdChallengeReadiness"
        :awd-challenge-load-error="props.awdChallengeLoadError"
        :loading-awd-challenge-catalog="loadingAwdChallengeCatalog ?? false"
        :selected-awd-challenge-ids="selectedAwdChallengeIds"
        :field-error="fieldErrors.awd_challenge_id"
        @select="selectAwdChallenge"
        @update-awd-challenge-keyword="emit('update-awd-challenge-keyword', $event)"
        @update-awd-challenge-service-type="emit('update-awd-challenge-service-type', $event)"
        @update-awd-challenge-deployment-mode="emit('update-awd-challenge-deployment-mode', $event)"
        @update-awd-challenge-readiness="emit('update-awd-challenge-readiness', $event)"
        @change-awd-challenge-page="emit('change-awd-challenge-page', $event)"
        @refresh-awd-challenge-catalog="emit('refresh-awd-challenge-catalog')"
      />
    </form>

    <template #footer>
      <div class="contest-challenge-dialog__footer">
        <button
          type="button"
          class="ui-btn ui-btn--secondary contest-challenge-dialog__button"
          @click="closeDialog"
        >
          取消
        </button>
        <button
          id="contest-challenge-dialog-submit"
          type="button"
          class="ui-btn ui-btn--primary contest-challenge-dialog__button"
          :disabled="saving"
          @click="submit"
        >
          {{ saving ? '保存中...' : mode === 'create' ? (isAwdContest ? '关联题目' : '关联题目') : '保存变更' }}
        </button>
      </div>
    </template>
  </AdminSurfaceModal>
</template>

<style scoped>
.contest-challenge-dialog {
  display: grid;
  gap: var(--space-4);
}

.contest-challenge-dialog__footer {
  display: flex;
  justify-content: flex-end;
  gap: var(--space-3);
}

.contest-challenge-dialog__button {
  min-width: 6rem;
}

@media (max-width: 767px) {
  .contest-challenge-dialog__footer {
    flex-direction: column-reverse;
  }
}
</style>
