<script setup lang="ts">
import { computed, reactive, watch } from 'vue'

import AdminSurfaceModal from '@/components/common/modal-templates/AdminSurfaceModal.vue'
import type {
  AdminChallengeListItem,
  AdminContestChallengeData,
  ContestDetailData,
} from '@/api/contracts'

type DialogMode = 'create' | 'edit'

const props = defineProps<{
  open: boolean
  mode: DialogMode
  contestMode: ContestDetailData['mode']
  challengeOptions: AdminChallengeListItem[]
  existingChallengeIds: string[]
  draft?: AdminContestChallengeData | null
  loadingChallengeCatalog: boolean
  saving: boolean
}>()

const emit = defineEmits<{
  'update:open': [value: boolean]
  save: [
    value: {
      challenge_id: number
      points: number
      order: number
      is_visible: boolean
    },
  ]
}>()

const form = reactive({
  challenge_id: '',
  points: '100',
  order: '0',
  is_visible: 'true',
})

const fieldErrors = reactive({
  challenge_id: '',
  points: '',
  order: '',
})

const dialogTitle = computed(() => (props.mode === 'create' ? '关联赛事题目' : '编辑赛事题目'))

const selectableChallenges = computed(() =>
  props.challengeOptions.filter(
    (item) => props.mode === 'edit' || !props.existingChallengeIds.includes(item.id)
  )
)

const isAwdContest = computed(() => props.contestMode === 'awd')

watch(
  () => [props.open, props.mode, props.draft] as const,
  ([open]) => {
    if (!open) {
      return
    }
    form.challenge_id =
      props.mode === 'edit'
        ? props.draft?.challenge_id || ''
        : selectableChallenges.value[0]?.id || ''
    form.points = String(props.draft?.points ?? 100)
    form.order = String(props.draft?.order ?? 0)
    form.is_visible = props.draft?.is_visible === false ? 'false' : 'true'
    clearErrors()
  },
  { immediate: true }
)

function clearErrors() {
  fieldErrors.challenge_id = ''
  fieldErrors.points = ''
  fieldErrors.order = ''
}

function closeDialog() {
  emit('update:open', false)
}

function submit() {
  clearErrors()

  if (!form.challenge_id.trim()) {
    fieldErrors.challenge_id = '请选择题目'
  }

  const points = Number(form.points)
  if (!Number.isFinite(points) || points < 1) {
    fieldErrors.points = '分值至少为 1'
  }

  const order = Number(form.order)
  if (!Number.isFinite(order) || order < 0) {
    fieldErrors.order = '顺序不能小于 0'
  }

  if (fieldErrors.challenge_id || fieldErrors.points || fieldErrors.order) {
    return
  }

  emit('save', {
    challenge_id: Number(form.challenge_id),
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
        ? '先处理题目关联、顺序、分值和可见性，AWD 深度配置继续留在专用工作台。'
        : '维护赛事题目的关联关系、顺序、分值和可见性。'
    "
    eyebrow="Contest Orchestration"
    width="40rem"
    @close="closeDialog"
    @update:open="emit('update:open', $event)"
  >
    <form class="contest-challenge-dialog" @submit.prevent="submit">
      <p v-if="isAwdContest" class="contest-challenge-dialog__hint">
        当前弹层只处理题目关联、顺序、分值和可见性；AWD 深度配置在下一阶段完成。
      </p>

      <label class="ui-field contest-challenge-dialog__field" for="contest-challenge-select">
        <span class="ui-field__label contest-challenge-dialog__label">题目</span>
        <template v-if="mode === 'create'">
          <span class="ui-control-wrap" :class="{ 'is-disabled': loadingChallengeCatalog || selectableChallenges.length === 0, 'is-error': !!fieldErrors.challenge_id }">
            <select
              id="contest-challenge-select"
              v-model="form.challenge_id"
              class="ui-control contest-challenge-dialog__control"
              :disabled="loadingChallengeCatalog || selectableChallenges.length === 0"
            >
              <option value="" disabled>
                {{ loadingChallengeCatalog ? '正在加载题目目录...' : '请选择题目' }}
              </option>
              <option
                v-for="challenge in selectableChallenges"
                :key="challenge.id"
                :value="challenge.id"
              >
                {{ challenge.title }}
              </option>
            </select>
          </span>
        </template>
        <template v-else>
          <span class="ui-control-wrap contest-challenge-dialog__readonly">
            <span class="ui-control contest-challenge-dialog__control">
              {{ draft?.title || `Challenge #${draft?.challenge_id || ''}` }}
            </span>
          </span>
        </template>
        <span v-if="fieldErrors.challenge_id" class="ui-field__error contest-challenge-dialog__error">
          {{ fieldErrors.challenge_id }}
        </span>
      </label>

      <div class="contest-challenge-dialog__grid">
        <label class="ui-field contest-challenge-dialog__field" for="contest-challenge-points">
          <span class="ui-field__label contest-challenge-dialog__label">分值</span>
          <span class="ui-control-wrap" :class="{ 'is-error': !!fieldErrors.points }">
            <input
              id="contest-challenge-points"
              v-model="form.points"
              type="number"
              min="1"
              step="1"
              class="ui-control contest-challenge-dialog__control"
            />
          </span>
          <span v-if="fieldErrors.points" class="ui-field__error contest-challenge-dialog__error">
            {{ fieldErrors.points }}
          </span>
        </label>

        <label class="ui-field contest-challenge-dialog__field" for="contest-challenge-order">
          <span class="ui-field__label contest-challenge-dialog__label">顺序</span>
          <span class="ui-control-wrap" :class="{ 'is-error': !!fieldErrors.order }">
            <input
              id="contest-challenge-order"
              v-model="form.order"
              type="number"
              min="0"
              step="1"
              class="ui-control contest-challenge-dialog__control"
            />
          </span>
          <span v-if="fieldErrors.order" class="ui-field__error contest-challenge-dialog__error">
            {{ fieldErrors.order }}
          </span>
        </label>
      </div>

      <label class="ui-field contest-challenge-dialog__field" for="contest-challenge-visibility">
        <span class="ui-field__label contest-challenge-dialog__label">可见性</span>
        <span class="ui-control-wrap">
          <select
            id="contest-challenge-visibility"
            v-model="form.is_visible"
            class="ui-control contest-challenge-dialog__control"
          >
            <option value="true">可见</option>
            <option value="false">隐藏</option>
          </select>
        </span>
      </label>
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
          {{ saving ? '保存中...' : mode === 'create' ? '关联题目' : '保存变更' }}
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

.contest-challenge-dialog__grid {
  display: grid;
  gap: var(--space-4);
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.contest-challenge-dialog__field {
  --ui-field-gap: var(--space-2);
}

.contest-challenge-dialog__label {
  font-size: var(--font-size-0-875);
}

.contest-challenge-dialog__control,
.contest-challenge-dialog__readonly {
  min-height: 2.75rem;
}

.contest-challenge-dialog__readonly {
  background: color-mix(in srgb, var(--journal-surface) 96%, transparent);
}

.contest-challenge-dialog__hint {
  margin: 0;
  font-size: var(--font-size-0-875);
  color: var(--journal-muted);
}

.contest-challenge-dialog__error {
  font-size: var(--font-size-0-75);
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
  .contest-challenge-dialog__grid {
    grid-template-columns: minmax(0, 1fr);
  }

  .contest-challenge-dialog__footer {
    flex-direction: column-reverse;
  }
}
</style>
