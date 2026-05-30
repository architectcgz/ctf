<script setup lang="ts">
import { MoreHorizontal } from 'lucide-vue-next'

import CActionMenu from '@/components/common/menus/CActionMenu.vue'
import type { WriteupDirectoryRow } from '../model/useChallengeWriteupManagement'

defineProps<{
  row: WriteupDirectoryRow
  deleting: boolean
  actionMenuOpen: boolean
  openWriteup: (mode: 'view' | 'edit') => void
  setActionMenuOpen: (nextOpen: boolean) => void
  handleDelete: () => void | Promise<void>
}>()
</script>

<template>
  <article class="writeup-row">
    <div class="writeup-row__title">
      <div class="writeup-row__name">
        {{ row.title }}
      </div>
      <div
        v-if="row.preview"
        class="writeup-row__preview"
      >
        {{ row.preview }}
      </div>
    </div>
    <div class="writeup-row__source">
      <span
        class="writeup-row__source-pill"
        :class="`writeup-row__source-pill--${row.source}`"
      >
        {{ row.source === 'official' ? '官方' : '学员' }}
      </span>
    </div>
    <div class="writeup-row__author">
      <div class="writeup-row__author-name">
        {{ row.authorPrimary }}
      </div>
      <div
        v-if="row.authorSecondary"
        class="writeup-row__author-meta"
      >
        {{ row.authorSecondary }}
      </div>
      <div
        v-if="row.authorTertiary"
        class="writeup-row__author-meta"
      >
        {{ row.authorTertiary }}
      </div>
    </div>
    <div class="writeup-row__student-no">
      {{ row.studentNo }}
    </div>
    <div class="writeup-row__status">
      <div>{{ row.statusPrimary }}</div>
      <div
        v-if="row.statusSecondary"
        class="writeup-row__status-subtle"
      >
        {{ row.statusSecondary }}
      </div>
    </div>
    <div class="writeup-row__updated">
      {{ row.updatedAt }}
    </div>
    <div
      class="writeup-row__actions"
      role="group"
      aria-label="题解目录操作"
    >
      <template v-if="row.source === 'official'">
        <button
          class="ui-btn ui-btn--secondary ui-btn--sm"
          type="button"
          @click="openWriteup('view')"
        >
          查看
        </button>
        <CActionMenu
          :open="actionMenuOpen"
          title="Management"
          menu-label="更多题解操作"
          @update:open="setActionMenuOpen"
        >
          <template #trigger="{ open, toggle, setTriggerRef }">
            <button
              :ref="setTriggerRef"
              class="c-action-menu__trigger c-action-menu__trigger--icon"
              data-testid="writeup-more-actions"
              type="button"
              aria-label="更多题解操作"
              aria-haspopup="menu"
              :aria-expanded="open ? 'true' : 'false'"
              @click="toggle"
            >
              <MoreHorizontal class="h-4 w-4" />
            </button>
          </template>

          <template #default>
            <button
              class="c-action-menu__item"
              role="menuitem"
              type="button"
              @click="openWriteup('edit')"
            >
              编辑
            </button>
            <button
              :disabled="deleting"
              class="c-action-menu__item c-action-menu__item--danger"
              role="menuitem"
              type="button"
              @click="void handleDelete()"
            >
              {{ deleting ? '删除中...' : '删除' }}
            </button>
          </template>
        </CActionMenu>
      </template>
      <span
        v-else
        class="writeup-row__placeholder"
      >--</span>
    </div>
  </article>
</template>
