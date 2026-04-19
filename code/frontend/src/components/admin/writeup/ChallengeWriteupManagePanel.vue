<script setup lang="ts">
import { MoreHorizontal } from 'lucide-vue-next'
import { computed, onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'

import { deleteChallengeWriteup, getChallengeWriteup } from '@/api/admin'
import { getTeacherWriteupSubmissions } from '@/api/teacher'
import type { AdminChallengeWriteupData, TeacherSubmissionWriteupItemData } from '@/api/contracts'
import AdminPaginationControls from '@/components/admin/AdminPaginationControls.vue'
import AppEmpty from '@/components/common/AppEmpty.vue'
import AppLoading from '@/components/common/AppLoading.vue'
import { confirmDestructiveAction } from '@/composables/useDestructiveConfirm'
import { useToast } from '@/composables/useToast'

type WriteupDirectoryRow = {
  key: string
  source: 'official' | 'student'
  title: string
  preview?: string
  authorPrimary: string
  authorSecondary?: string
  authorTertiary?: string
  studentNo: string
  statusPrimary: string
  statusSecondary?: string
  updatedAt: string
}

const props = defineProps<{
  challengeId: string
  challengeTitle?: string
}>()

const router = useRouter()
const toast = useToast()

const loading = ref(true)
const deleting = ref(false)
const submissionLoading = ref(true)
const actionMenuOpen = ref(false)
const writeup = ref<AdminChallengeWriteupData | null>(null)
const writeupSubmissions = ref<TeacherSubmissionWriteupItemData[]>([])
const submissionPage = ref(1)
const submissionPageSize = ref(6)
const submissionTotal = ref(0)

const submissionTotalPages = computed(() =>
  Math.max(1, Math.ceil(submissionTotal.value / Math.max(1, submissionPageSize.value)))
)
const officialWriteupCount = computed(() => (writeup.value ? 1 : 0))
const hasAnyWriteups = computed(() => Boolean(writeup.value) || writeupSubmissions.value.length > 0)
const directoryRows = computed<WriteupDirectoryRow[]>(() => {
  const rows: WriteupDirectoryRow[] = []

  if (writeup.value && submissionPage.value === 1) {
    rows.push({
      key: `official-${writeup.value.id}`,
      source: 'official',
      title: writeup.value.title,
      authorPrimary: '平台官方',
      authorSecondary: '独立查看 / 编辑入口',
      studentNo: '-',
      statusPrimary: writeup.value.visibility,
      statusSecondary: writeup.value.is_recommended ? '推荐题解' : '未推荐',
      updatedAt: formatDate(writeup.value.updated_at),
    })
  }

  rows.push(
    ...writeupSubmissions.value.map((item) => ({
      key: `student-${item.id}`,
      source: 'student' as const,
      title: item.title,
      preview: item.content_preview,
      authorPrimary: resolveAuthorName(item),
      authorSecondary: `@${item.student_username}`,
      authorTertiary: resolveClassName(item),
      studentNo: resolveStudentNo(item),
      statusPrimary: submissionStatusLabel(item.submission_status),
      statusSecondary: visibilityStatusLabel(item.visibility_status),
      updatedAt: formatDate(item.updated_at),
    }))
  )

  return rows
})

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

async function loadWriteupSubmissions(targetPage = 1) {
  if (!props.challengeId) {
    writeupSubmissions.value = []
    submissionPage.value = 1
    submissionTotal.value = 0
    submissionLoading.value = false
    return
  }

  submissionLoading.value = true
  try {
    const payload = await getTeacherWriteupSubmissions({
      challenge_id: props.challengeId,
      page: targetPage,
      page_size: submissionPageSize.value,
    })
    writeupSubmissions.value = payload.list
    submissionPage.value = payload.page
    submissionPageSize.value = payload.page_size
    submissionTotal.value = payload.total
  } catch {
    toast.error('加载题解投稿失败')
  } finally {
    submissionLoading.value = false
  }
}

function openWriteup(mode: 'view' | 'edit') {
  if (!props.challengeId) return
  actionMenuOpen.value = false
  void router.push({
    path:
      mode === 'view'
        ? `/platform/challenges/${props.challengeId}/writeup/view`
        : `/platform/challenges/${props.challengeId}/writeup`,
  })
}

function openActionMenu() {
  actionMenuOpen.value = true
}

function closeActionMenu() {
  actionMenuOpen.value = false
}

function toggleActionMenu() {
  actionMenuOpen.value = !actionMenuOpen.value
}

function handleActionMenuFocusout(event: FocusEvent) {
  const currentTarget = event.currentTarget
  if (!(currentTarget instanceof HTMLElement)) {
    closeActionMenu()
    return
  }

  const nextTarget = event.relatedTarget
  if (nextTarget instanceof Node && currentTarget.contains(nextTarget)) {
    return
  }

  closeActionMenu()
}

async function handleDelete() {
  if (!props.challengeId || !writeup.value || deleting.value) {
    return
  }

  const confirmed = await confirmDestructiveAction({
    message: '确定删除当前题解吗？删除后学员将无法继续查看。',
  })
  if (!confirmed) {
    return
  }

  deleting.value = true
  try {
    await deleteChallengeWriteup(props.challengeId)
    writeup.value = null
    closeActionMenu()
    toast.success('题解已删除')
  } catch (error) {
    const message = error instanceof Error && error.message.trim() ? error.message : '删除题解失败'
    toast.error(message)
  } finally {
    deleting.value = false
  }
}

async function changeSubmissionPage(page: number) {
  if (page < 1 || page === submissionPage.value || submissionLoading.value || !props.challengeId) {
    return
  }

  await loadWriteupSubmissions(page)
}

function formatDate(value: string): string {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return value
  }
  return date.toLocaleString('zh-CN')
}

