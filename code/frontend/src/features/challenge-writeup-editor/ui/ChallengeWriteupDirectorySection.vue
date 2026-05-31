<script setup lang="ts">
import AppEmpty from '@/shared/ui/common/AppEmpty.vue'
import AppLoading from '@/shared/ui/common/AppLoading.vue'
import PagePaginationControls from '@/shared/ui/common/PagePaginationControls.vue'

import type { WriteupDirectoryRow } from '../model/useChallengeWriteupManagement'
import ChallengeWriteupDirectoryRow from './ChallengeWriteupDirectoryRow.vue'

defineProps<{
  submissionLoading: boolean
  hasAnyWriteups: boolean
  challengeTitle?: string
  officialWriteupCount: number
  submissionTotal: number
  directoryRows: WriteupDirectoryRow[]
  submissionPage: number
  submissionTotalPages: number
  deleting: boolean
  actionMenuOpen: boolean
  openWriteup: (mode: 'view' | 'edit') => void
  setActionMenuOpen: (nextOpen: boolean) => void
  handleDelete: () => void | Promise<void>
  changeSubmissionPage: (page: number) => void | Promise<void>
}>()
</script>

<template>
  <section class="writeup-manage-section">
    <header class="list-heading writeup-manage-section__head">
      <div class="writeup-manage-section__intro">
        <div class="workspace-overline">
          Writeup Directory
        </div>
        <h2 class="list-heading__title">
          题解目录
        </h2>
      </div>
      <div class="writeup-manage-section__meta">
        共 {{ officialWriteupCount + submissionTotal }} 篇题解
      </div>
    </header>

    <AppLoading
      v-if="submissionLoading"
      class="writeup-manage-loading"
    >
      正在加载题解投稿...
    </AppLoading>

    <template v-else>
      <AppEmpty
        v-if="!hasAnyWriteups"
        icon="FileText"
        title="当前还没有题解"
        :description="
          challengeTitle
            ? `《${challengeTitle}》暂时还没有官方题解或学员题解。`
            : '当前题目暂时还没有官方题解或学员题解。'
        "
      >
        <template #actions>
          <button
            class="ui-btn ui-btn--primary"
            type="button"
            @click="openWriteup('edit')"
          >
            编写题解
          </button>
        </template>
      </AppEmpty>

      <template v-else>
        <section class="writeup-directory">
          <div
            class="writeup-directory-head"
            aria-hidden="true"
          >
            <span>题解标题</span>
            <span>来源</span>
            <span>作者</span>
            <span>学号</span>
            <span>状态</span>
            <span>更新时间</span>
            <span class="writeup-directory-head__actions">操作</span>
          </div>

          <ChallengeWriteupDirectoryRow
            v-for="row in directoryRows"
            :key="row.key"
            :row="row"
            :deleting="deleting"
            :action-menu-open="actionMenuOpen"
            :open-writeup="openWriteup"
            :set-action-menu-open="setActionMenuOpen"
            :handle-delete="handleDelete"
          />
        </section>

        <PagePaginationControls
          :page="submissionPage"
          :total-pages="submissionTotalPages"
          :total="submissionTotal"
          :disabled="submissionLoading"
          :total-label="`共 ${submissionTotal} 篇题解`"
          :show-jump="true"
          @change-page="void changeSubmissionPage($event)"
        />
      </template>
    </template>
  </section>
</template>
