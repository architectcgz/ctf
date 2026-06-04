<template>
  <header class="workspace-topbar">
    <div class="topbar-leading">
      <span class="workspace-overline">Challenge Workspace</span>
      <span class="class-chip">题解管理</span>
    </div>
    <div class="writeup-top-actions">
      <button class="ui-btn ui-btn--ghost" type="button" @click="emit('back')">
        返回题目
      </button>
      <button class="ui-btn ui-btn--ghost" type="button" @click="void loadPage()">
        <RefreshCw class="h-4 w-4" />
        刷新
      </button>
    </div>
  </header>

  <PageHeader
    class="writeup-page-header"
    eyebrow="Admin Writeup"
    title="题解管理"
    :description="
      challenge
        ? `为《${challenge.title}》维护管理员题解，控制公开范围。`
        : '为题目维护管理员题解，控制公开范围。'
    "
  />

  <div class="journal-divider" />

  <AppLoading v-if="loading" class="writeup-loading">
    正在加载题解数据...
  </AppLoading>

  <main v-else class="content-pane writeup-workspace">
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
</template>

<script setup lang="ts">
import { RefreshCw } from 'lucide-vue-next'

import AppLoading from '@/shared/ui/common/AppLoading.vue'
import PageHeader from '@/shared/ui/common/PageHeader.vue'
import { useChallengeWriteupEditorPage } from '../model'

import ChallengeWriteupChallengeRail from './ChallengeWriteupChallengeRail.vue'
import ChallengeWriteupEditorFormSection from './ChallengeWriteupEditorFormSection.vue'
import ChallengeWriteupSnapshotSection from './ChallengeWriteupSnapshotSection.vue'

const props = defineProps<{
  challengeId: string
}>()

const emit = defineEmits<{
  back: []
}>()

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
