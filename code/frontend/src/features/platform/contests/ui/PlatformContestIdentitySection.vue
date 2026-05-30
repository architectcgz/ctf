<script setup lang="ts">
import { FileText } from 'lucide-vue-next'

import PlatformContestFormSectionShell from './PlatformContestFormSectionShell.vue'

defineProps<{
  titleValue: string
  descriptionValue: string
  titleError: string
}>()

const emit = defineEmits<{
  'update:title': [value: string]
  'update:description': [value: string]
}>()
</script>

<template>
  <PlatformContestFormSectionShell
    title="基础信息"
    description="定义竞赛在平台展示的基础信息与访问权限。"
    :icon="FileText"
  >
    <div class="ui-field contest-form-field contest-form-row">
      <label class="contest-form-row__label">竞赛标题</label>
      <div class="contest-form-row__control">
        <div
          class="ui-control-wrap contest-form-control-wrap"
          :class="{ 'is-error': !!titleError }"
        >
          <input
            id="contest-title"
            :value="titleValue"
            type="text"
            class="ui-control contest-form-input"
            placeholder="输入竞赛标题..."
            @input="emit('update:title', ($event.target as HTMLInputElement).value)"
          >
        </div>
        <p
          v-if="titleError"
          class="contest-form-field-error"
        >
          {{ titleError }}
        </p>
        <p class="contest-form-field-hint">
          请控制在 40 个字以内，建议包含年份与赛季信息。
        </p>
      </div>
    </div>

    <div class="ui-field contest-form-field contest-form-row">
      <label class="contest-form-row__label">竞赛描述</label>
      <div class="contest-form-row__control">
        <div class="ui-control-wrap contest-form-control-wrap">
          <textarea
            id="contest-description"
            :value="descriptionValue"
            rows="4"
            class="ui-control contest-form-textarea"
            placeholder="描述竞赛的背景、赛制及对参赛者的要求..."
            @input="emit('update:description', ($event.target as HTMLTextAreaElement).value)"
          />
        </div>
        <p class="contest-form-field-hint">
          支持 Markdown 语法，将展示在竞赛详情页。
        </p>
      </div>
    </div>
  </PlatformContestFormSectionShell>
</template>
