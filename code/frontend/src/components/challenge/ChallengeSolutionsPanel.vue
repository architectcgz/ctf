<template>
  <section
    id="challenge-workspace-panel-solution"
    class="workspace-panel panel"
    role="tabpanel"
    aria-labelledby="challenge-workspace-tab-solution"
  >
    <section class="section section--flat">
      <div class="section-head workspace-tab-heading">
        <div class="workspace-tab-heading__main">
          <div class="workspace-overline">
            Solutions
          </div>
          <h2 class="section-title workspace-tab-heading__title">
            题解区
          </h2>
        </div>
        <div class="section-hint">
          推荐 {{ recommendedSolutionCount }} · 社区 {{ communitySolutionCount }}
        </div>
      </div>

      <div class="space-y-5">
        <div
          v-if="!challengeSolved"
          class="inline-note inline-note--warning"
        >
          解出题目后可查看推荐题解与社区题解。
        </div>

        <template v-else>
          <div class="solution-layout">
            <div class="solution-nav">
              <div
                class="solution-tabbar top-tabs challenge-subtabs"
                role="tablist"
                aria-label="题解分类"
              >
                <button
                  id="challenge-solutions-tab-recommended"
                  :ref="
                    (element) =>
                      setSolutionTabButtonRef(
                        'recommended',
                        element as HTMLButtonElement | null
                      )
                  "
                  type="button"
                  role="tab"
                  class="solution-tab top-tab challenge-subtab"
                  :class="{ active: activeSolutionTab === 'recommended' }"
                  :aria-selected="activeSolutionTab === 'recommended'"
                  aria-controls="challenge-solutions-panel-recommended"
                  :tabindex="activeSolutionTab === 'recommended' ? 0 : -1"
                  @click="emit('select-tab', 'recommended')"
                  @keydown="handleSolutionTabKeydown($event, 0)"
                >
                  推荐题解
                </button>
                <button
                  id="challenge-solutions-tab-community"
                  :ref="
                    (element) =>
                      setSolutionTabButtonRef(
                        'community',
                        element as HTMLButtonElement | null
                      )
                  "
                  type="button"
                  role="tab"
                  class="solution-tab top-tab challenge-subtab"
                  :class="{ active: activeSolutionTab === 'community' }"
                  :aria-selected="activeSolutionTab === 'community'"
                  aria-controls="challenge-solutions-panel-community"
                  :tabindex="activeSolutionTab === 'community' ? 0 : -1"
                  @click="emit('select-tab', 'community')"
                  @keydown="handleSolutionTabKeydown($event, 1)"
                >
                  社区题解
                </button>
              </div>

              <div
                v-if="displayedSolutionCards.length === 0"
                class="inline-note"
              >
                {{
                  activeSolutionTab === 'recommended'
                    ? '还没有推荐题解。'
                    : '还没有公开的社区题解。'
                }}
              </div>

              <button
                v-for="item in displayedSolutionCards"
                :key="item.id"
                type="button"
                class="solution-list-item solution-item"
                :class="{
                  'solution-list-item--active active': item.id === activeSolution?.id,
                }"
                @click="emit('select-solution', item.id)"
              >
                <strong>{{ item.title }}</strong>
                <span>{{ item.authorName }} · {{ formatWriteupTime(item.updatedAt) }}</span>
              </button>
            </div>

            <article
              :id="`challenge-solutions-panel-${activeSolutionTab}`"
              class="solution-preview"
              role="tabpanel"
              :aria-labelledby="`challenge-solutions-tab-${activeSolutionTab}`"
            >
              <template v-if="activeSolution">
                <div class="flex flex-wrap items-start justify-between gap-3">
                  <div>
                    <h3 class="text-lg font-semibold text-[var(--journal-ink)]">
                      {{ activeSolution.title }}
                    </h3>
                    <div class="mt-2 text-sm text-[var(--journal-muted)]">
                      {{ activeSolution.authorName }} · {{ activeSolution.sourceLabel }}
                    </div>
                  </div>
                  <div class="flex flex-wrap gap-2">
                    <span
                      v-if="activeSolution.badge"
                      class="writeup-status-pill"
                      :class="activeSolution.badgeClass"
                    >
                      {{ activeSolution.badge }}
                    </span>
                    <span class="writeup-status-pill writeup-status-pill--muted">
                      {{ formatWriteupTime(activeSolution.updatedAt) }}
                    </span>
                  </div>
                </div>
                <!-- eslint-disable-next-line vue/no-v-html -->
                <div
                  class="prose challenge-prose solution-preview__content mt-6 max-w-none"
                  v-html="sanitizedActiveSolutionContent"
                />
              </template>

              <div
                v-else
                class="inline-note"
              >
                当前分组还没有可展示的题解。
              </div>
            </article>
          </div>
        </template>
      </div>
    </section>
  </section>
</template>

<script setup lang="ts">
import type {
  ChallengeSolutionCard,
  ChallengeSolutionTab,
} from '@/composables/useChallengeDetailPresentation'

