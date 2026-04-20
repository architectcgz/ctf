<script setup lang="ts">
import TeacherAwdWorkspaceLayout from '@/modules/awd/layouts/TeacherAwdWorkspaceLayout.vue'

import { useTeacherAwdWorkspacePage } from './useTeacherAwdWorkspacePage'

const { pageModel, layoutProps, exportArchive, exportReport, exporting, polling } =
  useTeacherAwdWorkspacePage('export')
</script>

<template>
  <TeacherAwdWorkspaceLayout v-bind="layoutProps">
    <div class="awd-teacher-page">
      <section class="awd-teacher-section">
        <div class="awd-teacher-section__head">
          <div>
            <div class="workspace-overline">Export Actions</div>
            <h2>导出操作</h2>
          </div>
          <div class="awd-teacher-section__hint">
            {{ polling ? '导出轮询中' : pageModel.export.snapshotLabel }}
          </div>
        </div>

        <div class="awd-teacher-actions">
          <button
            type="button"
            class="ui-btn ui-btn--ghost"
            :disabled="exporting === 'archive'"
            @click="exportArchive"
          >
            {{ exporting === 'archive' ? '生成中...' : '导出复盘归档' }}
          </button>
          <button
            type="button"
            class="ui-btn ui-btn--primary"
            :disabled="exporting === 'report' || !pageModel.export.canExportReport"
            @click="exportReport"
          >
            {{ exporting === 'report' ? '生成中...' : '导出教师报告' }}
          </button>
        </div>
      </section>

      <section class="awd-teacher-section">
        <div class="awd-teacher-section__head">
          <div>
            <div class="workspace-overline">Export Cards</div>
            <h2>导出面板</h2>
          </div>
        </div>

        <div class="awd-teacher-card-grid">
          <article
            v-for="card in pageModel.export.cards"
            :key="card.title"
            class="awd-teacher-card"
          >
            <span>{{ card.title }}</span>
            <strong>{{ card.stateLabel }}</strong>
            <small>{{ card.description }}</small>
          </article>
        </div>
      </section>
    </div>
  </TeacherAwdWorkspaceLayout>
</template>

<style scoped src="./teacher-awd-view.css"></style>
