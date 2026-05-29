<script setup lang="ts">
import { Edit } from 'lucide-vue-next'

import AppRouteLink from '@/components/navigation/AppRouteLink.vue'
import { ChallengeCategoryPill, toChallengeCategory } from '@/entities/challenge'

import type { ContestAwdConfigRouteTarget } from '../model'
import type { AWDChallengeConfigDirectoryItemView } from './awdChallengeConfigPanel.types'

defineProps<{
  item: AWDChallengeConfigDirectoryItemView
  editRoute: ContestAwdConfigRouteTarget | null
}>()
</script>

<template>
  <tr class="studio-row">
    <td class="col-identity">
      <div class="challenge-identity">
        <RouterLink
          :id="`awd-challenge-preview-${item.challengeId}`"
          class="challenge-title challenge-title-link"
          :to="item.previewRoute"
          :title="`打开题目预览：${item.title}`"
        >
          {{ item.title }}
        </RouterLink>
        <div class="challenge-subtitle">
          <ChallengeCategoryPill
            v-if="toChallengeCategory(item.category)"
            :category="toChallengeCategory(item.category)!"
          />
          <span v-else>{{ item.category || '通用' }}</span>
          <span>RANK {{ item.order }}</span>
        </div>
      </div>
    </td>
    <td class="col-meta">
      <div class="engine-tag">
        {{ item.checkerTypeLabel }}
      </div>
    </td>
    <td class="col-meta">
      <div class="score-stack">
        <span class="score-main">SLA {{ item.slaScore }}</span>
        <span class="score-sub">Defense {{ item.defenseScore }}</span>
      </div>
    </td>
    <td class="col-meta">
      <div
        class="rules-summary"
        :title="item.configSummary"
      >
        {{ item.configSummary }}
      </div>
    </td>
    <td class="col-status">
      <div class="validation-block">
        <span
          class="validation-pill"
          :class="item.validationState"
        >
          {{ item.validationStateText }}
        </span>
        <span class="validation-time">{{ item.validationPrimaryHint }}</span>
      </div>
    </td>
    <td class="col-actions">
      <div class="ui-row-actions config-row__actions">
        <RouterLink
          class="ui-btn ui-btn--secondary"
          :to="item.previewRoute"
        >
          预览
        </RouterLink>
        <AppRouteLink
          v-if="editRoute"
          :id="`awd-challenge-config-edit-${item.challengeId}`"
          class="ui-btn ui-btn--primary"
          :to="editRoute"
        >
          <Edit class="h-3.5 w-3.5" />
          编辑
        </AppRouteLink>
      </div>
    </td>
  </tr>
</template>