interface Props {
  challengeSolved: boolean
  recommendedSolutionCount: number
  communitySolutionCount: number
  activeSolutionTab: ChallengeSolutionTab
  displayedSolutionCards: ChallengeSolutionCard[]
  activeSolution: ChallengeSolutionCard | null
  sanitizedActiveSolutionContent: string
  formatWriteupTime: (value?: string) => string
  setSolutionTabButtonRef: (
    tab: ChallengeSolutionTab,
    element: HTMLButtonElement | null,
  ) => void
  handleSolutionTabKeydown: (event: KeyboardEvent, index: number) => void
}

defineProps<Props>()

const emit = defineEmits<{
  'select-tab': [tab: ChallengeSolutionTab]
  'select-solution': [solutionId: string]
}>()
</script>

<style scoped>
.section--flat {
  padding-top: 0;
  border-top: 0;
}

.section-head {
  display: flex;
  align-items: end;
  justify-content: space-between;
  gap: var(--space-4);
  margin-bottom: var(--space-4);
}

.section-hint {
  font-size: var(--font-size-13);
  line-height: 1.75;
  color: var(--text-faint);
}

.challenge-subtabs {
  --page-top-tabs-gap: var(--space-4-5);
  --page-top-tabs-margin: 0;
  --page-top-tabs-padding: 0 0 var(--space-2-5);
  --page-top-tabs-border: var(--line-soft);
  --page-top-tab-min-height: 3rem;
  --page-top-tab-padding: 0 0 var(--space-2);
  --page-top-tab-font-size: var(--font-size-14);
  --page-top-tab-font-weight: 600;
  --page-top-tab-color: var(--text-faint);
  --page-top-tab-active-color: var(--journal-accent-strong);
  --page-top-tab-active-border: var(--journal-accent);
  scrollbar-width: none;
}

.challenge-subtab,
.solution-tab {
  min-width: fit-content;
}

.solution-layout {
  display: grid;
  grid-template-columns: minmax(240px, 0.54fr) minmax(0, 1fr);
  gap: var(--space-6);
}

.solution-nav {
  padding-right: var(--space-5);
  border-right: 1px solid var(--line-soft);
}

.solution-item,
.solution-list-item {
  width: 100%;
  text-align: left;
  padding: var(--space-3-5) 0 var(--space-4) var(--space-3-5);
  border: 0;
  border-left: 2px solid transparent;
  border-bottom: 1px solid var(--line-soft);
  background: transparent;
}

.solution-item strong,
.solution-list-item strong {
  display: block;
  font-size: var(--font-size-14);
  color: var(--text-main);
}

.solution-item span,
.solution-list-item span {
  display: block;
  margin-top: var(--space-1-5);
  font-size: var(--font-size-12);
  color: var(--text-faint);
}

.solution-item.active,
.solution-list-item--active,
.solution-list-item:hover {
  border-left-color: color-mix(in srgb, var(--journal-accent) 40%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 4%, transparent);
}

.solution-list-item:focus-visible {
  outline: 2px solid color-mix(in srgb, var(--brand) 44%, var(--color-bg-base));
  outline-offset: 3px;
}

.solution-preview {
  min-height: 22rem;
  font-size: var(--font-size-14);
  line-height: 1.9;
  color: var(--text-subtle);
}

.solution-preview__content {
  min-height: 15rem;
}

.challenge-prose :deep(p),
.challenge-prose :deep(ul),
.challenge-prose :deep(ol) {
  margin-bottom: var(--space-4);
}

.challenge-prose :deep(pre) {
  overflow: auto;
  margin: var(--space-5) 0;
  padding: var(--space-4-5) var(--space-5);
  border: 1px solid var(--line-soft);
  border-radius: 14px;
  background: color-mix(in srgb, var(--bg-panel) 72%, var(--color-bg-base));
  color: var(--text-main);
  font: 13px/1.7 var(--font-mono);
}

.challenge-prose :deep(h1),
.challenge-prose :deep(h2),
.challenge-prose :deep(h3),
.challenge-prose :deep(strong),
.challenge-prose :deep(code) {
  color: var(--journal-ink);
}

.solution-preview__content :deep(h1),
.solution-preview__content :deep(h2),
.solution-preview__content :deep(h3) {
  margin-top: var(--space-5);
}

.inline-note {
  padding-left: var(--space-4);
  border-left: 2px solid var(--line-soft);
  font-size: var(--font-size-0-90);
  line-height: 1.8;
  color: var(--text-subtle);
}

.inline-note--warning {
  border-left-color: color-mix(in srgb, var(--color-warning) 34%, transparent);
  color: var(--journal-warning-ink);
}

.writeup-status-pill {
  display: inline-flex;
  align-items: center;
  min-height: 34px;
  padding: 0 var(--space-3-5);
  border: 1px solid var(--line-soft);
  border-radius: 999px;
  background: color-mix(in srgb, var(--bg-panel) 72%, transparent);
  font-size: var(--font-size-13);
  font-weight: 600;
  color: var(--text-subtle);
}

.writeup-status-pill--primary {
  border-color: color-mix(in srgb, var(--journal-accent) 18%, transparent);
  background: var(--journal-accent-soft);
  color: var(--journal-accent-strong);
}

.writeup-status-pill--muted {
  border-color: var(--line-soft);
  background: var(--bg-muted);
  color: var(--text-subtle);
}

@media (max-width: 760px) {
  .solution-layout {
    grid-template-columns: minmax(0, 1fr);
  }

  .solution-nav {
    padding-right: 0;
    border-right: 0;
  }
}
</style>