function submissionStatusLabel(
  status: TeacherSubmissionWriteupItemData['submission_status']
): string {
  return status === 'draft' ? '草稿' : '已发布'
}

function visibilityStatusLabel(
  status: TeacherSubmissionWriteupItemData['visibility_status']
): string {
  return status === 'hidden' ? '已隐藏' : '已公开'
}

function resolveAuthorName(item: TeacherSubmissionWriteupItemData): string {
  const name = item.student_name?.trim()
  return name || item.student_username
}

function resolveStudentNo(item: TeacherSubmissionWriteupItemData): string {
  const studentNo = item.student_no?.trim()
  return studentNo || '未设置学号'
}

function resolveClassName(item: TeacherSubmissionWriteupItemData): string {
  const className = item.class_name?.trim()
  return className || '未分班'
}

watch(
  () => props.challengeId,
  () => {
    void loadWriteup()
    void loadWriteupSubmissions(1)
  }
)

onMounted(() => {
  void loadWriteup()
  void loadWriteupSubmissions(1)
})
</script>

<template>
  <section class="writeup-manage-panel">
    <div class="writeup-manage-header">
      <div class="list-heading writeup-manage-heading">
        <div>
          <div class="workspace-overline">Writeup Directory</div>
          <h1 class="workspace-page-title">题解管理</h1>
        </div>

        <div class="writeup-manage-actions">
          <button class="ui-btn ui-btn--primary" type="button" @click="openWriteup('edit')">
            编写题解
          </button>
        </div>
      </div>
    </div>

    <div class="writeup-manage-stats-shell">
      <div
        class="admin-summary-grid writeup-summary-grid progress-strip metric-panel-grid metric-panel-default-surface"
      >
        <article class="journal-note progress-card metric-panel-card">
          <div class="journal-note-label progress-card-label metric-panel-label">官方题解</div>
          <div class="journal-note-value progress-card-value metric-panel-value">
            {{ officialWriteupCount }}
          </div>
          <div class="journal-note-helper progress-card-hint metric-panel-helper">
            当前题目已创建的官方题解数量
          </div>
        </article>
        <article class="journal-note progress-card metric-panel-card">
          <div class="journal-note-label progress-card-label metric-panel-label">学员题解</div>
          <div class="journal-note-value progress-card-value metric-panel-value">
            {{ submissionTotal }}
          </div>
          <div class="journal-note-helper progress-card-hint metric-panel-helper">
            当前题目收到的学员题解投稿数量
          </div>
        </article>
      </div>
    </div>

    <AppLoading v-if="loading && submissionLoading" class="writeup-manage-loading">
      正在加载题解内容...
    </AppLoading>

    <template v-else>
      <section class="writeup-manage-section">
        <header class="list-heading writeup-manage-section__head">
          <div class="writeup-manage-section__intro">
            <div class="journal-note-label">Writeup Directory</div>
            <h2 class="list-heading__title">题解目录</h2>
          </div>
          <div class="writeup-manage-section__meta">
            共 {{ officialWriteupCount + submissionTotal }} 篇题解
          </div>
        </header>

        <AppLoading v-if="submissionLoading" class="writeup-manage-loading"
          >正在加载题解投稿...</AppLoading
        >

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
              <button class="ui-btn ui-btn--primary" type="button" @click="openWriteup('edit')">
                编写题解
              </button>
            </template>
          </AppEmpty>

          <template v-else>
            <section class="writeup-directory">
              <div class="writeup-directory-head" aria-hidden="true">
                <span>题解标题</span>
                <span>来源</span>
                <span>作者</span>
                <span>学号</span>
                <span>状态</span>
                <span>更新时间</span>
                <span class="writeup-directory-head__actions">操作</span>
              </div>

              <article v-for="row in directoryRows" :key="row.key" class="writeup-row">
                <div class="writeup-row__title">
                  <div class="writeup-row__name">{{ row.title }}</div>
                  <div v-if="row.preview" class="writeup-row__preview">{{ row.preview }}</div>
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
                  <div class="writeup-row__author-name">{{ row.authorPrimary }}</div>
                  <div v-if="row.authorSecondary" class="writeup-row__author-meta">
                    {{ row.authorSecondary }}
                  </div>
                  <div v-if="row.authorTertiary" class="writeup-row__author-meta">
                    {{ row.authorTertiary }}
                  </div>
                </div>
                <div class="writeup-row__student-no">{{ row.studentNo }}</div>
                <div class="writeup-row__status">
                  <div>{{ row.statusPrimary }}</div>
                  <div v-if="row.statusSecondary" class="writeup-row__status-subtle">
                    {{ row.statusSecondary }}
                  </div>
                </div>
                <div class="writeup-row__updated">{{ row.updatedAt }}</div>
                <div class="writeup-row__actions" role="group" aria-label="题解目录操作">
                  <template v-if="row.source === 'official'">
                    <button
                      class="ui-btn ui-btn--secondary ui-btn--sm"
                      type="button"
                      @click="openWriteup('view')"
                    >
                      查看
                    </button>
                    <div
                      class="writeup-actions-menu-shell"
                      @mouseenter="openActionMenu"
                      @mouseleave="closeActionMenu"
                      @focusin="openActionMenu"
                      @focusout="handleActionMenuFocusout"
                      @keydown.esc="closeActionMenu"
                    >
                      <button
                        class="ui-btn ui-btn--ghost ui-btn--sm writeup-actions-menu-trigger"
                        data-testid="writeup-more-actions"
                        type="button"
                        aria-label="更多题解操作"
                        :aria-expanded="actionMenuOpen ? 'true' : 'false'"
                        @mouseenter="openActionMenu"
                        @focus="openActionMenu"
                        @click="toggleActionMenu"
                      >
                        <MoreHorizontal class="h-4 w-4" />
                      </button>

                      <div
                        v-if="actionMenuOpen"
                        class="writeup-actions-menu"
                        role="menu"
                        aria-label="更多题解操作"
                      >
                        <button
                          class="ui-btn ui-btn--ghost ui-btn--sm writeup-actions-menu__button"
                          role="menuitem"
                          type="button"
                          @click="openWriteup('edit')"
                        >
                          编辑
                        </button>
                        <button
                          :disabled="deleting"
                          class="ui-btn ui-btn--danger ui-btn--sm writeup-actions-menu__button"
                          role="menuitem"
                          type="button"
                          @click="void handleDelete()"
                        >
                          {{ deleting ? '删除中...' : '删除' }}
                        </button>
                      </div>
                    </div>
                  </template>
                  <span v-else class="writeup-row__placeholder">--</span>
                </div>
              </article>
            </section>

            <AdminPaginationControls
              :page="submissionPage"
              :total-pages="submissionTotalPages"
              :total="submissionTotal"
              :disabled="submissionLoading"
              :total-label="`共 ${submissionTotal} 篇题解`"
              @change-page="void changeSubmissionPage($event)"
            />
          </template>
        </template>
      </section>
    </template>
  </section>
