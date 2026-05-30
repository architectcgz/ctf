<script setup lang="ts">
import { Save, Trash2 } from 'lucide-vue-next'

import type { AdminChallengeWriteupData, WriteupVisibility } from '@/api/contracts'

interface ChallengeWriteupEditorFormState {
  title: string
  content: string
  visibility: WriteupVisibility
}

defineProps<{
  form: ChallengeWriteupEditorFormState
  writeup: AdminChallengeWriteupData | null
  hasWriteup: boolean
  visibilityLabel: string
  saving: boolean
  deleting: boolean
  togglingRecommendation: boolean
  handleSave: () => void | Promise<void>
  handleDelete: () => void | Promise<void>
  handleToggleRecommendation: () => void | Promise<void>
  restoreExistingWriteup: () => void
}>()
</script>

<template>
  <section class="writeup-section writeup-editor-section">
    <header class="writeup-editor-head">
      <div>
        <div class="journal-note-label">
          Writeup Editor
        </div>
        <h2 class="writeup-section-title">
          编辑器
        </h2>
      </div>
      <div class="writeup-badges">
        <span
          class="writeup-badge"
          :class="hasWriteup ? 'writeup-badge--ok' : 'writeup-badge--warn'"
        >
          {{ hasWriteup ? '已存在题解' : '尚未创建' }}
        </span>
        <span
          v-if="writeup?.is_recommended"
          class="writeup-badge writeup-badge--accent"
        >
          推荐题解
        </span>
      </div>
    </header>

    <div class="writeup-form-grid">
      <label class="writeup-field writeup-field--title">
        <span class="writeup-field-label">题解标题</span>
        <input
          v-model="form.title"
          type="text"
          class="writeup-field-input"
          placeholder="例如：官方解题思路 / 赛后复盘"
        >
      </label>

      <label class="writeup-field writeup-field--visibility">
        <span class="writeup-field-label">可见性</span>
        <select
          v-model="form.visibility"
          class="writeup-field-input"
        >
          <option value="private">private</option>
          <option value="public">public</option>
        </select>
      </label>
    </div>

    <div class="writeup-visibility-note">
      {{ visibilityLabel }}
    </div>

    <label class="writeup-field writeup-field--content">
      <span class="writeup-field-label">题解正文</span>
      <textarea
        v-model="form.content"
        rows="16"
        class="writeup-content-input"
        placeholder="输入官方题解、赛后复盘或教学讲解内容。"
      />
    </label>

    <div
      class="writeup-editor-actions"
      role="group"
      aria-label="题解编辑操作"
    >
      <button
        :disabled="saving"
        class="ui-btn ui-btn--primary"
        type="button"
        @click="void handleSave()"
      >
        <Save class="h-4 w-4" />
        {{ saving ? '保存中...' : '保存题解' }}
      </button>
      <button
        v-if="hasWriteup"
        :disabled="togglingRecommendation"
        class="ui-btn ui-btn--secondary"
        type="button"
        @click="void handleToggleRecommendation()"
      >
        {{
          togglingRecommendation
            ? '处理中...'
            : writeup?.is_recommended
              ? '取消推荐'
              : '设为推荐'
        }}
      </button>
      <button
        v-if="hasWriteup"
        class="ui-btn ui-btn--ghost"
        type="button"
        @click="restoreExistingWriteup"
      >
        恢复已保存版本
      </button>
      <button
        v-if="hasWriteup"
        :disabled="deleting"
        class="ui-btn ui-btn--danger"
        type="button"
        @click="void handleDelete()"
      >
        <Trash2 class="h-4 w-4" />
        {{ deleting ? '删除中...' : '删除题解' }}
      </button>
    </div>
  </section>
</template>
