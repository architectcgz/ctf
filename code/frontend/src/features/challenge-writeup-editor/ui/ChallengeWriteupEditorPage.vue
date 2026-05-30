<script setup lang="ts">
import { RefreshCw } from 'lucide-vue-next'
import { computed } from 'vue'

import AppLoading from '@/components/common/AppLoading.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import { useChallengeWriteupEditorPage } from '../model'

import ChallengeWriteupChallengeRail from './ChallengeWriteupChallengeRail.vue'
import ChallengeWriteupEditorFormSection from './ChallengeWriteupEditorFormSection.vue'
import ChallengeWriteupSnapshotSection from './ChallengeWriteupSnapshotSection.vue'
import './challengeWriteupEditorPage.css'

const props = withDefaults(
  defineProps<{
    challengeId: string
    embedded?: boolean
  }>(),
  {
    embedded: false,
  }
)

const emit = defineEmits<{
  back: []
}>()

const isEmbedded = computed(() => props.embedded)
const pageShellClass = computed(() =>
  isEmbedded.value
    ? 'writeup-embedded-shell'
    : 'writeup-editor-page-shell workspace-shell journal-shell journal-shell-admin journal-notes-card journal-hero flex min-h-full flex-1 flex-col'
)
const {
  loading,
  saving,
  deleting,
  togglingRecommendation,
  challenge,
  writeup,
  form,
  hasWriteup,
  visibilityLabel,
  loadPage,
  handleSave,
  handleDelete,
  handleToggleRecommendation,
  restoreExistingWriteup,
} = useChallengeWriteupEditorPage(props.challengeId)
</script>

<template>
  <component
    :is="isEmbedded ? 'div' : 'section'"
    :class="pageShellClass"
  >
    <header
      v-if="!isEmbedded"
      class="workspace-topbar"
    >
      <div class="topbar-leading">
        <span class="workspace-overline">Challenge Workspace</span>
        <span class="class-chip">题解管理</span>
      </div>
      <div class="writeup-top-actions">
        <button
          class="ui-btn ui-btn--ghost"
          type="button"
          @click="emit('back')"
        >
          返回题目
        </button>
        <button
          class="ui-btn ui-btn--ghost"
          type="button"
          @click="void loadPage()"
        >
          <RefreshCw class="h-4 w-4" />
          刷新
        </button>
      </div>
    </header>

    <PageHeader
      v-if="!isEmbedded"
      class="writeup-page-header"
      eyebrow="Admin Writeup"
      title="题解管理"
      :description="
        challenge
          ? `为《${challenge.title}》维护管理员题解，控制公开范围。`
          : '为题目维护管理员题解，控制公开范围。'
      "
    />
    <div
      v-else
      class="list-heading writeup-tab-heading"
    >
      <div>
        <div class="workspace-overline">Admin Writeup</div>
        <h1 class="workspace-page-title">题解管理</h1>
      </div>
      <p class="workspace-page-copy">
        {{
          challenge
            ? `为《${challenge.title}》维护管理员题解，控制公开范围。`
            : '为题目维护管理员题解，控制公开范围。'
        }}
      </p>
    </div>

    <div class="journal-divider" />

    <AppLoading
      v-if="loading"
      class="writeup-loading"
    >
      正在加载题解数据...
    </AppLoading>

    <main
      v-else
      :class="isEmbedded ? 'writeup-workspace' : 'content-pane writeup-workspace'"
    >
      <section class="writeup-main">
        <ChallengeWriteupEditorFormSection
          :form="form"
          :writeup="writeup"
          :has-writeup="hasWriteup"
          :visibility-label="visibilityLabel"
          :saving="saving"
          :deleting="deleting"
          :toggling-recommendation="togglingRecommendation"
          :handle-save="handleSave"
          :handle-delete="handleDelete"
          :handle-toggle-recommendation="handleToggleRecommendation"
          :restore-existing-writeup="restoreExistingWriteup"
        />

        <ChallengeWriteupSnapshotSection :writeup="writeup" />
      </section>

      <aside class="context-rail writeup-rail">
        <ChallengeWriteupChallengeRail :challenge="challenge" />
      </aside>
    </main>
  </component>
</template>
