<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'

import { getChallengeWriteup } from '@/api/admin'
import type { AdminChallengeWriteupData } from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'
import AppLoading from '@/components/common/AppLoading.vue'
import { useToast } from '@/composables/useToast'

const props = defineProps<{
  challengeId: string
  challengeTitle?: string
}>()

const router = useRouter()
const toast = useToast()

const loading = ref(true)
const writeup = ref<AdminChallengeWriteupData | null>(null)

async function loadWriteup() {
  if (!props.challengeId) {
    writeup.value = null
    loading.value = false
    return
  }

  loading.value = true
  try {
    writeup.value = await getChallengeWriteup(props.challengeId)
  } catch {
    toast.error('加载题解目录失败')
  } finally {
    loading.value = false
  }
}

function openWriteupEditor() {
  if (!props.challengeId) return
  void router.push(`/platform/challenges/${props.challengeId}/writeup`)
}

watch(
  () => props.challengeId,
  () => {
    void loadWriteup()
  }
)

onMounted(() => {
  void loadWriteup()
})
</script>

<template>
  <section class="writeup-manage-panel">
    <div class="workspace-tab-heading">
      <div class="workspace-tab-heading__main">
        <div class="journal-note-label">Writeup Directory</div>
        <h1 class="workspace-tab-heading__title">题解管理</h1>
      </div>
      <p class="workspace-tab-copy">
        {{ challengeTitle ? `查看《${challengeTitle}》的题解目录，并进入独立编辑页维护内容。` : '查看当前题目的题解目录，并进入独立编辑页维护内容。' }}
      </p>
    </div>

    <div class="writeup-manage-actions">
      <button class="admin-btn admin-btn-primary" type="button" @click="openWriteupEditor">
        编写题解
      </button>
    </div>

    <AppLoading v-if="loading" class="writeup-manage-loading">正在加载题解目录...</AppLoading>

    <template v-else>
      <section v-if="writeup" class="writeup-directory">
        <div class="writeup-directory-head" aria-hidden="true">
          <span>题解标题</span>
          <span>可见性</span>
          <span>推荐状态</span>
          <span>更新时间</span>
          <span class="writeup-directory-head__actions">操作</span>
        </div>

        <article class="writeup-row">
          <div class="writeup-row__title">
            <div class="writeup-row__name">{{ writeup.title }}</div>
          </div>
          <div class="writeup-row__visibility">{{ writeup.visibility }}</div>
          <div class="writeup-row__recommendation">
            {{ writeup.is_recommended ? '推荐题解' : '未推荐' }}
          </div>
          <div class="writeup-row__updated">{{ writeup.updated_at }}</div>
          <div class="writeup-row__actions" role="group" aria-label="题解目录操作">
            <button class="admin-btn admin-btn-primary admin-btn-compact" type="button" @click="openWriteupEditor">
              查看 / 编辑
            </button>
          </div>
        </article>
      </section>

      <AppEmpty
        v-else
        icon="BookOpen"
        title="当前还没有题解"
        :description="challengeTitle ? `为《${challengeTitle}》创建第一份官方题解。` : '为当前题目创建第一份官方题解。'"
      >
        <template #actions>
          <button class="admin-btn admin-btn-primary" type="button" @click="openWriteupEditor">
            编写题解
          </button>
        </template>
      </AppEmpty>
    </template>
  </section>
</template>

<style scoped>
.writeup-manage-panel {
  display: grid;
  gap: var(--space-5);
}

.writeup-manage-actions {
  display: flex;
  justify-content: flex-end;
}

.writeup-manage-loading {
  padding-block: var(--space-7);
}

.writeup-directory {
  --writeup-directory-columns: minmax(14rem, 1.6fr) minmax(7rem, 0.7fr) minmax(7rem, 0.8fr)
    minmax(11rem, 1fr) minmax(8.5rem, 8.5rem);
  display: grid;
  gap: 0;
}

.writeup-directory-head,
.writeup-row {
  display: grid;
  grid-template-columns: var(--writeup-directory-columns);
  gap: var(--space-4);
}

.writeup-directory-head {
  padding: 0 0 var(--space-3);
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
  font-size: var(--font-size-0-72);
  font-weight: 700;
  letter-spacing: 0.16em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.writeup-directory-head__actions {
  text-align: right;
}

.writeup-row {
  align-items: center;
  padding: var(--space-4) 0;
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
}

.writeup-row__title,
.writeup-row__visibility,
.writeup-row__recommendation,
.writeup-row__updated,
.writeup-row__actions {
  min-width: 0;
}

.writeup-row__name {
  font-size: var(--font-size-0-92);
  font-weight: 600;
  color: var(--journal-ink);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.writeup-row__visibility,
.writeup-row__recommendation,
.writeup-row__updated {
  font-size: var(--font-size-0-86);
  color: var(--journal-ink);
}

.writeup-row__actions {
  display: flex;
  justify-content: flex-end;
}

.admin-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-2);
  min-height: 2.25rem;
  border-radius: 0.65rem;
  border: 1px solid transparent;
  padding: var(--space-2) var(--space-3-5);
  font-size: var(--font-size-0-84);
  font-weight: 600;
  transition:
    border-color 150ms ease,
    background 150ms ease,
    color 150ms ease,
    box-shadow 150ms ease;
}

.admin-btn:focus-visible {
  outline: none;
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--journal-accent) 18%, transparent);
}

.admin-btn-primary {
  border-color: color-mix(in srgb, var(--journal-accent) 18%, transparent);
  background: var(--journal-accent);
  color: #fff;
}

.admin-btn-compact {
  min-height: 2.1rem;
}

@media (max-width: 960px) {
  .writeup-directory-head {
    display: none;
  }

  .writeup-row {
    grid-template-columns: minmax(0, 1fr);
    gap: var(--space-2);
    align-items: start;
  }

  .writeup-row__actions {
    justify-content: flex-start;
    margin-top: var(--space-2);
  }
}
</style>
