<script setup lang="ts">
import { CheckCircle2, Target, Trophy, UsersRound } from 'lucide-vue-next'

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

function difficultyLabel(difficulty: ContestChallengeItem['difficulty']): string {
  const labels: Record<ContestChallengeItem['difficulty'], string> = {
    beginner: '入门',
    easy: '简单',
    medium: '中等',
    hard: '困难',
    insane: '极难',
  }

  return labels[difficulty]
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
    <div
      class="contest-challenge-list"
      aria-label="竞赛题目列表"
    >
      <button
        v-for="challenge in challenges"
        :key="challenge.id"
        type="button"
        :class="challengeClass(challenge.id, challenge.is_solved)"
        :aria-pressed="selectedChallenge?.id === challenge.id"
        :title="challenge.title"
        @click="emit('select-challenge', challenge)"
      >
        <div class="contest-challenge__head">
          <h3 class="contest-challenge__title">
            {{ challenge.title }}
          </h3>
          <span
            v-if="challenge.is_solved"
            class="contest-challenge__solved"
            aria-label="已解出"
          >
            <CheckCircle2
              class="contest-challenge__solved-icon"
              aria-hidden="true"
            />
          </span>
        </div>
        <div class="contest-challenge__meta">
          <span class="contest-challenge__meta-item">{{ challenge.category }}</span>
          <span class="contest-challenge__meta-item">{{ challenge.points }} 分</span>
          <span class="contest-challenge__solve-count">{{ challenge.solved_count }} 人解出</span>
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

        <div class="challenge-focus__summary">
          <div class="challenge-focus__summary-item">
            <Target
              class="challenge-focus__summary-icon"
              aria-hidden="true"
            />
            <span>难度</span>
            <strong>{{ difficultyLabel(selectedChallenge.difficulty) }}</strong>
          </div>
          <div class="challenge-focus__summary-item">
            <Trophy
              class="challenge-focus__summary-icon"
              aria-hidden="true"
            />
            <span>积分</span>
            <strong>{{ selectedChallenge.points }}</strong>
          </div>
          <div class="challenge-focus__summary-item">
            <UsersRound
              class="challenge-focus__summary-icon"
              aria-hidden="true"
            />
            <span>解出</span>
            <strong>{{ selectedChallenge.solved_count }}</strong>
          </div>
          <span
            v-if="selectedChallenge.is_solved"
            class="contest-chip contest-chip--success challenge-focus__status"
          >
            <CheckCircle2
              class="challenge-focus__status-icon"
              aria-hidden="true"
            />
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
              class="ui-btn ui-btn--primary flag-submit__button"
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
  margin-top: var(--space-4);
}

.contest-challenge-workspace {
  --contest-challenge-list-track: clamp(17rem, 28vw, 22rem);
  margin-top: var(--space-4);
  display: grid;
  align-items: start;
  gap: var(--space-6);
  grid-template-columns: minmax(0, var(--contest-challenge-list-track)) minmax(0, 1fr);
}

.contest-challenge-list {
  display: grid;
  gap: var(--space-2);
}

.contest-challenge {
  min-width: 0;
  border: 0;
  border-inline-start: 2px solid
    color-mix(in srgb, var(--contest-accent) 24%, var(--color-border-default));
  border-bottom: 1px solid color-mix(in srgb, var(--color-border-default) 82%, transparent);
  background: transparent;
  padding: var(--space-3) var(--space-2) var(--space-3) var(--space-3);
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
  outline: 2px solid color-mix(in srgb, var(--contest-accent) 26%, transparent);
  outline-offset: var(--space-1);
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
  gap: var(--space-2);
}

.contest-challenge__title {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: var(--font-size-0-90);
  font-weight: 700;
  color: var(--color-text-primary);
}

.contest-challenge__solved {
  display: inline-flex;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  color: var(--color-success);
}

.contest-challenge__solved-icon {
  width: var(--space-4);
  height: var(--space-4);
}

.contest-challenge__meta {
  margin-top: var(--space-2);
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--space-1-5) var(--space-2);
  font-size: var(--font-size-0-76);
  color: var(--color-text-secondary);
}

.contest-challenge__meta-item,
.contest-challenge__solve-count {
  min-width: 0;
}

