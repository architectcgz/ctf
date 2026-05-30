<script setup lang="ts">
defineProps<{
  openWriteup: (mode: 'view' | 'edit') => void
}>()
</script>

<template>
  <div class="writeup-manage-header">
    <div class="list-heading writeup-manage-heading">
      <div>
        <div class="workspace-overline">
          Writeup Directory
        </div>
        <h1 class="workspace-page-title">题解管理</h1>
      </div>

      <div class="writeup-manage-actions">
        <button
          class="ui-btn ui-btn--primary"
          type="button"
          @click="openWriteup('edit')"
        >
          编写题解
        </button>
      </div>
    </div>
  </div>
</template>