</template>

<style scoped>
.writeup-manage-panel {
  display: grid;
  gap: var(--space-6);
}

.writeup-manage-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-4);
}

.list-heading {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-end;
  justify-content: space-between;
  gap: var(--space-3);
}

.list-heading__title {
  margin: var(--space-1) 0 0;
  font-size: var(--font-size-1-20);
  font-weight: 700;
  color: var(--journal-ink);
}

.workspace-overline {
  font-size: var(--font-size-0-70);
  font-weight: 700;
  letter-spacing: 0.2em;
  text-transform: uppercase;
  color: var(--journal-accent);
}

.writeup-manage-heading {
  width: 100%;
}

.writeup-manage-section {
  display: grid;
  gap: var(--space-4);
}

.writeup-manage-section + .writeup-manage-section {
  padding-top: var(--space-5);
  border-top: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
}

.writeup-manage-section__head {
  margin: 0;
}

.writeup-manage-section__intro {
  display: grid;
  gap: var(--space-4);
  min-width: 0;
}

.writeup-manage-section__meta {
  font-size: var(--font-size-0-82);
  color: var(--journal-muted);
}

.writeup-summary-grid {
  --admin-summary-grid-columns: repeat(2, minmax(0, 12rem));
  --admin-summary-grid-gap: var(--space-3);
}