.contest-challenge__meta-item {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.contest-challenge__solve-count {
  border-radius: 999px;
  background: color-mix(in srgb, var(--color-bg-elevated) 74%, transparent);
  padding: var(--space-0-5) var(--space-2);
  color: color-mix(in srgb, var(--color-text-secondary) 90%, var(--color-text-primary));
}

.challenge-focus {
  min-width: 0;
  border-inline-start: 1px solid color-mix(in srgb, var(--color-border-default) 82%, transparent);
  padding-inline-start: var(--space-6);
}

.challenge-focus__head {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-end;
  justify-content: space-between;
  gap: var(--space-3);
}

.challenge-focus__title {
  margin-top: var(--space-1-5);
  overflow-wrap: anywhere;
  font-size: var(--font-size-1-20);
  font-weight: 700;
  color: var(--color-text-primary);
}

.challenge-focus__meta {
  font-size: var(--font-size-0-82);
  color: var(--color-text-secondary);
}

.challenge-focus__summary {
  margin-top: var(--space-4);
  display: grid;
  gap: var(--space-2);
  grid-template-columns: repeat(3, minmax(0, 1fr)) auto;
  align-items: stretch;
}

.challenge-focus__summary-item {
  display: grid;
  min-width: 0;
  grid-template-columns: auto minmax(0, 1fr);
  gap: var(--space-1) var(--space-2);
  border-bottom: 1px solid color-mix(in srgb, var(--color-border-default) 80%, transparent);
  padding: var(--space-2) 0 var(--space-2-5);
  color: var(--color-text-secondary);
  font-size: var(--font-size-0-78);
}

.challenge-focus__summary-icon {
  width: var(--space-4);
  height: var(--space-4);
  color: color-mix(in srgb, var(--contest-accent) 72%, var(--color-text-secondary));
}

.challenge-focus__summary-item strong {
  grid-column: 2;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  color: var(--color-text-primary);
  font-size: var(--font-size-0-92);
}

.challenge-focus__status {
  align-self: center;
  gap: var(--space-1-5);
  white-space: nowrap;
}

.challenge-focus__status-icon {
  width: var(--space-4);
  height: var(--space-4);
}

.challenge-focus__form-title {
  margin-top: var(--space-1-5);
  font-size: var(--font-size-1-00);
  font-weight: 700;
  color: var(--color-text-primary);
}

.flag-submit {
  margin-top: var(--space-2);
  display: flex;
  flex-wrap: wrap;
  align-items: stretch;
  gap: var(--space-2);
  max-width: 44rem;
}

.flag-submit__label {
  display: inline-flex;
  margin-top: var(--space-4);
}

.flag-submit__control {
  flex: 1 1 18rem;
  min-width: 0;
}

.flag-submit__button {
  min-width: 6rem;
}

.contest-alert {
  margin-top: var(--space-3);
  border-inline-start: 2px solid transparent;
  padding: var(--space-2-5) var(--space-3);
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
  padding-inline-start: var(--space-3);
  font-size: var(--font-size-0-88);
  line-height: 1.7;
  color: var(--color-text-secondary);
}

@media (max-width: 1100px) {
  .contest-challenge-workspace {
    gap: var(--space-5);
    grid-template-columns: minmax(0, 1fr);
  }

  .contest-challenge-list {
    display: flex;
    gap: var(--space-2);
    margin-inline: calc(var(--space-1) * -1);
    overflow-x: auto;
    padding: var(--space-1) var(--space-1) var(--space-2);
    scroll-snap-type: x proximity;
    scrollbar-width: thin;
  }

  .contest-challenge {
    flex: 0 0 min(19rem, 82vw);
    scroll-snap-align: start;
  }

  .challenge-focus {
    border-inline-start: 0;
    border-top: 1px solid color-mix(in srgb, var(--color-border-default) 82%, transparent);
    padding-block-start: var(--space-5);
    padding-inline-start: 0;
  }
}

@media (max-width: 640px) {
  .challenge-focus__summary {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .challenge-focus__status {
    justify-self: start;
  }

  .flag-submit {
    flex-direction: column;
    max-width: none;
  }

  .flag-submit__button {
    width: 100%;
  }
}
</style>
