<template>
  <section
    id="contest-panel-create"
    class="contest-panel contest-panel--create"
    :aria-hidden="active ? 'false' : 'true'"
  >
    <section class="workspace-directory-section contest-create-panel">
      <header class="list-heading contest-create-head">
        <div class="workspace-tab-heading__main">
          <div class="workspace-overline">
            Contest Setup
          </div>
          <h2 class="workspace-page-title">
            创建竞赛
          </h2>
          <p class="workspace-page-copy">
            在当前工作区里补齐竞赛基础信息和时间窗口，保存后直接回到赛事工作台继续编排。
          </p>
        </div>

        <div class="ui-toolbar-actions contest-panel-actions">
          <button
            type="button"
            class="ui-btn ui-btn--ghost"
            @click="emit('back')"
          >
            返回工作台
          </button>
        </div>
      </header>

      <PlatformContestFormPanel
        :mode="'create'"
        :draft="draft"
        :saving="saving"
        :field-locks="fieldLocks"
        :show-cancel="true"
        :note="'创建后可继续在赛事工作台中筛选目录、编辑详情或进入具体 AWD 赛区。'"
        @cancel="emit('back')"
        @save="emit('save', $event)"
      />
    </section>
  </section>
</template>

<script setup lang="ts">
import type { ContestFieldLocks, ContestFormDraft } from '../model'
import PlatformContestFormPanel from './PlatformContestFormPanel.vue'

defineProps<{
  active: boolean
  draft: ContestFormDraft
  saving: boolean
  fieldLocks: ContestFieldLocks
}>()

const emit = defineEmits<{
  back: []
  save: [value: ContestFormDraft]
}>()
</script>
