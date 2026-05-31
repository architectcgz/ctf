<script setup lang="ts">
import AppEmpty from '@/shared/ui/common/AppEmpty.vue'
import type { AdminChallengeWriteupData } from '@/api/contracts'

defineProps<{
  writeup: AdminChallengeWriteupData | null
}>()
</script>

<template>
  <section class="writeup-section writeup-snapshot-section">
    <header class="writeup-subsection-head">
      <div class="journal-note-label">
        Snapshot
      </div>
      <h2 class="writeup-section-title">
        当前已保存版本
      </h2>
    </header>

    <template v-if="writeup">
      <dl class="writeup-snapshot-grid">
        <div class="writeup-snapshot-item">
          <dt>标题</dt>
          <dd>{{ writeup.title }}</dd>
        </div>
        <div class="writeup-snapshot-item">
          <dt>可见性</dt>
          <dd>{{ writeup.visibility }}</dd>
        </div>
        <div class="writeup-snapshot-item">
          <dt>推荐状态</dt>
          <dd>{{ writeup.is_recommended ? '推荐题解' : '未推荐' }}</dd>
        </div>
        <div class="writeup-snapshot-item">
          <dt>创建时间</dt>
          <dd>{{ writeup.created_at }}</dd>
        </div>
        <div class="writeup-snapshot-item">
          <dt>更新时间</dt>
          <dd>{{ writeup.updated_at }}</dd>
        </div>
      </dl>
    </template>

    <AppEmpty
      v-else
      title="当前还没有管理员题解"
      description="填写表单后点击保存，即可创建题解并控制公开范围。"
      icon="BookOpen"
    />
  </section>
</template>