.writeup-manage-actions {
  display: flex;
  justify-content: flex-end;
}

.writeup-manage-stats-shell {
  display: grid;
}

.writeup-manage-loading {
  padding-block: var(--space-7);
}

.writeup-directory {
  --writeup-directory-columns: minmax(14rem, 1.55fr) minmax(6rem, 0.58fr) minmax(12rem, 1.1fr)
    minmax(8.5rem, 0.72fr) minmax(8rem, 0.78fr) minmax(10.5rem, 0.9fr) minmax(10rem, 10rem);
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
.writeup-row__source,
.writeup-row__author,
.writeup-row__student-no,
.writeup-row__status,
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

.writeup-row__preview,
.writeup-row__author-meta,
.writeup-row__status-subtle {
  margin-top: 0.18rem;
  font-size: var(--font-size-0-78);
  color: var(--journal-muted);
}

.writeup-row__preview {
  display: -webkit-box;
  overflow: hidden;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  white-space: normal;
}

.writeup-row__source-pill {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 1.85rem;
  padding: 0 var(--space-3);
  border-radius: 999px;
  border: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
  background: color-mix(in srgb, var(--journal-surface) 96%, var(--color-bg-base));
  font-size: var(--font-size-0-74);
  font-weight: 700;
  letter-spacing: 0.08em;
  color: var(--journal-ink);
}

.writeup-row__source-pill--official {
  border-color: color-mix(in srgb, var(--journal-accent) 26%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 14%, var(--journal-surface));
  color: color-mix(in srgb, var(--journal-accent) 82%, white);
}

.writeup-row__source-pill--student {
  border-color: color-mix(in srgb, var(--color-primary) 24%, transparent);
  background: color-mix(in srgb, var(--color-primary) 12%, var(--journal-surface));
  color: color-mix(in srgb, var(--color-primary) 78%, white);
}

.writeup-row__author-name {
  font-size: var(--font-size-0-92);
  font-weight: 600;
  color: var(--journal-ink);
}

.writeup-row__student-no,
.writeup-row__status,
.writeup-row__updated {
  font-size: var(--font-size-0-86);
  color: var(--journal-ink);
}

.writeup-row__actions {
  display: flex;
  gap: var(--space-2);
  justify-content: flex-end;
  position: relative;
  justify-self: end;
}

.writeup-row__placeholder {
  display: inline-flex;
  align-items: center;
  min-height: 2.1rem;
  color: var(--journal-muted);
}

.writeup-actions-menu-shell {
  position: relative;
}

.writeup-actions-menu-trigger {
  min-width: 2.5rem;
  padding-inline: var(--space-2-5);
}

.writeup-actions-menu {
  position: absolute;
  top: calc(100% + var(--space-1-5));
  right: 0;
  z-index: 10;
  display: grid;
  gap: var(--space-2);
  min-width: 9rem;
  padding: var(--space-2-5);
  border: 1px solid color-mix(in srgb, var(--journal-border) 92%, transparent);
  border-radius: 0.9rem;
  background: linear-gradient(
    180deg,
    color-mix(in srgb, var(--journal-surface) 98%, var(--color-bg-base)),
    color-mix(in srgb, var(--journal-surface-subtle) 96%, var(--color-bg-base))
  );
  box-shadow: 0 16px 32px var(--color-shadow-soft);
}

.writeup-actions-menu__button {
  justify-content: flex-start;
  width: 100%;
}

.writeup-manage-actions > .ui-btn,
.writeup-row__actions > .ui-btn,
.writeup-actions-menu__button.ui-btn {
  --ui-btn-secondary-background: color-mix(
    in srgb,
    var(--journal-surface) 96%,
    var(--color-bg-base)
  );
  --ui-btn-secondary-border: color-mix(in srgb, var(--journal-border) 92%, transparent);
  --ui-btn-secondary-hover-border: color-mix(in srgb, var(--journal-accent) 42%, transparent);
  --ui-btn-secondary-hover-background: color-mix(
    in srgb,
    var(--journal-accent) 4%,
    var(--journal-surface)
  );
  --ui-btn-secondary-hover-shadow: 0 8px 18px color-mix(in srgb, var(--color-shadow-soft) 72%, transparent);
  --ui-btn-secondary-color: var(--journal-ink);
  --ui-btn-secondary-hover-color: var(--journal-accent);
  --ui-btn-ghost-color: var(--journal-ink);
  --ui-btn-ghost-hover-color: var(--journal-accent);
  --ui-btn-ghost-hover-background: color-mix(in srgb, var(--journal-accent) 4%, var(--journal-surface));
  --ui-btn-danger-border: color-mix(in srgb, var(--color-danger) 20%, transparent);
  --ui-btn-danger-background: color-mix(in srgb, var(--color-danger) 10%, var(--journal-surface));
  --ui-btn-danger-color: color-mix(in srgb, var(--color-danger) 88%, var(--journal-ink));
  --ui-btn-danger-hover-border: color-mix(in srgb, var(--color-danger) 26%, transparent);
  --ui-btn-danger-hover-background: color-mix(in srgb, var(--color-danger) 14%, var(--journal-surface));
}

.writeup-manage-actions > .ui-btn,
.writeup-manage-section :deep(.app-empty__actions .ui-btn) {
  --ui-btn-primary-background: var(--journal-accent);
  --ui-btn-primary-hover-background: color-mix(in srgb, var(--journal-accent) 88%, black);
  --ui-btn-primary-border: color-mix(in srgb, var(--journal-accent) 18%, transparent);
}

@media (max-width: 960px) {
  .writeup-manage-header {
    flex-direction: column;
  }

  .list-heading {
    align-items: flex-start;
    flex-direction: column;
  }

  .writeup-manage-section__head {
    width: 100%;
  }

  .writeup-summary-grid {
    --admin-summary-grid-columns: 1fr;
  }

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

  .writeup-actions-menu {
    right: auto;
    left: 0;
  }
}
</style>
