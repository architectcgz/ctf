<script setup lang="ts">
import { computed, reactive, watch } from 'vue'

import type { AdminChallengeListItem, AdminContestChallengeData, ContestDetailData } from '@/api/contracts'

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

const dialogTitle = computed(() =>
  props.mode === 'create' ? '关联赛事题目' : '编辑赛事题目'
)

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
  <ElDialog
    :model-value="open"
    :title="dialogTitle"
    width="640px"
    @update:modelValue="emit('update:open', $event)"
  >
    <form class="contest-challenge-dialog" @submit.prevent="submit">
      <p v-if="isAwdContest" class="contest-challenge-dialog__hint">
        当前弹层只处理题目关联、顺序、分值和可见性；AWD 深度配置在下一阶段完成。
      </p>

      <label class="contest-challenge-dialog__field" for="contest-challenge-select">
        <span class="contest-challenge-dialog__label">题目</span>
        <template v-if="mode === 'create'">
          <select
            id="contest-challenge-select"
            v-model="form.challenge_id"
            class="contest-challenge-dialog__control"
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
        </template>
        <template v-else>
          <div class="contest-challenge-dialog__readonly">
            {{ draft?.title || `Challenge #${draft?.challenge_id || ''}` }}
          </div>
        </template>
        <span v-if="fieldErrors.challenge_id" class="contest-challenge-dialog__error">
          {{ fieldErrors.challenge_id }}
        </span>
      </label>

      <div class="contest-challenge-dialog__grid">
        <label class="contest-challenge-dialog__field" for="contest-challenge-points">
          <span class="contest-challenge-dialog__label">分值</span>
          <input
            id="contest-challenge-points"
            v-model="form.points"
            type="number"
            min="1"
            step="1"
            class="contest-challenge-dialog__control"
          />
          <span v-if="fieldErrors.points" class="contest-challenge-dialog__error">
            {{ fieldErrors.points }}
          </span>
        </label>

        <label class="contest-challenge-dialog__field" for="contest-challenge-order">
          <span class="contest-challenge-dialog__label">顺序</span>
          <input
            id="contest-challenge-order"
            v-model="form.order"
            type="number"
            min="0"
            step="1"
            class="contest-challenge-dialog__control"
          />
          <span v-if="fieldErrors.order" class="contest-challenge-dialog__error">
            {{ fieldErrors.order }}
          </span>
        </label>
      </div>

      <label class="contest-challenge-dialog__field" for="contest-challenge-visibility">
        <span class="contest-challenge-dialog__label">可见性</span>
        <select
          id="contest-challenge-visibility"
          v-model="form.is_visible"
          class="contest-challenge-dialog__control"
        >
          <option value="true">可见</option>
          <option value="false">隐藏</option>
        </select>
      </label>
    </form>

    <template #footer>
      <div class="contest-challenge-dialog__footer">
        <button
          type="button"
          class="contest-challenge-dialog__button contest-challenge-dialog__button--ghost"
          @click="closeDialog"
        >
          取消
        </button>
        <button
          id="contest-challenge-dialog-submit"
          type="button"
          class="contest-challenge-dialog__button contest-challenge-dialog__button--primary"
          :disabled="saving"
          @click="submit"
        >
          {{ saving ? '保存中...' : mode === 'create' ? '关联题目' : '保存变更' }}
        </button>
      </div>
    </template>
  </ElDialog>
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
  display: grid;
  gap: var(--space-2);
}

.contest-challenge-dialog__label {
  font-size: var(--font-size-0-875);
  font-weight: 600;
  color: var(--journal-ink);
}

.contest-challenge-dialog__control,
.contest-challenge-dialog__readonly {
  min-height: 2.75rem;
  border: 1px solid color-mix(in srgb, var(--journal-border) 76%, transparent);
  border-radius: 0.9rem;
  background: color-mix(in srgb, var(--journal-surface) 96%, transparent);
  padding: 0.75rem 0.95rem;
  color: var(--journal-ink);
}

.contest-challenge-dialog__readonly {
  display: flex;
  align-items: center;
}

.contest-challenge-dialog__hint {
  margin: 0;
  font-size: var(--font-size-0-875);
  color: var(--journal-muted);
}

.contest-challenge-dialog__error {
  font-size: var(--font-size-0-75);
  color: var(--color-danger);
}

.contest-challenge-dialog__footer {
  display: flex;
  justify-content: flex-end;
  gap: var(--space-3);
}

.contest-challenge-dialog__button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 2.5rem;
  border-radius: 0.8rem;
  padding: 0.65rem 1rem;
  font-size: var(--font-size-0-875);
  font-weight: 600;
  transition: all 150ms ease;
}

.contest-challenge-dialog__button--ghost {
  border: 1px solid color-mix(in srgb, var(--journal-border) 76%, transparent);
  background: color-mix(in srgb, var(--journal-surface) 94%, transparent);
  color: var(--journal-ink);
}

.contest-challenge-dialog__button--primary {
  border: 1px solid transparent;
  background: var(--color-primary);
  color: #fff;
}

.contest-challenge-dialog__button:disabled {
  cursor: not-allowed;
  opacity: 0.6;
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
