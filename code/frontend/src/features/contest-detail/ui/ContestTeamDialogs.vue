<template>
  <CFocusedInputDialog
    :open="showCreateTeam"
    title="创建新队伍"
    description="为你的战队起一个响亮的代号。创建完成后，你可以生成邀请链接让其他队友加入。"
    width="35rem"
    aria-label="创建队伍"
    overlay-class="c-focused-input-shell--plain"
    :close-on-backdrop="false"
    @update:open="emit('update:showCreateTeam', $event)"
    @close="emit('closeCreateTeam')"
  >
    <template #icon>
      <UsersRound
        class="h-6 w-6"
        :stroke-width="2"
      />
    </template>

    <div class="contest-team-dialog-field">
      <label for="contest-create-team-name">队伍名称</label>
      <input
        id="contest-create-team-name"
        :value="teamName"
        type="text"
        placeholder="例如：HackerG1"
        @input="updateTeamName"
        @keyup.enter="emit('createTeam')"
      >
    </div>

    <template #footer="{ close }">
      <button
        type="button"
        data-c-modal-action="ghost"
        @click="close"
      >
        取消
      </button>
      <button
        type="button"
        data-c-modal-action="primary"
        :disabled="creatingTeam"
        @click="emit('createTeam')"
      >
        {{ creatingTeam ? '创建中...' : '确认创建' }}
      </button>
    </template>
  </CFocusedInputDialog>

  <CFocusedInputDialog
    :open="showJoinTeam"
    title="加入现有队伍"
    description="输入队伍 ID 后立即加入当前战队。加入成功后，你会同步看到队伍成员与竞赛工作区。"
    width="34rem"
    aria-label="加入队伍"
    overlay-class="c-focused-input-shell--plain"
    :close-on-backdrop="false"
    @update:open="emit('update:showJoinTeam', $event)"
    @close="emit('closeJoinTeam')"
  >
    <template #icon>
      <UsersRound
        class="h-6 w-6"
        :stroke-width="2"
      />
    </template>

    <div class="contest-team-dialog-field">
      <label for="contest-join-team-id">队伍 ID</label>
      <input
        id="contest-join-team-id"
        :value="teamIdInput"
        type="text"
        placeholder="输入队伍 ID"
        @input="updateTeamIdInput"
        @keyup.enter="emit('joinTeam')"
      >
    </div>

    <template #footer="{ close }">
      <button
        type="button"
        data-c-modal-action="ghost"
        @click="close"
      >
        取消
      </button>
      <button
        type="button"
        data-c-modal-action="primary"
        :disabled="joiningTeam"
        @click="emit('joinTeam')"
      >
        {{ joiningTeam ? '加入中...' : '确认加入' }}
      </button>
    </template>
  </CFocusedInputDialog>
</template>

<style scoped>
.contest-team-dialog-field {
  display: grid;
  gap: 0.5rem;
}
</style>

<script setup lang="ts">
import { UsersRound } from 'lucide-vue-next'

import CFocusedInputDialog from '@/shared/ui/common/modal-templates/CFocusedInputDialog.vue'

defineProps<{
  showCreateTeam: boolean
  showJoinTeam: boolean
  teamName: string
  teamIdInput: string
  creatingTeam: boolean
  joiningTeam: boolean
}>()

const emit = defineEmits<{
  'update:showCreateTeam': [value: boolean]
  'update:showJoinTeam': [value: boolean]
  'update:teamName': [value: string]
  'update:teamIdInput': [value: string]
  closeCreateTeam: []
  closeJoinTeam: []
  createTeam: []
  joinTeam: []
}>()

function updateTeamName(event: Event): void {
  emit('update:teamName', (event.target as HTMLInputElement).value)
}

function updateTeamIdInput(event: Event): void {
  emit('update:teamIdInput', (event.target as HTMLInputElement).value)
}
</script>
