<script setup lang="ts">
import AppEmpty from '@/components/common/AppEmpty.vue'
import type { ContestChallengeItem, SubmitFlagData } from '@/api/contracts'

interface Props {
  challenges: ContestChallengeItem[]
  selectedChallenge: ContestChallengeItem | null
  flagInput: string
  submitting: boolean
  submitResult: SubmitFlagData | null
}

const props = defineProps<Props>()

const emit = defineEmits<{
  'select-challenge': [challenge: ContestChallengeItem]
  'update:flagInput': [value: string]
  'submit-flag': []
}>()

function updateFlagInput(event: Event): void {
  emit('update:flagInput', (event.target as HTMLInputElement).value)
}

function challengeClass(challengeId: string, solved: boolean): string[] {
  const active = props.selectedChallenge?.id === challengeId
  return [
    'contest-challenge',
    active ? 'contest-challenge--active' : '',
    solved ? 'contest-challenge--solved' : '',
  ]
}

function selectedChallengeMeta(): string {
  if (!props.selectedChallenge) return ''
  return `${props.selectedChallenge.category} · ${props.selectedChallenge.points} 分`
}
</script>

<template>
  <div
    v-if="challenges.length === 0"
    class="contest-empty-state"
  >
    <AppEmpty
      icon="Flag"
      title="暂无题目"
      description="当前竞赛尚未发布题目。"
    />
  </div>

  <div
    v-else
    class="contest-challenge-workspace"
  >
    <div class="contest-challenge-list">
      <button
        v-for="challenge in challenges"
        :key="challenge.id"
        type="button"
        :class="challengeClass(challenge.id, challenge.is_solved)"
        @click="emit('select-challenge', challenge)"
      >
        <div class="contest-challenge__head">
          <h3 class="contest-challenge__title">
            {{ challenge.title }}
          </h3>
          <span
            v-if="challenge.is_solved"
            class="contest-challenge__solved"
          >✓</span>
        </div>
        <div class="contest-challenge__meta">
          <span>{{ challenge.category }}</span>
          <span>{{ challenge.points }} 分</span>
          <span>{{ challenge.solved_count }} 人解出</span>
        </div>
      </button>
    </div>

    <article class="challenge-focus">
      <template v-if="selectedChallenge">
        <div class="challenge-focus__head">
          <div>
            <div class="workspace-overline">
              已选题目
            </div>
            <h3 class="challenge-focus__title">
              {{ selectedChallenge.title }}
            </h3>
          </div>
          <div class="challenge-focus__meta">
            {{ selectedChallengeMeta() }}
          </div>
        </div>

        <div class="challenge-focus__stats">
          <span class="contest-chip contest-chip--neutral">
            解出人数 {{ selectedChallenge.solved_count }}
          </span>
          <span
            v-if="selectedChallenge.is_solved"
            class="contest-chip contest-chip--success"
          >
            已解出
          </span>
        </div>

        <div class="contest-divider contest-divider--compact" />

        <div class="challenge-focus__form">
          <div>
            <div class="workspace-overline">
              主要操作
            </div>
            <h4 class="challenge-focus__form-title">
              提交 Flag
            </h4>
          </div>

          <label
            class="ui-field__label flag-submit__label"
            for="contest-flag-input"
          >
            Flag 值
          </label>
          <div class="flag-submit">
            <div class="ui-control-wrap flag-submit__control">
              <input
                id="contest-flag-input"
                :value="flagInput"
                type="text"
                placeholder="flag{...}"
                class="ui-control"
                @input="updateFlagInput"
                @keyup.enter="emit('submit-flag')"
              >
            </div>
            <button
              type="button"
              :disabled="submitting"
              class="ui-btn ui-btn--primary"
              @click="emit('submit-flag')"
            >
              {{ submitting ? '提交中...' : '提交' }}
            </button>
          </div>

          <div
            v-if="submitResult"
            class="contest-alert"
            :class="submitResult.is_correct ? 'contest-alert--success' : 'contest-alert--danger'"
          >
            {{ submitResult.is_correct ? `正确！+${submitResult.points ?? 0} 分` : submitResult.message }}
          </div>
        </div>
      </template>

      <div
        v-else
        class="contest-inline-note"
      >
        从左侧选择题目后可在这里提交 Flag。
      </div>
    </article>
  </div>
</template>

