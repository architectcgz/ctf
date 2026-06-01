<script setup lang="ts">
import AppEmpty from '@/shared/ui/common/AppEmpty.vue'

import type { ContestAwdConfigRouteTarget } from '../model'
import AWDChallengeConfigDirectoryRow from './AWDChallengeConfigDirectoryRow.vue'
import type { AWDChallengeConfigDirectoryItemView } from './awdChallengeConfigPanel.types'

const props = defineProps<{
  items: AWDChallengeConfigDirectoryItemView[]
  buildEditRoute?: (
    item: AWDChallengeConfigDirectoryItemView['source']
  ) => ContestAwdConfigRouteTarget | null
}>()
</script>

<template>
  <section class="workspace-directory-section awd-config-directory">
    <header class="list-heading">
      <div>
        <div class="journal-note-label">
          Challenge Directory
        </div>
        <h3 class="list-heading__title">
          题目目录
        </h3>
      </div>
    </header>

    <AppEmpty
      v-if="items.length === 0"
      title="暂无关联服务"
      description="请先在题目编排中关联题目。"
      icon="Layers"
      class="py-20"
    />

    <div
      v-else
      class="studio-table-wrap"
    >
      <table class="studio-table">
        <thead>
          <tr>
            <th class="col-identity">
              服务身份
            </th>
            <th class="col-meta">
              裁判引擎
            </th>
            <th class="col-meta">
              SLA / 防守权重
            </th>
            <th class="col-meta">
              规则摘要
            </th>
            <th class="col-status">
              就绪验证
            </th>
            <th class="col-actions">
              操作
            </th>
          </tr>
        </thead>
        <tbody>
          <AWDChallengeConfigDirectoryRow
            v-for="item in items"
            :key="item.source.id"
            :item="item"
            :edit-route="props.buildEditRoute?.(item.source) ?? null"
          />
        </tbody>
      </table>
    </div>
  </section>
</template>