<style scoped>
.contest-empty-state {
  margin-top: 1rem;
}

.contest-challenge-workspace {
  margin-top: 1rem;
  display: grid;
  gap: 1.25rem;
  grid-template-columns: minmax(0, 18rem) minmax(0, 1fr);
}

.contest-challenge-list {
  display: grid;
  gap: 0.45rem;
}

.contest-challenge {
  border: 0;
  border-inline-start: 2px solid
    color-mix(in srgb, var(--contest-accent) 24%, var(--color-border-default));
  border-bottom: 1px solid color-mix(in srgb, var(--color-border-default) 82%, transparent);
  background: transparent;
  padding: 0.75rem 0.35rem 0.75rem 0.85rem;
  text-align: left;
  transition:
    border-color 150ms ease,
    background 150ms ease,
    color 150ms ease;
}

.contest-challenge:hover,
.contest-challenge:focus-visible {
  border-inline-start-color: color-mix(in srgb, var(--contest-accent) 72%, var(--color-border-default));
  background: color-mix(in srgb, var(--contest-accent) 5%, transparent);
  outline: none;
}

.contest-challenge--active {
  border-inline-start-color: color-mix(in srgb, var(--contest-accent) 86%, var(--color-border-default));
  background: color-mix(in srgb, var(--contest-accent) 7%, transparent);
}

.contest-challenge--solved {
  border-inline-start-color: color-mix(in srgb, var(--color-success) 68%, var(--color-border-default));
}

.contest-challenge__head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 0.55rem;
}

.contest-challenge__title {
  font-size: var(--font-size-0-90);
  font-weight: 700;
  color: var(--color-text-primary);
}

.contest-challenge__solved {
  color: var(--color-success);
  font-weight: 700;
}

.contest-challenge__meta {
  margin-top: 0.35rem;
  display: flex;
  flex-wrap: wrap;
  gap: 0.4rem 0.7rem;
  font-size: var(--font-size-0-76);
  color: var(--color-text-secondary);
}

.challenge-focus {
  min-width: 0;
}

.challenge-focus__head {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-end;
  justify-content: space-between;
  gap: 0.75rem;
}

.challenge-focus__title {
  margin-top: 0.35rem;
  font-size: var(--font-size-1-20);
  font-weight: 700;
  color: var(--color-text-primary);
}

.challenge-focus__meta {
  font-size: var(--font-size-0-82);
  color: var(--color-text-secondary);
}

.challenge-focus__stats {
  margin-top: 1rem;
  display: flex;
  flex-wrap: wrap;
  gap: 0.55rem;
}

.challenge-focus__form-title {
  margin-top: 0.35rem;
  font-size: var(--font-size-1-00);
  font-weight: 700;
  color: var(--color-text-primary);
}

.flag-submit {
  margin-top: 0.55rem;
  display: flex;
  flex-wrap: wrap;
  gap: 0.6rem;
}

.flag-submit__label {
  display: inline-flex;
  margin-top: 1rem;
}

.flag-submit__control {
  flex: 1 1 18rem;
  min-width: 0;
}

.contest-alert {
  margin-top: 0.8rem;
  border-inline-start: 2px solid transparent;
  padding: 0.6rem 0.75rem;
  font-size: var(--font-size-0-84);
  line-height: 1.6;
}

.contest-alert--success {
  border-inline-start-color: color-mix(in srgb, var(--color-success) 60%, transparent);
  background: color-mix(in srgb, var(--color-success) 8%, transparent);
  color: color-mix(in srgb, var(--color-success) 86%, var(--color-text-primary));
}

.contest-alert--danger {
  border-inline-start-color: color-mix(in srgb, var(--color-danger) 60%, transparent);
  background: color-mix(in srgb, var(--color-danger) 8%, transparent);
  color: color-mix(in srgb, var(--color-danger) 86%, var(--color-text-primary));
}

.contest-inline-note {
  border-inline-start: 2px solid color-mix(in srgb, var(--color-border-default) 84%, transparent);
  padding-inline-start: 0.85rem;
  font-size: var(--font-size-0-88);
  line-height: 1.7;
  color: var(--color-text-secondary);
}

@media (max-width: 1100px) {
  .contest-challenge-workspace {
    grid-template-columns: minmax(0, 1fr);
  }
}

@media (max-width: 640px) {
  .flag-submit {
    flex-direction: column;
  }
}
</style>
